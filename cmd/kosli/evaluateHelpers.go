package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

// policyFetchTimeout caps how long a remote --policy fetch can take.
var policyFetchTimeout = 10 * time.Second

// policyMaxBytes caps how much of a remote --policy response we read into
// memory. Real Rego policies are kilobytes; this guards against a malicious or
// misconfigured server streaming an unbounded body. 5 * 2^20 (5*1MiB)
const policyMaxBytes = 5 << 20 // 5 MiB

// maxServerSideTrails mirrors the ceiling the evaluations API publishes in its
// OpenAPI schema. Checked here so that a caller naming too many is told which
// limit they crossed, rather than reading it out of a rejected request. Drift
// makes this refuse what the API would accept, so the two are worth comparing
// whenever that schema changes.
const maxServerSideTrails = 100

// maxPolicyBundleFiles mirrors the ceiling the evaluations API publishes on
// the entries in a policy bundle.
const maxPolicyBundleFiles = 100

// serverPolicyMaxBytes mirrors the cap the evaluations API publishes on a
// policy bundle, which counts the names as well as the sources. It is a fifth
// of what a remote --policy read allows, so a policy can be fetched in full
// and still be too big to send; saying so here beats a rejected request.
const serverPolicyMaxBytes = 1 << 20 // 1 MiB

// serverSideWaitOptions is how long a verdict is waited for, and the seam a
// test shrinks it through. It lives here rather than being an override of the
// package default, so a test never reaches into another package's state to set
// it and can never leave a shortened budget behind for whatever runs next.
var serverSideWaitOptions = evaluations.WaitOptions{}

type commonEvaluateOptions struct {
	flowName     string
	policyRef    string
	output       string
	showInput    bool
	attestations []string
	params       string
	assert       bool
	noAssert     bool
	serverSide   bool
	outputRules  []string
}

func (o *commonEvaluateOptions) addFlags(cmd *cobra.Command, policyDesc string) {
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.policyRef, "policy", "p", "", policyDesc)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")
	cmd.Flags().StringSliceVar(&o.attestations, "attestations", nil, "[optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.")
	cmd.Flags().StringVar(&o.params, "params", "", policyParamsFlag)
	cmd.Flags().BoolVar(&o.assert, "assert", false, "[optional] Exit with a non-zero status when the policy denies. This is the current default; pass --assert to lock it in across future releases.")
	cmd.Flags().BoolVar(&o.noAssert, "no-assert", false, "[optional] Print the result and always exit 0, even when the policy denies. Use when this command feeds another tool as a policy decision point.")
	cmd.Flags().StringSliceVar(&o.outputRules, "output-rule", nil, policyOutputRuleFlag)
	cmd.MarkFlagsMutuallyExclusive("assert", "no-assert")
}

// addServerSideFlag offers the evaluation to the Kosli server instead of
// running it here. It is hidden, and stays hidden: it exists to run the two
// evaluation paths against each other while neither is a contract anyone can
// rely on, and the two do not yet agree on what a policy may contain or on
// what a policy sees. Only the trail commands have it, because an evaluation
// is created from trail references and `evaluate input` has none to send.
func (o *commonEvaluateOptions) addServerSideFlag(cmd *cobra.Command) {
	cmd.Flags().BoolVar(&o.serverSide, "server-side", false, serverSideFlag)
	if err := cmd.Flags().MarkHidden("server-side"); err != nil {
		logger.Error("failed to hide the server-side flag: %v", err)
	}
}

// assertOnDeny resolves the --assert / --no-assert pair into a single bool.
// Today the default is true (assert); a future major release flips this by
// returning o.assert directly.
func (o *commonEvaluateOptions) assertOnDeny() bool {
	return !o.noAssert
}

