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

const listFlowsDesc = `List flows for an org.`

type listFlowsOptions struct {
	output string
}

func newListFlowsCmd(out io.Writer) *cobra.Command {
	o := new(listFlowsOptions)
	cmd := &cobra.Command{
		Use:   "flows",
		Short: listFlowsDesc,
		Long:  listFlowsDesc,
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
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

func (o *listFlowsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/flows/%s", global.Host, global.Org)

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
			"table": printFlowsListAsTable,
			"json":  output.PrintJson,
		})
}

func printFlowsListAsTable(raw string, out io.Writer, page int) error {
	var flows []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &flows)
	if err != nil {
		return err
	}

	if len(flows) == 0 {
		logger.Info("No flows were found.")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "VISIBILITY"}
	rows := []string{}
	for _, flow := range flows {
		row := fmt.Sprintf("%s\t%s\t%s", flow["name"], flow["description"], flow["visibility"])
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
