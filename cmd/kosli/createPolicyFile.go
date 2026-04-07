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
			p.Artifacts.Provenance = &policy.BooleanRule{Required: true}
		}
		if requireTrailCompliance {
			p.Artifacts.TrailCompliance = &policy.BooleanRule{Required: true}
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

	return rule, nil
}