func fetchAndEnrichTrail(flowName, trailName string, attestations []string) (any, error) {
	trailURL, err := url.JoinPath(global.Host, "api/v2/trails", global.Org, flowName, trailName)
	if err != nil {
		return nil, err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    trailURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return nil, err
	}

	var trailData any
	err = json.Unmarshal([]byte(response.Body), &trailData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trail response: %v", err)
	}

	trailData = evaluate.TransformTrail(trailData)
	trailData = evaluate.FilterAttestations(trailData, attestations)

	ids := evaluate.CollectAttestationIDs(trailData)
	if len(ids) > 0 {
		details := make(map[string]any)
		for _, id := range ids {
			detailURL, err := url.JoinPath(global.Host, "api/v2/attestations", global.Org)
			if err != nil {
				return nil, err
			}
			q := url.Values{}
			q.Set("attestation_id", id)
			detailURL += "?" + q.Encode()
			detailResp, err := kosliClient.Do(&requests.RequestParams{
				Method: http.MethodGet,
				URL:    detailURL,
				Token:  global.ApiToken,
			})
			if err != nil {
				return nil, fmt.Errorf("failed to fetch attestation detail for %s: %w", id, err)
			}
			var wrapper map[string]any
			if err := json.Unmarshal([]byte(detailResp.Body), &wrapper); err != nil {
				return nil, fmt.Errorf("failed to parse attestation detail for %s: %w", id, err)
			}
			if data, ok := wrapper["data"].([]any); ok && len(data) > 0 {
				if entry, ok := data[0].(map[string]any); ok {
					details[id] = entry
				}
			}
		}
		trailData = evaluate.RehydrateTrail(trailData, details)
	}

	return trailData, nil
}

// loadPolicy reads a Rego policy from a local file path or, when ref starts
// with http:// or https://, fetches it over HTTP. Remote fetches are
// unauthenticated and uncached; callers are responsible for the integrity of
// the source.
func loadPolicy(ref string) ([]byte, error) {
	if isRemotePolicyRef(ref) {
		return fetchRemotePolicy(ref)
	}
	body, err := os.ReadFile(ref)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}
	return body, nil
}

func isRemotePolicyRef(ref string) bool {
	return strings.HasPrefix(ref, "http://") || strings.HasPrefix(ref, "https://")
}

func fetchRemotePolicy(remoteURL string) ([]byte, error) {
	if strings.HasPrefix(remoteURL, "http://") {
		logger.Warn("fetching policy over plain HTTP from %s; prefer https://", remoteURL)
	}

	client := &http.Client{
		Timeout:       policyFetchTimeout,
		CheckRedirect: sameHostRedirectPolicy,
	}

	if global != nil && global.HttpProxy != "" {
		proxyURL, err := url.Parse(global.HttpProxy)
		if err != nil {
			return nil, fmt.Errorf("failed to parse --http-proxy %q: %w", global.HttpProxy, err)
		}
		// This builds a fresh http.Transport with Go defaults. If the
		// codebase later adopts a shared transport with custom TLS/dial
		// settings, route policy fetches through it (or clone it) so they
		// inherit those settings.
		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	}

	resp, err := client.Get(remoteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch policy from %s: %w", remoteURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("failed to fetch policy from %s: HTTP %d", remoteURL, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, policyMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read policy response from %s: %w", remoteURL, err)
	}
	if int64(len(body)) > policyMaxBytes {
		return nil, fmt.Errorf("policy at %s exceeds %d-byte limit", remoteURL, policyMaxBytes)
	}

	return body, nil
}

// sameHostRedirectPolicy allows redirects only when the target host matches
// the most recent request's host. This blocks an SSRF vector where a trusted
// remote redirects the CLI to an internal address.
func sameHostRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) >= 5 {
		return fmt.Errorf("stopped after %d redirects", len(via))
	}
	if req.URL.Host != via[len(via)-1].URL.Host {
		return fmt.Errorf("cross-host redirect to %s blocked", req.URL.Host)
	}
	return nil
}

func parseParams(raw string) (map[string]any, error) {
	if raw == "" {
		return nil, nil
	}

	var jsonBytes []byte
	if strings.HasPrefix(raw, "@") {
		var err error
		jsonBytes, err = os.ReadFile(raw[1:])
		if err != nil {
			return nil, fmt.Errorf("failed to read --params file: %w", err)
		}
	} else {
		jsonBytes = []byte(raw)
	}

	var params map[string]any
	if err := json.Unmarshal(jsonBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to parse --params: %w", err)
	}
	return params, nil
}

