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

const pipelineLsDesc = `List pipelines for an org.`

type pipelineLsOptions struct {
	output string
}

func newPipelineLsCmd(out io.Writer) *cobra.Command {
	o := new(pipelineLsOptions)
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   pipelineLsDesc,
		Long:    pipelineLsDesc,
		Args:    NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *pipelineLsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/", global.Host, global.Owner)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{})

	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printPipelinesListAsTable,
			"json":  output.PrintJson,
		})
}

func printPipelinesListAsTable(raw string, out io.Writer, page int) error {
	var pipelines []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &pipelines)
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		_, err := out.Write([]byte("No pipelines were found\n"))
		if err != nil {
			return err
		}
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "VISIBILITY"}
	rows := []string{}
	for _, pipeline := range pipelines {
		row := fmt.Sprintf("%s\t%s\t%s", pipeline["name"], pipeline["description"], pipeline["visibility"])
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
