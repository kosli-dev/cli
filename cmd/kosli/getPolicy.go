package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getPolicyDesc = `Get a policy's metadata.`

type getPolicyOptions struct {
	output string
}

func newGetPolicyCmd(out io.Writer) *cobra.Command {
	o := new(getPolicyOptions)
	cmd := &cobra.Command{
		Use:    "policy POLICY-NAME",
		Short:  getPolicyDesc,
		Long:   getPolicyDesc,
		Args:   cobra.ExactArgs(1),
		Hidden: true,
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
	url := fmt.Sprintf("%s/api/v2/policies/%s/%s", global.Host, global.Org, args[0])

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
	var policy map[string]interface{}
	err := json.Unmarshal([]byte(raw), &policy)
	if err != nil {
		return err
	}

	createdAt, err := formattedTimestamp(policy["created_at"], false)
	if err != nil {
		return err
	}

	consumingEnvs := policy["consuming_envs"].([]interface{})

	versions := policy["versions"].([]interface{})

	latestVersion := versions[len(versions)-1].(map[string]interface{})
	policyYaml := latestVersion["policy_yaml"].(string)
	policyYamlIndented := "\t" + strings.ReplaceAll(policyYaml, "\n", "\n\t")

	header := []string{}
	rows := []string{}
	rows = append(rows, fmt.Sprintf("Name:\t%s", policy["name"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", policy["description"]))
	rows = append(rows, fmt.Sprintf("Created At:\t%s", createdAt))
	rows = append(rows, fmt.Sprintf("Versions:\t%d", len(versions)))
	rows = append(rows, fmt.Sprintf("Attached to environments:\t%s", consumingEnvs))
	rows = append(rows, fmt.Sprintf("Policy content:\n%s", policyYamlIndented))

	tabFormattedPrint(out, header, rows)

	return nil
}
