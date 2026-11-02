package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getEvaluationShortDesc = `Get a server-side evaluation, or what it was given.`

const getEvaluationLongDesc = getEvaluationShortDesc + `
The evaluation is shown with its status, the policy it ran, the trails it was
asked about, its outcome when it completed, the error when it failed, and the
decision it recorded, if any. The command exits 0 whatever the outcome.

With ^--show-input^ and ^--output json^, the document the policy read as ^input^ is
added under ^.input^: each trail as it stood at the evaluation's request time.

With ^--fixture^ and ^--dir^, what the evaluation was given is written to the
directory, so it can be replayed with stock OPA:
  - ^input.json^: the document the policy read as ^input^.
  - ^data.json^: the params, which the policy reads as ^data.params^.
  - ^policy/^: the files of the policy that ran, at their paths in the bundle.
The directory must be empty or not exist yet. An evaluation that is still
pending can be written too, as its input does not depend on its outcome.`

const getEvaluationExample = `
# get an evaluation:
kosli get evaluation yourEvaluationID \
	--api-token yourAPIToken \
	--org yourOrgName

# get an evaluation with the input its policy read:
kosli get evaluation yourEvaluationID \
	--show-input \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName

# write what an evaluation was given to ./replay, and replay it with OPA:
kosli get evaluation yourEvaluationID \
	--fixture \
	--dir ./replay \
	--api-token yourAPIToken \
	--org yourOrgName
opa eval --input replay/input.json --data replay/data.json \
	--bundle replay/policy/ 'data.policy.allow'
`

type getEvaluationOptions struct {
	output    string
	showInput bool
	fixture   bool
	dir       string
}

func newGetEvaluationCmd(out io.Writer) *cobra.Command {
	o := new(getEvaluationOptions)
	cmd := &cobra.Command{
		Use:         "evaluation EVALUATION-ID",
		Short:       getEvaluationShortDesc,
		Long:        getEvaluationLongDesc,
		Example:     getEvaluationExample,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if err := o.validate(); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if o.fixture {
				return o.runFixture(out, args[0])
			}
			return o.run(out, args[0])
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().BoolVar(&o.showInput, "show-input", false, evaluationShowInputFlag)
	cmd.Flags().BoolVar(&o.fixture, "fixture", false, evaluationFixtureFlag)
	cmd.Flags().StringVar(&o.dir, "dir", "", evaluationFixtureDirFlag)

	return cmd
}

// validate refuses a combination before anything is read, so a refusal never
// leaves a half-written fixture behind.
func (o *getEvaluationOptions) validate() error {
	if o.fixture && o.showInput {
		return fmt.Errorf("only one of --fixture or --show-input can be given")
	}
	if o.showInput && o.output != "json" {
		return fmt.Errorf("--show-input needs --output json")
	}
	if o.fixture && o.dir == "" {
		return fmt.Errorf("--fixture needs --dir")
	}
	if o.dir != "" && !o.fixture {
		return fmt.Errorf("--dir needs --fixture")
	}
	if o.fixture {
		return requireEmptyDir(o.dir)
	}
	return nil
}

func (o *getEvaluationOptions) run(out io.Writer, id string) error {
	client := evaluations.NewClient(kosliClient, global.Host, global.ApiToken, global.DryRun)
	evaluation, err := client.Get(global.Org, id)
	if err != nil {
		return err
	}
	raw := evaluation.Raw
	if o.showInput {
		input, err := client.Input(global.Org, id)
		if err != nil {
			return err
		}
		raw, err = withInput(raw, input)
		if err != nil {
			return err
		}
	}

	return output.FormattedPrint(raw, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printEvaluationAsTable,
			"json":  output.PrintJson,
		})
}

// withInput adds the input as the evaluation's last key, leaving the server's
// body as it was sent, keys in their order, rather than re-encoding it.
func withInput(evaluation, input string) (string, error) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal([]byte(evaluation), &fields); err != nil {
		return "", fmt.Errorf("failed to parse the evaluation: %w", err)
	}
	if !json.Valid([]byte(input)) {
		return "", fmt.Errorf("failed to parse the evaluation's input")
	}
	body := strings.TrimSuffix(strings.TrimSpace(evaluation), "}")
	return body + `,"input":` + input + "}", nil
}

// evaluationRead is what the table needs from an evaluation; -o json prints
// the server's body instead.
type evaluationRead struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	RequestedAt float64 `json:"requested_at"`
	RecordedAt  float64 `json:"recorded_at"`
	Context     struct {
		Trails []evaluations.TrailRef `json:"trails"`
	} `json:"context"`
	Policy struct {
		Identifier string            `json:"identifier"`
		Version    int               `json:"version"`
		Digest     string            `json:"digest"`
		Files      map[string]string `json:"files"`
	} `json:"policy"`
	Params                map[string]any        `json:"params"`
	Result                *evaluations.Result   `json:"result"`
	Failure               *evaluations.Failure  `json:"error"`
	OPAVersion            string                `json:"opa_version"`
	Decision              *evaluations.Decision `json:"decision"`
	DecisionAttestationID string                `json:"decision_attestation_id"`
}

