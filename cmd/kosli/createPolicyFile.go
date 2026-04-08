package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

	ctx := &policywizard.Context{}
	if global.ApiToken != "" && global.Org != "" {
		fmt.Fprint(os.Stderr, "Starting Kosli Policy Builder...\r")
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

	wm, ok := finalModel.(policywizard.Model)
	if !ok {
		return fmt.Errorf("unexpected model type from wizard")
	}
	if wm.Cancelled {
		logger.Info("policy file creation cancelled")
		return nil
	}

	yamlBytes, err := wm.Policy.ToYAML()
	if err != nil {
		return fmt.Errorf("failed to generate policy YAML: %w", err)
	}

	outPath := filepath.Clean(wm.OutputFile)
	if err := validateOutputFile(outPath); err != nil {
		return err
	}

	err = os.WriteFile(outPath, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}
	logger.Info("policy file written to %s", outPath)
	return nil
}

func validateOutputFile(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".yaml" && ext != ".yml" {
		return fmt.Errorf("output file must have a .yaml or .yml extension, got %q", filepath.Base(path))
	}
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file %q already exists; remove it or choose a different name", path)
	}
	return nil
}

func fetchFlowNames() []string {
	return fetchNameList("api/v2/flows", nil)
}

func fetchCustomAttestationTypes() []string {
	return fetchNameList("api/v2/custom-attestation-types", func(name string) string {
		return "custom:" + name
	})
}

func fetchNameList(apiPath string, transform func(string) string) []string {
	u, err := url.JoinPath(global.Host, apiPath, global.Org)
	if err != nil {
		logger.Debug("failed to build URL for %s: %v", apiPath, err)
		return nil
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("failed to fetch %s: %v", apiPath, err)
		return nil
	}

	var items []map[string]any
	if err := json.Unmarshal([]byte(response.Body), &items); err != nil {
		logger.Debug("failed to parse %s response: %v", apiPath, err)
		return nil
	}

	names := make([]string, 0, len(items))
	for _, item := range items {
		if name, ok := item["name"].(string); ok {
			if transform != nil {
				name = transform(name)
			}
			names = append(names, name)
		}
	}
	return names
}
