package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/kosli-dev/cli/internal/policy"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// wizardContext holds data fetched from the API to populate wizard options.
type wizardContext struct {
	flowNames            []string
	customAttestTypes    []string // e.g. ["custom:coverage-metrics", "custom:compliance-check"]
}

const createPolicyFileShortDesc = `Interactively create a Kosli environment policy YAML file.`

const createPolicyFileLongDesc = createPolicyFileShortDesc + `
Launches an interactive wizard that guides you through building a policy file
conforming to the Kosli environment policy schema. The generated YAML is
written to stdout by default, or to a file with ^--output-file^.

This command does not upload the policy to Kosli. Use ^kosli create policy^
to upload the generated file.

If ^--api-token^ and ^--org^ are set, the wizard will fetch flow names and
custom attestation types from the Kosli API to offer as suggestions.
`

const createPolicyFileExample = `
# create a policy file interactively (output to stdout):
kosli create policy-file

# create a policy file and write to a file:
kosli create policy-file --output-file policy.yml
`

type createPolicyFileOptions struct {
	outputFile string
}

func newCreatePolicyFileCmd(out io.Writer) *cobra.Command {
	o := new(createPolicyFileOptions)
	cmd := &cobra.Command{
		Use:     "policy-file",
		Short:   createPolicyFileShortDesc,
		Long:    createPolicyFileLongDesc,
		Example: createPolicyFileExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.outputFile, "output-file", "o", "", "write policy YAML to this file instead of stdout")

	return cmd
}

func (o *createPolicyFileOptions) run(out io.Writer) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("this command requires an interactive terminal; write policy YAML manually or use 'kosli create policy' directly")
	}

	wctx := &wizardContext{}
	if global.ApiToken != "" && global.Org != "" {
		wctx.fetchFromAPI()
	}

	p := policy.NewPolicy()

	var requireProvenance bool
	var requireTrailCompliance bool

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Require artifact provenance?").
				Description("All artifacts must belong to a Kosli flow").
				Value(&requireProvenance).
				Affirmative("Yes").
				Negative("No"),
			huh.NewConfirm().
				Title("Require trail compliance?").
				Description("All artifacts must be part of compliant trails").
				Value(&requireTrailCompliance).
				Affirmative("Yes").
				Negative("No"),
		),
	).Run()
	if err != nil {
		return err
	}

	if requireProvenance || requireTrailCompliance {
		p.Artifacts = &policy.ArtifactRules{}
		if requireProvenance {
			exceptions, exErr := collectExceptions("provenance", wctx)
			if exErr != nil {
				return exErr
			}
			p.Artifacts.Provenance = &policy.BooleanRule{Required: true, Exceptions: exceptions}
		}
		if requireTrailCompliance {
			exceptions, exErr := collectExceptions("trail compliance", wctx)
			if exErr != nil {
				return exErr
			}
			p.Artifacts.TrailCompliance = &policy.BooleanRule{Required: true, Exceptions: exceptions}
		}
	}

	// Attestation loop
	attestations, err := collectAttestations(wctx)
	if err != nil {
		return err
	}
	if len(attestations) > 0 {
		if p.Artifacts == nil {
			p.Artifacts = &policy.ArtifactRules{}
		}
		p.Artifacts.Attestations = attestations
	}

	yamlBytes, err := p.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate policy YAML: %w", err)
	}

	// Preview the generated YAML
	var confirm bool
	err = huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Generated policy YAML").
				Description(string(yamlBytes)),
			huh.NewConfirm().
				Title("Write this policy?").
				Value(&confirm).
				Affirmative("Yes").
				Negative("No"),
		),
	).Run()
	if err != nil {
		return err
	}
	if !confirm {
		logger.Info("policy file creation cancelled")
		return nil
	}

	if o.outputFile != "" {
		err = os.WriteFile(o.outputFile, yamlBytes, 0644)
		if err != nil {
			return fmt.Errorf("failed to write policy file: %w", err)
		}
		logger.Info("policy file written to %s", o.outputFile)
		return nil
	}

	_, err = out.Write(yamlBytes)
	return err
}

var builtInAttestationTypes = []string{
	"generic",
	"junit",
	"snyk",
	"pull_request",
	"jira",
	"sonar",
}