func printEvaluationAsTable(raw string, out io.Writer, page int) error {
	var evaluation evaluationRead
	if err := json.Unmarshal([]byte(raw), &evaluation); err != nil {
		return err
	}

	policy := map[string]any{"identifier": evaluation.Policy.Identifier, "digest": evaluation.Policy.Digest}
	if evaluation.Policy.Identifier != "" {
		policy["version"] = float64(evaluation.Policy.Version)
	}
	trails := make([]string, len(evaluation.Context.Trails))
	for i, trail := range evaluation.Context.Trails {
		trails[i] = trail.Flow + "/" + trail.Trail
	}
	rows := []string{
		fmt.Sprintf("ID:\t%s", evaluation.ID),
		fmt.Sprintf("Status:\t%s", evaluation.Status),
		fmt.Sprintf("Policy:\t%s", citedPolicyText(policy)),
		fmt.Sprintf("Trails:\t%s", strings.Join(trails, ", ")),
	}
	for _, field := range []struct {
		label string
		at    float64
	}{
		{"Requested at", evaluation.RequestedAt},
		{"Recorded at", evaluation.RecordedAt},
	} {
		at, err := formattedTimestamp(field.at, true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s:\t%s", field.label, at))
	}

	if evaluation.Status == evaluations.StatusCompleted && evaluation.Result != nil {
		verdict := serverVerdict(evaluation.Result)
		outcome := "DENIED"
		if verdict.Allow {
			outcome = "ALLOWED"
		}
		rows = append(rows, fmt.Sprintf("Outcome:\t%s", outcome))
		violations := verdict.Violations
		if listed, ok := verdict.Outputs["violations"]; ok {
			violations = nil
			for _, violation := range violationList(listed) {
				violations = append(violations, violationText(violation))
			}
		}
		for i, violation := range violations {
			label := ""
			if i == 0 {
				label = "Violations:"
			}
			rows = append(rows, fmt.Sprintf("%s\t%s", label, violation))
		}
	}
	if evaluation.Failure != nil {
		rows = append(rows, fmt.Sprintf("Error:\t%s: %s", evaluation.Failure.Kind, evaluation.Failure.Message))
	}
	if evaluation.OPAVersion != "" {
		rows = append(rows, fmt.Sprintf("OPA version:\t%s", evaluation.OPAVersion))
	}
	if evaluation.Decision != nil {
		rows = append(rows, fmt.Sprintf("Decision:\t%s for control %s in %s/%s",
			evaluation.Decision.Name, evaluation.Decision.Control,
			evaluation.Decision.Flow, evaluation.Decision.Trail))
	}
	if evaluation.DecisionAttestationID != "" {
		rows = append(rows, fmt.Sprintf("Decision attestation:\t%s", evaluation.DecisionAttestationID))
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}

func (o *getEvaluationOptions) runFixture(out io.Writer, id string) error {
	client := evaluations.NewClient(kosliClient, global.Host, global.ApiToken, global.DryRun)
	read, err := client.Get(global.Org, id)
	if err != nil {
		return err
	}
	var evaluation evaluationRead
	if err := json.Unmarshal([]byte(read.Raw), &evaluation); err != nil {
		return fmt.Errorf("failed to parse the evaluation: %w", err)
	}
	input, err := client.Input(global.Org, id)
	if err != nil {
		return err
	}
	files, err := policyFiles(evaluation)
	if err != nil {
		return err
	}

	params := evaluation.Params
	if params == nil {
		params = map[string]any{}
	}
	data, err := json.MarshalIndent(map[string]any{"params": params}, "", "  ")
	if err != nil {
		return err
	}
	var indented bytes.Buffer
	if err := json.Indent(&indented, []byte(input), "", "  "); err != nil {
		return fmt.Errorf("failed to parse the evaluation's input: %w", err)
	}

	written := map[string][]byte{
		"input.json": indented.Bytes(),
		"data.json":  data,
	}
	for name, content := range files {
		if !filepath.IsLocal(name) {
			return fmt.Errorf("%q is not a path inside the policy directory", name)
		}
		written[filepath.Join("policy", name)] = []byte(content)
	}
	for name, content := range written {
		path := filepath.Join(o.dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, content, 0o644); err != nil {
			return err
		}
	}

	_, err = fmt.Fprintf(out, "Wrote the input, params and policy of evaluation %s to %s\n", id, o.dir)
	return err
}

// policyFiles is the bundle that ran: an inline policy's files come with the
// evaluation, and a published version's are read from the rego policy.
func policyFiles(evaluation evaluationRead) (map[string]string, error) {
	if evaluation.Policy.Identifier == "" {
		return evaluation.Policy.Files, nil
	}
	endpoint, err := regopolicies.BundleURL(global.Host, global.Org,
		evaluation.Policy.Identifier, evaluation.Policy.Version)
	if err != nil {
		return nil, err
	}
	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    endpoint,
		Token:  global.ApiToken,
	})
	if err != nil {
		return nil, err
	}
	return untarBundle([]byte(response.Body))
}

func untarBundle(archive []byte) (map[string]string, error) {
	zipped, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, fmt.Errorf("failed to read the policy bundle: %w", err)
	}
	files := map[string]string{}
	reader := tar.NewReader(zipped)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return files, nil
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read the policy bundle: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		content, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to read the policy bundle: %w", err)
		}
		files[header.Name] = string(content)
	}
}

func requireEmptyDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(entries) > 0 {
		return fmt.Errorf("%s is not empty; choose an empty or new directory", dir)
	}
	return nil
}