func evaluateAndPrintResult(out io.Writer, policyRef string, input map[string]any, outputFormat string, showInput bool, params map[string]any, assertOnDeny bool, outputRules []string) error {
	policySource, err := loadPolicy(policyRef)
	if err != nil {
		return err
	}

	result, err := evaluate.Evaluate(string(policySource), input, params, outputRules...)
	if err != nil {
		return err
	}

	return printEvaluateResult(out, result, input, outputFormat, showInput, params, assertOnDeny, "")
}

// evaluateServerSide asks the Kosli server to evaluate the named trails and
// prints the verdict it answers with. Nothing about the trails is read here:
// the server assembles what the policy sees, which is the point of the flag.
func evaluateServerSide(out io.Writer, o *commonEvaluateOptions, trails []evaluations.TrailRef) error {
	if err := o.refuseWhatTheServerCannotDo(); err != nil {
		return err
	}

	return runServerEvaluation(out, serverEvaluation{
		policyRef:    o.policyRef,
		params:       o.params,
		trails:       trails,
		output:       o.output,
		assertOnDeny: o.assertOnDeny(),
	})
}

type serverEvaluation struct {
	policyRef    string
	params       string
	trails       []evaluations.TrailRef
	decision     *evaluations.Decision
	output       string
	assertOnDeny bool
}

// runServerEvaluation is shared by every command that evaluates away from this
// machine, so they cannot drift on what they send or on how an outcome reads.
func runServerEvaluation(out io.Writer, spec serverEvaluation) error {
	if len(spec.trails) > maxServerSideTrails {
		return fmt.Errorf("a server-side evaluation takes at most %d trails, got %d",
			maxServerSideTrails, len(spec.trails))
	}

	// Parsed before the policy is read: --params is local and cheap to check,
	// and a remote policy fetched first would be thrown away by a typo in it.
	params, err := parseParams(spec.params)
	if err != nil {
		return err
	}

	files, err := policyBundle(spec.policyRef)
	if err != nil {
		return err
	}

	client := evaluations.NewClient(kosliClient, global.Host, global.ApiToken, global.DryRun)
	created, err := client.Create(global.Org, evaluations.CreateRequest{
		Trails:   spec.trails,
		Files:    files,
		Params:   params,
		Decision: spec.decision,
	})
	if err != nil {
		return serverSideRequestError(err)
	}
	if created == nil {
		// A dry run sent nothing, so there is no evaluation to wait for.
		return nil
	}
	if created.ID == "" {
		// Without an id there is nothing to read the verdict back from, and
		// asking anyway would fetch a different resource and blame the answer.
		return errors.New("the Kosli server accepted the evaluation but named no id, " +
			"so its verdict cannot be read back")
	}

	evaluation, err := client.WaitForTerminal(context.Background(), global.Org, created.ID,
		serverSideWaitOptions)
	if err != nil {
		return serverSideReadError(created.ID, err)
	}
	// Status decides, not the presence of a result: an evaluation that reports
	// a failure has decided nothing, whatever else it carries. Reading it the
	// other way round would let a stray result print as a verdict, which is the
	// one outcome none of this may produce.
	if evaluation.Status != evaluations.StatusCompleted || evaluation.Result == nil {
		return serverSideFailure(evaluation)
	}

	return printEvaluateResult(out, serverVerdict(evaluation.Result), nil,
		spec.output, false, nil, spec.assertOnDeny, evaluation.DecisionAttestationID)
}

// refuseWhatTheServerCannotDo rejects the options that have no server-side
// answer, rather than accepting them and quietly doing something else.
// Honouring either one only halfway would be worse than refusing it: a filter
// that was ignored would evaluate more than the caller asked about, and an
// input printed from here would not be the input the server judged.
//
// Stated as errors rather than as a cobra exclusion group so that each one can
// say why, which is what an insider reaching for an undocumented flag needs.
func (o *commonEvaluateOptions) refuseWhatTheServerCannotDo() error {
	if len(o.attestations) > 0 {
		return fmt.Errorf(
			"--attestations is not supported with --server-side; " +
				"filtering is done here and the server has no equivalent")
	}
	if o.showInput {
		return fmt.Errorf(
			"--show-input is not supported with --server-side; " +
				"the server does not return the input it evaluated")
	}
	if len(o.outputRules) > 0 {
		return fmt.Errorf(
			"--output-rule is not supported with --server-side; " +
				"the server only returns allow and violations")
	}
	return nil
}

