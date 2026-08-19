package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createAttestationTypeShortDesc = `Create or update a Kosli custom attestation type.`

const createAttestationTypeLongDesc = createAttestationTypeShortDesc + `
You can specify attestation type parameters in flags.

^TYPE-NAME^ must start with a letter or number, and only contain letters, numbers, ^.^, ^-^, ^_^, and ^~^.

^--schema^ is a path to a file containing a JSON schema which will be used to validate attestations made using this type.
The schema is used to specify the structure of the attestation data, e.g. any fields that are required or
the expected type of the data.
See an example schema file
[here](https://github.com/cyber-dojo/kosli-attestation-types/blob/f9130c58d3a8151b0b0e7c5db284e4380eb2d2cf/metrics-coverage.schema.json).

^--jq^ defines an evaluation rule, given in jq-format, for this attestation type. The flag can be repeated in order to add additional rules.
These rules specify acceptable values for attestation data, e.g. ^.age >= 21^ or ^.failing_tests == 0^.
When a custom attestation is reported, the provided data is evaluated according to the rules defined in its attestation-type.
All rules must return ^true^ for the evaluation to pass and the attestation to be determined compliant.

^--summary^ defines one entry of the summary shown for attestations of this type, given as
^'NAME=EXPRESSION'^ where the expression is a ^jq^ expression evaluated against the attestation data.
The flag can be repeated to add further entries, which are displayed in the order given, e.g.
^--summary "Critical=.critical_count" --summary "Tool=.scanner.name"^.
Each value is split on its first ^=^ only, so ^jq^ expressions containing ^==^ are unaffected.

^--summary-json^ is an alternative to ^--summary^ for summaries that are easier to express as JSON,
given as a JSON array of ^{"name": ..., "expression": ...}^ entries, e.g.
^'[{"name":"Critical","expression":".critical_count"}]'^. The two summary flags cannot be combined.

Attestation types created without a summary fall back to the ^jq^ evaluation rules checklist.
`

const createAttestationTypeExample = `
# create/update a custom attestation type with no schema no evaluation rules:
kosli create attestation-type customTypeName

# create/update a custom attestation type with schema and ^jq^ evaluation rules:
kosli create attestation-type customTypeName \
    --description "Attest that a person meets the age requirements." \
    --schema person-schema.json \
    --jq ".age >= 18"
    --jq ".age < 65"

# create/update a custom attestation type with a summary:
kosli create attestation-type customTypeName \
    --schema scan-schema.json \
    --summary "Critical=.critical_count" \
    --summary "Tool=.scanner.name"

# create/update a custom attestation type with a summary given as JSON:
kosli create attestation-type customTypeName \
    --schema scan-schema.json \
    --summary-json '[{"name":"Critical","expression":".critical_count"},{"name":"Tool","expression":".scanner.name"}]'
`

type createAttestationTypeOptions struct {
	payload         CreateAttestationTypePayload
	schemaFilePath  string
	jqRules         []string
	summaryJSON     string
	summaryKeyValue []string
}

// SummaryEntry is one named jq expression displayed in the summary of
// attestations made using a custom attestation type.
type SummaryEntry struct {
	Name       string `json:"name"`
	Expression string `json:"expression"`
}

type JQEvaluatorPayload struct {
	ContentType string   `json:"content_type"`
	Rules       []string `json:"rules"`
}

func NewJQEvaluatorPayload(rules []string) *JQEvaluatorPayload {
	return &JQEvaluatorPayload{"jq", rules}
}

type CreateAttestationTypePayload struct {
	TypeName    string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Evaluator   *JQEvaluatorPayload `json:"evaluator,omitempty"`
	Summary     []SummaryEntry      `json:"summary,omitempty"`
}

// parseSummaryJSON parses the --summary-json flag value into an ordered list of
// summary entries. The value must be a JSON array of {name, expression} objects,
// both fields non-empty.
func parseSummaryJSON(value string) ([]SummaryEntry, error) {
	// Only a wholly absent value is no summary. The empty-value rule refuses an
	// empty one on the command line, so "" here means the flag was left out;
	// whitespace was written on purpose and falls through to be rejected.
	if value == "" {
		return nil, nil
	}

	var summary []SummaryEntry
	if err := json.Unmarshal([]byte(value), &summary); err != nil {
		var syntaxErr *json.SyntaxError
		if errors.As(err, &syntaxErr) {
			return nil, fmt.Errorf("--summary-json is not valid JSON: %s", err.Error())
		}
		return nil, fmt.Errorf("--summary-json must be a JSON array of {name, expression} entries")
	}

	for i, entry := range summary {
		if strings.TrimSpace(entry.Name) == "" {
			return nil, fmt.Errorf("--summary-json entry %d is missing a name", i+1)
		}
		if strings.TrimSpace(entry.Expression) == "" {
			return nil, fmt.Errorf("--summary-json entry %d is missing an expression", i+1)
		}
	}

	return summary, nil
}

