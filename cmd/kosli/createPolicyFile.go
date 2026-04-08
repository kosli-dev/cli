package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kosli-dev/cli/internal/policywizard"
	"github.com/kosli-dev/cli/internal/requests"
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

	hasAPI := global.ApiToken != "" && global.Org != ""
	ctx := &policywizard.Context{
		HasAPICredentials: hasAPI,
	}
	if hasAPI {
		ctx.FetchFunc = func() policywizard.FetchResult {
			return policywizard.FetchResult{
				FlowNames:         fetchFlowNames(),
				CustomAttestTypes: fetchCustomAttestationTypes(),
			}
		}
	}

	m := policywizard.NewModel(ctx)
	finalModel, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	wm := finalModel.(policywizard.Model)
	if wm.Cancelled {
		logger.Info("policy file creation cancelled")
		return nil
	}

	yamlBytes, err := wm.Policy.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate policy YAML: %w", err)
	}

	err = os.WriteFile(wm.OutputFile, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}
	logger.Info("policy file written to %s", wm.OutputFile)
	return nil
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