func collectAttestations(wctx *wizardContext) ([]policy.AttestationRule, error) {
	var attestations []policy.AttestationRule

	for {
		prompt := "Add a required attestation?"
		if len(attestations) > 0 {
			prompt = "Add another required attestation?"
		}

		var addAttestation bool
		err := huh.NewConfirm().
			Title(prompt).
			Value(&addAttestation).
			Affirmative("Yes").
			Negative("No").
			Run()
		if err != nil {
			return nil, err
		}
		if !addAttestation {
			break
		}

		rule, err := collectOneAttestation(wctx)
		if err != nil {
			return nil, err
		}
		attestations = append(attestations, rule)
	}
	return attestations, nil
}

func collectOneAttestation(wctx *wizardContext) (policy.AttestationRule, error) {
	var attType string
	var attName string

	allTypes := builtInAttestationTypes
	if len(wctx.customAttestTypes) > 0 {
		allTypes = append(allTypes, wctx.customAttestTypes...)
	}
	typeOptions := make([]huh.Option[string], len(allTypes))
	for i, t := range allTypes {
		typeOptions[i] = huh.NewOption(t, t)
	}

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Attestation type").
				Options(typeOptions...).
				Value(&attType),
			huh.NewInput().
				Title("Attestation name").
				Description("Use * to match any name for this type").
				Placeholder("*").
				Value(&attName),
		),
	).Run()
	if err != nil {
		return policy.AttestationRule{}, err
	}

	if attName == "" {
		attName = "*"
	}

	rule := policy.AttestationRule{
		Type: attType,
		Name: attName,
	}

	var addCondition bool
	err = huh.NewConfirm().
		Title("Add a condition for this attestation?").
		Description("Only require this attestation when the condition is met").
		Value(&addCondition).
		Affirmative("Yes").
		Negative("No").
		Run()
	if err != nil {
		return policy.AttestationRule{}, err
	}
	if addCondition {
		expr, exprErr := collectExpression(wctx)
		if exprErr != nil {
			return policy.AttestationRule{}, exprErr
		}
		rule.If = expr
	}

	return rule, nil
}

func collectExceptions(ruleName string, wctx *wizardContext) ([]policy.ExceptionRule, error) {
	var exceptions []policy.ExceptionRule

	for {
		prompt := fmt.Sprintf("Add an exception to %s?", ruleName)
		if len(exceptions) > 0 {
			prompt = fmt.Sprintf("Add another exception to %s?", ruleName)
		}

		var addException bool
		err := huh.NewConfirm().
			Title(prompt).
			Description("Exceptions waive this requirement for matching artifacts").
			Value(&addException).
			Affirmative("Yes").
			Negative("No").
			Run()
		if err != nil {
			return nil, err
		}
		if !addException {
			break
		}

		expr, err := collectExpression(wctx)
		if err != nil {
			return nil, err
		}
		exceptions = append(exceptions, policy.ExceptionRule{If: expr})
	}
	return exceptions, nil
}

const (
	exprModeFlowName     = "flow_name"
	exprModeFlowTag      = "flow_tag"
	exprModeArtifactName = "artifact_name"
	exprModeCustom       = "custom"
	exprModeRaw          = "raw"
)

func collectExpression(wctx *wizardContext) (string, error) {
	var mode string
	err := huh.NewSelect[string]().
		Title("How do you want to define this condition?").
		Options(
			huh.NewOption("Match by flow name", exprModeFlowName),
			huh.NewOption("Match by flow tag", exprModeFlowTag),
			huh.NewOption("Match by artifact name pattern", exprModeArtifactName),
			huh.NewOption("Custom comparison", exprModeCustom),
			huh.NewOption("Write raw expression", exprModeRaw),
		).
		Value(&mode).
		Run()
	if err != nil {
		return "", err
	}

	switch mode {
	case exprModeFlowName:
		return collectFlowNameExpr(wctx)
	case exprModeFlowTag:
		return collectFlowTagExpr()
	case exprModeArtifactName:
		return collectArtifactNameExpr()
	case exprModeCustom:
		return collectCustomExpr()
	case exprModeRaw:
		return collectRawExpr()
	}
	return "", fmt.Errorf("unknown expression mode: %s", mode)
}

