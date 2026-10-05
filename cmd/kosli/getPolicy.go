package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getPolicyDesc = `Get a policy's metadata.`

type getPolicyOptions struct {
	output string
}

type policyResponse struct {
	Name          string          `json:"name"`
	Description   string          `json:"description"`
	CreatedAt     json.Number     `json:"created_at"`
	ConsumingEnvs []string        `json:"consuming_envs"`
	Versions      []policyVersion `json:"versions"`
}

type policyVersion struct {
	PolicyYaml string `json:"policy_yaml"`
}

func newGetPolicyCmd(out io.Writer) *cobra.Command {
	o := new(getPolicyOptions)
	cmd := &cobra.Command{
		Use:   "policy POLICY-NAME",
		Short: getPolicyDesc,
		Long:  getPolicyDesc,
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *getPolicyOptions) run(out io.Writer, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/policies", global.Org, args[0])
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printPolicyAsTable,
			"json":  output.PrintJson,
		})
}

func printPolicyAsTable(raw string, out io.Writer, page int) error {
	var policy policyResponse
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return err
	}
	if len(policy.Versions) == 0 {
		return fmt.Errorf("policy '%s' has no versions", policy.Name)
	}
	createdAt, err := formattedTimestamp(policy.CreatedAt, false)
	if err != nil {
		return err
	}

	latest := policy.Versions[len(policy.Versions)-1]
	policyYamlIndented := "\t" + strings.ReplaceAll(latest.PolicyYaml, "\n", "\n\t")

	rows := []string{
		fmt.Sprintf("Name:\t%s", policy.Name),
		fmt.Sprintf("Description:\t%s", policy.Description),
		fmt.Sprintf("Created At:\t%s", createdAt),
		fmt.Sprintf("Versions:\t%d", len(policy.Versions)),
		fmt.Sprintf("Attached to environments:\t%s", policy.ConsumingEnvs),
		fmt.Sprintf("Policy content:\n%s", policyYamlIndented),
	}
	tabFormattedPrint(out, []string{}, rows)

	return nil
}
