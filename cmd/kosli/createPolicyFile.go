package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
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

	hasAPI := global.ApiToken != "" && global.Org != ""
	host := global.Host
	ctx := &policywizard.Context{
		HasAPICredentials: hasAPI,
		Org:               global.Org,
		Host:              host,
		WriteFunc:         writeAndUploadPolicy,
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
	_, err := tea.NewProgram(m, tea.WithAltScreen()).Run()
	if err != nil {
		return fmt.Errorf("wizard error: %w", err)
	}

	return nil
}

func writeAndUploadPolicy(req policywizard.WriteRequest) policywizard.WriteResult {
	result := policywizard.WriteResult{
		Filename:   req.Filename,
		PolicyName: req.PolicyName,
	}

	// Write file
	err := os.WriteFile(req.Filename, req.YAMLBytes, 0644)
	if err != nil {
		result.Err = fmt.Errorf("failed to write policy file: %w", err)
		return result
	}

	// Upload if requested
	if req.Upload && req.PolicyName != "" {
		if req.Org != "" {
			global.Org = req.Org
		}
		err = uploadPolicy(req.PolicyName, req.Description, req.Filename)
		if err != nil {
			result.Err = fmt.Errorf("policy file saved to %s but upload failed: %w", req.Filename, err)
			return result
		}
		result.Uploaded = true
		result.PolicyURL = policyURL(global.Host, req.Org, req.PolicyName)
	}

	return result
}

func policyURL(host, org, policyName string) string {
	// Map API host to UI host
	uiHost := host
	uiHost = strings.TrimSuffix(uiHost, "/")
	// The API host is the same as the UI host for Kosli
	return fmt.Sprintf("%s/%s/policies/%s", uiHost, org, policyName)
}

func uploadPolicy(name, description, policyFile string) error {
	o := &createPolicyOptions{
		payload: PolicyPayload{
			Name:        name,
			Description: description,
			Type:        "env",
		},
	}
	return o.run([]string{name, policyFile})
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