func collectFlowNameExpr(wctx *wizardContext) (string, error) {
	var flowName string

	if len(wctx.flowNames) > 0 {
		options := make([]huh.Option[string], len(wctx.flowNames))
		for i, name := range wctx.flowNames {
			options[i] = huh.NewOption(name, name)
		}
		err := huh.NewSelect[string]().
			Title("Select a flow").
			Options(options...).
			Value(&flowName).
			Run()
		if err != nil {
			return "", err
		}
	} else {
		err := huh.NewInput().
			Title("Flow name").
			Description("The flow name to match").
			Value(&flowName).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("flow name is required")
				}
				return nil
			}).
			Run()
		if err != nil {
			return "", err
		}
	}
	return policy.FlowNameExpr(flowName), nil
}

var flowTagOperators = []string{"==", "!=", ">", "<", ">=", "<="}

func collectFlowTagExpr() (string, error) {
	var tagKey string
	var operator string
	var value string

	err := huh.NewInput().
		Title("Tag key").
		Description("The flow tag key (e.g. team, risk-level, key.with.dots)").
		Value(&tagKey).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("tag key is required")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Operator").
				Options(huh.NewOptions(flowTagOperators...)...).
				Value(&operator),
			huh.NewInput().
				Title("Value").
				Description("The value to compare against").
				Value(&value).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("value is required")
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}

	return policy.FlowTagExpr(tagKey, operator, value), nil
}

func collectArtifactNameExpr() (string, error) {
	var regex string
	err := huh.NewInput().
		Title("Artifact name regex").
		Description("Regular expression to match artifact names (e.g. ^datadog:.*)").
		Placeholder("^datadog:.*").
		Value(&regex).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("regex is required")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}
	return policy.ArtifactNameMatchExpr(regex), nil
}

var exprContexts = []string{
	"flow.name",
	"flow.tags.<key>",
	"artifact.name",
	"artifact.fingerprint",
}

var exprOperators = []string{"==", "!=", "in", "matches"}

func collectCustomExpr() (string, error) {
	var context string
	var operator string
	var value string

	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Context field").
				Options(huh.NewOptions(exprContexts...)...).
				Value(&context),
		),
	).Run()
	if err != nil {
		return "", err
	}

	if context == "flow.tags.<key>" {
		var tagKey string
		err = huh.NewInput().
			Title("Tag key").
			Description("The flow tag key (e.g. team, risk-level)").
			Value(&tagKey).
			Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("tag key is required")
				}
				return nil
			}).
			Run()
		if err != nil {
			return "", err
		}
		context = "flow.tags." + tagKey
	}

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Operator").
				Options(huh.NewOptions(exprOperators...)...).
				Value(&operator),
			huh.NewInput().
				Title("Value").
				Description("The value to compare against").
				Value(&value).
				Validate(func(s string) error {
					if s == "" {
						return fmt.Errorf("value is required")
					}
					return nil
				}),
		),
	).Run()
	if err != nil {
		return "", err
	}

	return policy.ComparisonExpr(context, operator, value), nil
}

func collectRawExpr() (string, error) {
	var raw string
	err := huh.NewInput().
		Title("Raw expression").
		Description("Enter a policy expression (e.g. flow.name == \"prod\" and artifact.name == \"svc\")").
		Placeholder(`flow.name == "prod"`).
		Value(&raw).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("expression is required")
			}
			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}
	return policy.WrapExpr(raw), nil
}

func (wctx *wizardContext) fetchFromAPI() {
	wctx.flowNames = fetchFlowNames()
	wctx.customAttestTypes = fetchCustomAttestationTypes()
}

func fetchFlowNames() []string {
	u, err := url.JoinPath(global.Host, "api/v2/flows", global.Org)
	if err != nil {
		logger.Debug("failed to build flows URL: %v", err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch flows: %v", err)
		return nil
	}

	var flows []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &flows); err != nil {
		logger.Debug("failed to parse flows response: %v", err)
		return nil
	}

	names := make([]string, 0, len(flows))
	for _, flow := range flows {
		if name, ok := flow["name"].(string); ok {
			names = append(names, name)
		}
	}
	return names
}

func fetchCustomAttestationTypes() []string {
	u, err := url.JoinPath(global.Host, "api/v2/custom-attestation-types", global.Org)
	if err != nil {
		logger.Debug("failed to build attestation types URL: %v", err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch attestation types: %v", err)
		return nil
	}

	var types []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &types); err != nil {
		logger.Debug("failed to parse attestation types response: %v", err)
		return nil
	}

	names := make([]string, 0, len(types))
	for _, t := range types {
		if name, ok := t["name"].(string); ok {
			names = append(names, "custom:"+name)
		}
	}
	return names
}
