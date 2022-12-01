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

const pipelineInspectDesc = `Inspect the metadata of a single pipeline`

type pipelineInspectOptions struct {
	output string
}

func newPipelineInspectCmd(out io.Writer) *cobra.Command {
	o := new(pipelineInspectOptions)
	cmd := &cobra.Command{
		Use:   "inspect PIPELINE-NAME",
		Short: pipelineInspectDesc,
		Long:  pipelineInspectDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "pipeline name argument is required")
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

func (o *pipelineInspectOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s", global.Host, global.Owner, args[0])
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{})

	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printPipelineAsTable,
			"json":  output.PrintJson,
		})
}

func printPipelineAsTable(raw string, out io.Writer, page int) error {
	var pipeline map[string]interface{}
	err := json.Unmarshal([]byte(raw), &pipeline)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}

	lastDeployedAt, err := formattedTimestamp(pipeline["last_deployment_at"], false)
	if err != nil {
		return err
	}
	template := fmt.Sprintf("%s", pipeline["template"])
	template = strings.Replace(template, " ", ", ", -1)

	rows = append(rows, fmt.Sprintf("Name:\t%s", pipeline["name"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", pipeline["description"]))
	rows = append(rows, fmt.Sprintf("Visibility:\t%s", pipeline["visibility"]))
	rows = append(rows, fmt.Sprintf("Template:\t%s", template))
	rows = append(rows, fmt.Sprintf("Last Deployment At:\t%s", lastDeployedAt))

	tabFormattedPrint(out, header, rows)
	return nil
}
