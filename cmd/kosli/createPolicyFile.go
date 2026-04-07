package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/kosli-dev/cli/internal/policy"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

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
			exceptions, exErr := collectExceptions("provenance")
			if exErr != nil {
				return exErr
			}
			p.Artifacts.Provenance = &policy.BooleanRule{Required: true, Exceptions: exceptions}
		}
		if requireTrailCompliance {
			exceptions, exErr := collectExceptions("trail compliance")
			if exErr != nil {
				return exErr
			}
			p.Artifacts.TrailCompliance = &policy.BooleanRule{Required: true, Exceptions: exceptions}
		}
	}

	// Attestation loop
	attestations, err := collectAttestations()
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

func collectAttestations() ([]policy.AttestationRule, error) {
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

		rule, err := collectOneAttestation()
		if err != nil {
			return nil, err
		}
		attestations = append(attestations, rule)
	}
	return attestations, nil
}

func collectOneAttestation() (policy.AttestationRule, error) {
	var attType string
	var attName string

	typeOptions := make([]huh.Option[string], len(builtInAttestationTypes))
	for i, t := range builtInAttestationTypes {
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
		expr, exprErr := collectExpression()
		if exprErr != nil {
			return policy.AttestationRule{}, exprErr
		}
		rule.If = expr
	}

	return rule, nil
}

func collectExceptions(ruleName string) ([]policy.ExceptionRule, error) {
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

		expr, err := collectExpression()
		if err != nil {
			return nil, err
		}
		exceptions = append(exceptions, policy.ExceptionRule{If: expr})
	}
	return exceptions, nil
}

const (
	exprModeFlowName     = "flow_name"
	exprModeArtifactName = "artifact_name"
	exprModeCustom       = "custom"
	exprModeRaw          = "raw"
)

func collectExpression() (string, error) {
	var mode string
	err := huh.NewSelect[string]().
		Title("How do you want to define this condition?").
		Options(
			huh.NewOption("Match by flow name", exprModeFlowName),
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
		return collectFlowNameExpr()
	case exprModeArtifactName:
		return collectArtifactNameExpr()
	case exprModeCustom:
		return collectCustomExpr()
	case exprModeRaw:
		return collectRawExpr()
	}
	return "", fmt.Errorf("unknown expression mode: %s", mode)
}

func collectFlowNameExpr() (string, error) {
	var flowName string
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
	return policy.FlowNameExpr(flowName), nil
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