// parseSummaryFlags parses repeated --summary NAME=EXPRESSION values into an
// ordered list of summary entries, preserving the order the flags were given in.
// Each value is split on its first "=" only, so JQ expressions containing "=="
// or assignments keep their meaning; whitespace around both halves is trimmed.
func parseSummaryFlags(values []string) ([]SummaryEntry, error) {
	if len(values) == 0 {
		return nil, nil
	}

	summary := make([]SummaryEntry, 0, len(values))
	for i, value := range values {
		name, expression, found := strings.Cut(value, "=")
		if !found {
			return nil, fmt.Errorf("--summary entry %d must be in the form NAME=EXPRESSION", i+1)
		}

		name = strings.TrimSpace(name)
		expression = strings.TrimSpace(expression)
		if name == "" {
			return nil, fmt.Errorf("--summary entry %d is missing a name", i+1)
		}
		if expression == "" {
			return nil, fmt.Errorf("--summary entry %d is missing an expression", i+1)
		}
		// Catches a bare jq expression passed without a name: ".failing == 0"
		// splits to {".failing", "= 0"}. No valid jq expression starts with "=",
		// so a leading one always means the value was not NAME=EXPRESSION.
		if strings.HasPrefix(expression, "=") {
			return nil, fmt.Errorf("--summary entry %d expression cannot start with '='", i+1)
		}

		summary = append(summary, SummaryEntry{Name: name, Expression: expression})
	}

	return summary, nil
}

func newCreateAttestationTypeCmd(out io.Writer) *cobra.Command {
	o := new(createAttestationTypeOptions)
	cmd := &cobra.Command{
		Use:     "attestation-type TYPE-NAME",
		Short:   createAttestationTypeShortDesc,
		Long:    createAttestationTypeLongDesc,
		Example: createAttestationTypeExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			// Both flags set the same payload field, so letting one silently win
			// would be a trap.
			return MuXRequiredFlags(cmd, []string{"summary", "summary-json"}, false)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", attestationTypeDescriptionFlag)
	cmd.Flags().StringVarP(&o.schemaFilePath, "schema", "s", "", attestationTypeSchemaFlag)
	cmd.Flags().StringArrayVar(&o.jqRules, "jq", []string{}, attestationTypeJqFlag)
	cmd.Flags().StringArrayVar(&o.summaryKeyValue, "summary", []string{}, attestationTypeSummaryFlag)
	cmd.Flags().StringVar(&o.summaryJSON, "summary-json", "", attestationTypeSummaryJsonFlag)

	addDryRunFlag(cmd)
	return cmd
}

func (o *createAttestationTypeOptions) run(args []string) error {
	o.payload.TypeName = args[0]
	if len(o.jqRules) > 0 {
		o.payload.Evaluator = NewJQEvaluatorPayload(o.jqRules)
	}

	summary, err := parseSummaryJSON(o.summaryJSON)
	if err != nil {
		return err
	}
	if summary == nil {
		// --summary and --summary-json are mutually exclusive, so at most one of
		// these two parses can yield entries.
		summary, err = parseSummaryFlags(o.summaryKeyValue)
		if err != nil {
			return err
		}
	}
	o.payload.Summary = summary

	form, err := prepareAttestationTypeForm(o.payload, o.schemaFilePath)
	if err != nil {
		return err
	}

	url, err := url.JoinPath(global.Host, "api/v2/custom-attestation-types", global.Org)
	if err != nil {
		return err
	}
	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("attestation-type %s was created", o.payload.TypeName)
	}
	return err
}

func prepareAttestationTypeForm(payload interface{}, schemaFilePath string) ([]requests.FormItem, error) {
	form, err := newAttestationTypeForm(payload, schemaFilePath)
	if err != nil {
		return []requests.FormItem{}, err
	}
	return form, nil
}

// newAttestationTypeForm constructs a list of FormItems for an attestation-type
// form submission.
func newAttestationTypeForm(payload interface{}, schemaFilePath string) (
	[]requests.FormItem, error,
) {
	form := []requests.FormItem{
		{Type: "field", FieldName: "data_json", Content: payload},
	}

	if schemaFilePath != "" {
		form = append(form, requests.FormItem{Type: "file", FieldName: "type_schema", Content: schemaFilePath})
	}

	return form, nil
}
