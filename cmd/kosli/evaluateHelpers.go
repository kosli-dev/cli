package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
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
}

func (o *commonEvaluateOptions) addFlags(cmd *cobra.Command, policyDesc string) {
	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.policyRef, "policy", "p", "", policyDesc)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, "[optional] Include the policy input data in the output.")
	cmd.Flags().StringSliceVar(&o.attestations, "attestations", nil, "[optional] Limit which attestations are included. Plain name for trail-level, dot-qualified (artifact.name) for artifact-level.")
	cmd.Flags().StringVar(&o.params, "params", "", "[optional] Policy parameters as inline JSON or @file.json. Available in policies as data.params.")
	cmd.Flags().BoolVar(&o.assert, "assert", false, "[optional] Exit with a non-zero status when the policy denies. This is the current default; pass --assert to lock it in across future releases.")
	cmd.Flags().BoolVar(&o.noAssert, "no-assert", false, "[optional] Print the result and always exit 0, even when the policy denies. Use when this command feeds another tool as a policy decision point.")
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

func fetchAndEnrichTrail(flowName, trailName string, attestations []string) (interface{}, error) {
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

	var trailData interface{}
	err = json.Unmarshal([]byte(response.Body), &trailData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trail response: %v", err)
	}

	trailData = evaluate.TransformTrail(trailData)
	trailData = evaluate.FilterAttestations(trailData, attestations)

	ids := evaluate.CollectAttestationIDs(trailData)
	if len(ids) > 0 {
		details := make(map[string]interface{})
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
			var wrapper map[string]interface{}
			if err := json.Unmarshal([]byte(detailResp.Body), &wrapper); err != nil {
				return nil, fmt.Errorf("failed to parse attestation detail for %s: %w", id, err)
			}
			if data, ok := wrapper["data"].([]interface{}); ok && len(data) > 0 {
				if entry, ok := data[0].(map[string]interface{}); ok {
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

func parseParams(raw string) (map[string]interface{}, error) {
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

	var params map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &params); err != nil {
		return nil, fmt.Errorf("failed to parse --params: %w", err)
	}
	return params, nil
}

func evaluateAndPrintResult(out io.Writer, policyRef string, input map[string]interface{}, outputFormat string, showInput bool, params map[string]interface{}, assertOnDeny bool) error {
	policySource, err := loadPolicy(policyRef)
	if err != nil {
		return err
	}

	result, err := evaluate.Evaluate(string(policySource), input, params)
	if err != nil {
		return err
	}

	return printEvaluateResult(out, result, input, outputFormat, showInput, params, assertOnDeny)
}

// evaluateServerSide asks the Kosli server to evaluate the named trails and
// prints the verdict it answers with. Nothing about the trails is read here:
// the server assembles what the policy sees, which is the point of the flag.
func evaluateServerSide(out io.Writer, o *commonEvaluateOptions, trails []evaluations.TrailRef) error {
	if err := o.refuseWhatTheServerCannotDo(); err != nil {
		return err
	}

	if len(trails) > maxServerSideTrails {
		return fmt.Errorf("a server-side evaluation takes at most %d trails, got %d",
			maxServerSideTrails, len(trails))
	}

	// Parsed before the policy is read: --params is local and cheap to check,
	// and a remote policy fetched first would be thrown away by a typo in it.
	params, err := parseParams(o.params)
	if err != nil {
		return err
	}

	policySource, err := loadPolicy(o.policyRef)
	if err != nil {
		return err
	}

	files, err := policyBundle(o.policyRef, policySource)
	if err != nil {
		return err
	}

	client := evaluations.NewClient(kosliClient, global.Host, global.ApiToken, global.DryRun)
	created, err := client.Create(global.Org, evaluations.CreateRequest{
		Trails: trails,
		Files:  files,
		Params: params,
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
		o.output, false, nil, o.assertOnDeny())
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
	return nil
}

// serverSideRequestError turns a refusal from the API into something a caller
// can act on. Where the server explained itself, its own words are passed on
// unchanged; the two cases below are the ones where they are missing or not
// enough on their own.
func serverSideRequestError(err error) error {
	// An expired wait already says exactly what happened and names the
	// evaluation. Dressing it as a transport failure would put a sentence
	// about not reaching Kosli in front of one saying Kosli was reached.
	var stillPending *evaluations.StillPendingError
	if errors.As(err, &stillPending) {
		return err
	}

	var apiError *requests.APIError
	if !errors.As(err, &apiError) {
		// A retryable status or network trouble, which the shared client has
		// already retried and given up on, discarding the body as it went. The
		// server's own sentence is gone, so this supplies one.
		return fmt.Errorf("could not get a server-side evaluation from Kosli: %w", err)
	}

	switch {
	case apiError.StatusCode == http.StatusForbidden && apiError.HasServerMessage && apiError.Message != "":
		// The server's own words travel too: a refusal can come from a token
		// without rights on the org rather than from the feature flag, and
		// naming only the flag would send the reader after the wrong thing.
		return fmt.Errorf("server-side evaluation was refused for org '%s': %s. "+
			"It is gated on the is-server-side-evaluation-enabled feature flag; "+
			"remove --server-side to evaluate on this machine instead",
			global.Org, apiError.Message)

	case apiError.StatusCode == http.StatusForbidden:
		// A refusal from something in front of Kosli, such as a proxy, which
		// has no sentence of the API's to pass on. Quoting what it did send
		// would dress a decoder complaint as the server's reason.
		return fmt.Errorf("server-side evaluation was refused for org '%s'. "+
			"It is gated on the is-server-side-evaluation-enabled feature flag; "+
			"remove --server-side to evaluate on this machine instead", global.Org)

	// Not a sentence the API wrote, so this 404 came from something that does
	// not serve the route at all: a server too old to have it, or a proxy in
	// front of one. The shared client records the difference where the body is
	// decoded, since nothing downstream could tell afterwards.
	case apiError.StatusCode == http.StatusNotFound && !apiError.HasServerMessage:
		// Deliberately not certain which: an unmatched route and a route that
		// has since moved look alike from here, and a sentence that has to be
		// right about the difference is one a support thread quotes back.
		return errors.New("the evaluation request was answered with a 404: either this " +
			"Kosli server does not support server-side evaluation, or the route has " +
			"moved; remove --server-side to evaluate on this machine instead")
	}

	// An API error prints as its message and nothing else, so one that arrived
	// without a message prints as nothing at all. The status is the only thing
	// left to say, and saying it beats a bare "Error:".
	if apiError.Message == "" {
		return fmt.Errorf("the Kosli server answered %d and said nothing about why", apiError.StatusCode)
	}
	return err
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

// policyBundle wraps the policy source as the one-file bundle the API takes,
// refusing one too large for it rather than letting the request be rejected.
// The cap counts the names as well as the sources, exactly as the API counts.
func policyBundle(ref string, source []byte) (map[string]string, error) {
	key := policyBundleKey(ref)
	if size := len(key) + len(source); size > serverPolicyMaxBytes {
		return nil, fmt.Errorf("policy bundle is %d bytes, over the %d byte limit",
			size, serverPolicyMaxBytes)
	}
	return map[string]string{key: string(source)}, nil
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
func printEvaluateResult(out io.Writer, result *evaluate.Result, input map[string]interface{}, outputFormat string, showInput bool, params map[string]interface{}, assertOnDeny bool) error {
	auditResult := map[string]interface{}{
		"allow":      result.Allow,
		"violations": result.Violations,
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

		var result map[string]interface{}
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
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(raw), &result); err != nil {
			return err
		}

		allow, _ := result["allow"].(bool)

		var rows []string
		if allow {
			rows = append(rows, "RESULT:\tALLOWED")
			tabFormattedPrint(out, []string{}, rows)
			return nil
		}

		rows = append(rows, "RESULT:\tDENIED")

		if violations, ok := result["violations"].([]interface{}); ok && len(violations) > 0 {
			for i, v := range violations {
				if i == 0 {
					rows = append(rows, fmt.Sprintf("VIOLATIONS:\t%s", v))
				} else {
					rows = append(rows, fmt.Sprintf("\t%s", v))
				}
			}
			tabFormattedPrint(out, []string{}, rows)
			if assertOnDeny {
				return fmt.Errorf("policy denied: %v", violations)
			}
			return nil
		}
		tabFormattedPrint(out, []string{}, rows)
		if assertOnDeny {
			return fmt.Errorf("policy denied")
		}
		return nil
	}
}
