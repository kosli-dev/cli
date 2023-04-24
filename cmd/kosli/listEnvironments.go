package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listEnvironmentsDesc = `List environments for an org.`

type environmentLsOptions struct {
	output string
}

func newListEnvironmentsCmd(out io.Writer) *cobra.Command {
	o := new(environmentLsOptions)
	cmd := &cobra.Command{
		Use:     "environments",
		Aliases: []string{"env", "envs"},
		Short:   listEnvironmentsDesc,
		Long:    listEnvironmentsDesc,
		Args:    cobra.NoArgs,
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

func (o *environmentLsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/environments/%s", global.Host, global.Org)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printEnvListAsTable,
			"json":  output.PrintJson,
		})
}

func printEnvListAsTable(raw string, out io.Writer, page int) error {
	var envs []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &envs)
	if err != nil {
		return err
	}

	if len(envs) == 0 {
		logger.Info("No environments were found.")
		return nil
	}

	header := []string{"NAME", "TYPE", "LAST REPORT", "LAST MODIFIED"}
	rows := []string{}
	for _, env := range envs {
		last_reported_str := ""
		last_reported_at := env["last_reported_at"]
		if last_reported_at != nil {
			last_reported_str = time.Unix(int64(last_reported_at.(float64)), 0).Format(time.RFC3339)
		}
		last_modified_str := ""
		last_modified_at := env["last_modified_at"]
		if last_modified_at != nil {
			last_modified_str = time.Unix(int64(last_modified_at.(float64)), 0).Format(time.RFC3339)
		}
		row := fmt.Sprintf("%s\t%s\t%s\t%s", env["name"], env["type"], last_reported_str, last_modified_str)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