// serverSideRequestError reports a refusal from the API as the status it
// answered with and whatever it said about why.
func serverSideRequestError(err error) error {
	// An expired wait is not a failed request: it already says what happened
	// and names the evaluation, so it travels untouched.
	var stillPending *evaluations.StillPendingError
	if errors.As(err, &stillPending) {
		return err
	}

	var apiError *requests.APIError
	if !errors.As(err, &apiError) {
		// Retried and given up on inside the shared HTTP client, which throws
		// the body away as it goes, so there is no status or message to report.
		return fmt.Errorf("could not get a server-side evaluation from Kosli: %w", err)
	}
	// Without a message the API wrote there is only the status to report.
	// Quoting what arrived instead would dress a proxy's page, or the decoder's
	// complaint about one, as the server's own reason.
	if !apiError.HasServerMessage || apiError.Message == "" {
		return fmt.Errorf("the Kosli server answered %d", apiError.StatusCode)
	}
	return fmt.Errorf("the Kosli server answered %d: %s", apiError.StatusCode, apiError.Message)
}

// serverSideReadError reports a failure to read a verdict back. By this point
// the evaluation exists and the server holds its answer, so the id travels: it
// is the one thing that makes the outcome recoverable.
//
// Separate from the create's own mapping because every sentence there is
// written for a request that has not happened yet. Reused here, a refusal would
// blame a feature flag that has already let a create through, and a 404 would
// deny support for a route the create just used.
func serverSideReadError(id string, err error) error {
	// An expired wait already names the evaluation and says where to read it.
	var stillPending *evaluations.StillPendingError
	if errors.As(err, &stillPending) {
		return err
	}
	return fmt.Errorf("could not read server-side evaluation %s back: %w", id, err)
}

// serverSideFailure reports an evaluation that answered no verdict. It is
// never a denial: a policy that could not run has decided nothing.
func serverSideFailure(evaluation *evaluations.Evaluation) error {
	if evaluation.Failure == nil {
		return fmt.Errorf("server-side evaluation %s answered no verdict and no reason", evaluation.ID)
	}
	return fmt.Errorf("server-side evaluation failed (%s): %s",
		evaluation.Failure.Kind, evaluation.Failure.Message)
}

// serverVerdict maps a server verdict onto the shared one. An empty list of
// violations becomes no list at all, because the local evaluator returns
// nothing rather than an empty slice and the two paths have to print alike.
func serverVerdict(result *evaluations.Result) *evaluate.Result {
	violations := result.Violations
	if len(violations) == 0 {
		violations = nil
	}
	return &evaluate.Result{Allow: result.Allow, Violations: violations}
}

// policyBundle reads what --policy names as the bundle the API takes. The caps
// are checked here so an oversized bundle is named as such rather than rejected
// as an opaque 422, and the byte cap counts the names as well as the sources,
// exactly as the API counts.
func policyBundle(ref string) (map[string]string, error) {
	files, err := policyBundleFiles(ref)
	if err != nil {
		return nil, err
	}

	if len(files) > maxPolicyBundleFiles {
		return nil, fmt.Errorf("policy bundle holds %d files, over the limit of %d",
			len(files), maxPolicyBundleFiles)
	}
	size := 0
	for name, source := range files {
		size += len(name) + len(source)
	}
	if size > serverPolicyMaxBytes {
		return nil, fmt.Errorf("policy bundle is %d bytes, over the %d byte limit",
			size, serverPolicyMaxBytes)
	}
	return files, nil
}

func policyBundleFiles(ref string) (map[string]string, error) {
	if !isRemotePolicyRef(ref) {
		if info, err := os.Stat(ref); err == nil && info.IsDir() {
			return policyDirectory(ref)
		}
	}

	source, err := loadPolicy(ref)
	if err != nil {
		return nil, err
	}
	return map[string]string{policyBundleKey(ref): string(source)}, nil
}

// policyDirectory collects every file below root, keyed by its path relative
// to it. Nothing is left out by name: a rule here would refuse bundles the
// evaluator that runs them would have accepted.
func policyDirectory(root string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read policy file: %w", err)
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		// Relative paths reach the API spelled one way, whatever this machine
		// spells them with.
		files[filepath.ToSlash(name)] = string(source)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		// The API takes at least one file, so an empty directory is named here
		// rather than sent to be refused.
		return nil, fmt.Errorf("no file found under %s", root)
	}
	return files, nil
}

