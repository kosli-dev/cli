package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listPoliciesDesc = `List environment policies for an org.`

type policiesLsOptions struct {
	output string
}

func newListPoliciesCmd(out io.Writer) *cobra.Command {
	o := new(policiesLsOptions)
	cmd := &cobra.Command{
		Use:   "policies",
		Short: listPoliciesDesc,
		Long:  listPoliciesDesc,
		Args:  cobra.NoArgs,
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

func (o *policiesLsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/policies/%s", global.Host, global.Org)

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
			"table": printPolicyListAsTable,
			"json":  output.PrintJson,
		})
}

func printPolicyListAsTable(raw string, out io.Writer, page int) error {
	var policies []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &policies)
	if err != nil {
		return err
	}

	if len(policies) == 0 {
		logger.Info("No environment policies were found.")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "CREATED AT", "LAST VERSION", "USED BY ENVS"}
	rows := []string{}
	for _, policy := range policies {
		createdAt, err := formattedTimestamp(policy["created_at"], false)
		if err != nil {
			return err
		}

		versions := policy["versions"].([]interface{})

		usedByEnvs := policy["consuming_envs"].([]interface{})

		row := fmt.Sprintf("%s\t%s\t%s\t%d\t%s", policy["name"], policy["description"], createdAt, len(versions), usedByEnvs)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
