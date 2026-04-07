package main

import (
	"fmt"
	"io"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kosli-dev/cli/internal/policy"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const createPolicyFileShortDesc = `Interactively create a Kosli environment policy YAML file.`

const createPolicyFileLongDesc = createPolicyFileShortDesc + `
Launches an interactive wizard that guides you through building a policy file
conforming to the Kosli environment policy schema. The generated YAML is
written to a file you specify at the end of the wizard.

This command does not upload the policy to Kosli. Use ^kosli create policy^
to upload the generated file.

If ^--api-token^ and ^--org^ are set, the wizard will fetch flow names and
custom attestation types from the Kosli API to offer as suggestions.
`

const createPolicyFileExample = `
# create a policy file interactively:
kosli create policy-file
`

func newCreatePolicyFileCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "policy-file",
		Short:   createPolicyFileShortDesc,
		Long:    createPolicyFileLongDesc,
		Example: createPolicyFileExample,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreatePolicyFile()
		},
	}

	return cmd
}

func runCreatePolicyFile() error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("this command requires an interactive terminal; write policy YAML manually or use 'kosli create policy' directly")
	}

	wctx := &wizardContext{}
	if global.ApiToken != "" && global.Org != "" {
		wctx.fetchFromAPI()
	}

	m := newPolicyWizardModel(wctx)
	finalModel, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	wm := finalModel.(policyWizardModel)
	if wm.cancelled {
		logger.Info("policy file creation cancelled")
		return nil
	}

	return writePolicyFile(wm.policy, wm.outputFile)
}

func writePolicyFile(p *policy.Policy, filename string) error {
	if filename == "" {
		filename = "policy.yaml"
	}

	yamlBytes, err := p.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate policy YAML: %w", err)
	}

	err = os.WriteFile(filename, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}
	logger.Info("policy file written to %s", filename)
	return nil
}