// policyBundleKey names the policy inside the uploaded bundle. Only the base
// name travels: the server refuses a path that is absolute or that climbs out
// of the bundle, and where the file sits on this machine is not its business.
// The name is a label rather than a selector, since the evaluator parses every
// entry as a module whatever it is called, so no extension is imposed here.
func policyBundleKey(ref string) string {
	base := filepath.Base(ref)
	if isRemotePolicyRef(ref) {
		if parsed, err := url.Parse(ref); err == nil {
			base = path.Base(parsed.Path)
		}
	}
	switch base {
	case ".", "..", "/", "":
		return "policy.rego"
	}
	return base
}

// printEvaluateResult renders a verdict, whatever produced it, so that every
// evaluation path prints the same bytes for the same verdict.
var evaluateResultKeys = []string{"allow", "violations", "input", "params", "decision_attestation_id"}

func validateOutputRules(rules []string) error {
	for _, rule := range rules {
		if slices.Contains(evaluateResultKeys, rule) {
			return fmt.Errorf("--output-rule cannot be '%s', it is already part of the output", rule)
		}
	}
	return nil
}

func printEvaluateResult(out io.Writer, result *evaluate.Result, input map[string]any, outputFormat string, showInput bool, params map[string]any, assertOnDeny bool, decisionID string) error {
	auditResult := map[string]any{
		"allow":      result.Allow,
		"violations": result.Violations,
	}
	for rule, value := range result.Outputs {
		auditResult[rule] = value
	}
	if len(result.Outputs) > 0 && outputFormat == "table" {
		logger.Warn("--output-rule values are only shown with --output json")
	}
	// Absent everywhere else, so a caller reading a verdict alone parses the
	// same page as before.
	if decisionID != "" {
		auditResult["decision_attestation_id"] = decisionID
	}
	if showInput {
		auditResult["input"] = input
	}
	if showInput && params != nil {
		auditResult["params"] = params
	}

	raw, err := json.Marshal(auditResult)
	if err != nil {
		return fmt.Errorf("failed to marshal output: %v", err)
	}

	return output.FormattedPrint(string(raw), outputFormat, out, 0,
		map[string]output.FormatOutputFunc{
			"json":  printEvaluateResultAsJsonFn(assertOnDeny),
			"table": printEvaluateResultAsTableFn(assertOnDeny),
		})
}

func printEvaluateResultAsJsonFn(assertOnDeny bool) output.FormatOutputFunc {
	return func(raw string, out io.Writer, _ int) error {
		if err := output.PrintJson(raw, out, 0); err != nil {
			return err
		}

		var result map[string]any
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return err
		}
		if allow, ok := result["allow"].(bool); ok && !allow && assertOnDeny {
			return fmt.Errorf("policy denied")
		}
		return nil
	}
}

func printEvaluateResultAsTableFn(assertOnDeny bool) output.FormatOutputFunc {
	return func(raw string, out io.Writer, _ int) error {
		var result map[string]any
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return err
		}

		allow, _ := result["allow"].(bool)
		decisionRow := []string{}
		if id, ok := result["decision_attestation_id"].(string); ok && id != "" {
			decisionRow = append(decisionRow, fmt.Sprintf("DECISION:\t%s", id))
		}

		var rows []string
		if allow {
			rows = append(rows, "RESULT:\tALLOWED")
			tabFormattedPrint(out, []string{}, append(rows, decisionRow...))
			return nil
		}

		rows = append(rows, "RESULT:\tDENIED")

		if violations, ok := result["violations"].([]any); ok && len(violations) > 0 {
			for i, v := range violations {
				if i == 0 {
					rows = append(rows, fmt.Sprintf("VIOLATIONS:\t%s", v))
				} else {
					rows = append(rows, fmt.Sprintf("\t%s", v))
				}
			}
			tabFormattedPrint(out, []string{}, append(rows, decisionRow...))
			if assertOnDeny {
				return fmt.Errorf("policy denied: %v", violations)
			}
			return nil
		}
		tabFormattedPrint(out, []string{}, append(rows, decisionRow...))
		if assertOnDeny {
			return fmt.Errorf("policy denied")
		}
		return nil
	}
}
