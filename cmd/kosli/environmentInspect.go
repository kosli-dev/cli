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

const environmentInspectDesc = `Inspect an environment metadata.`

type environmentInspectOptions struct {
	output string
}

func newEnvironmentInspectCmd(out io.Writer) *cobra.Command {
	o := new(environmentInspectOptions)
	cmd := &cobra.Command{
		Use:   "inspect [ENVIRONMENT-NAME]",
		Short: environmentInspectDesc,
		Long:  environmentInspectDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "environment name argument is required")
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

func (o *environmentInspectOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s", global.Host, global.Owner, args[0])
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printEnvironmentAsTable,
			"json":  output.PrintJson,
		})
}

func printEnvironmentAsTable(raw string, out io.Writer, page int) error {
	var env map[string]interface{}
	err := json.Unmarshal([]byte(raw), &env)
	if err != nil {
		return err
	}

	lastReportedAt, err := formattedTimestamp(env["last_reported_at"], false)
	if err != nil {
		return err
	}

	state := "N/A"
	if env["state"] != nil && env["state"].(bool) {
		state = "COMPLIANT"
	} else if env["state"] != nil {
		state = "INCOMPLIANT"
	}

	header := []string{}
	rows := []string{}
	rows = append(rows, fmt.Sprintf("Name:\t%s", env["name"]))
	rows = append(rows, fmt.Sprintf("Type:\t%s", env["type"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", env["description"]))
	rows = append(rows, fmt.Sprintf("State:\t%s", state))
	rows = append(rows, fmt.Sprintf("Last Reported At:\t%s", lastReportedAt))

	tabFormattedPrint(out, header, rows)

	return nil
}
