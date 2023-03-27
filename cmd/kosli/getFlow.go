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

const getFlowDesc = `Get the metadata of a specific flow.`

type getFlowOptions struct {
	output string
}

func newGetFlowCmd(out io.Writer) *cobra.Command {
	o := new(getFlowOptions)
	cmd := &cobra.Command{
		Use:   "flow FLOW-NAME",
		Short: getFlowDesc,
		Long:  getFlowDesc,
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

func (o *getFlowOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v2/flows/%s/%s", global.Host, global.Org, args[0])

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
			"table": printFlowAsTable,
			"json":  output.PrintJson,
		})
}

func printFlowAsTable(raw string, out io.Writer, page int) error {
	var flow map[string]interface{}
	err := json.Unmarshal([]byte(raw), &flow)
	if err != nil {
		return err
	}

	header := []string{}
	rows := []string{}

	lastDeployedAt, err := formattedTimestamp(flow["last_deployment_at"], false)
	if err != nil {
		return err
	}
	template := fmt.Sprintf("%s", flow["template"])
	template = strings.Replace(template, " ", ", ", -1)

	rows = append(rows, fmt.Sprintf("Name:\t%s", flow["name"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", flow["description"]))
	rows = append(rows, fmt.Sprintf("Visibility:\t%s", flow["visibility"]))
	rows = append(rows, fmt.Sprintf("Template:\t%s", template))
	rows = append(rows, fmt.Sprintf("Last Deployment At:\t%s", lastDeployedAt))

	tabFormattedPrint(out, header, rows)
	return nil
}
