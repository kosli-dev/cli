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

const listTrailsDesc = `List Trails for a Flow in an org.`

type listTrailsOptions struct {
	flowName string
	output   string
}

func newListTrailsCmd(out io.Writer) *cobra.Command {
	o := new(listTrailsOptions)
	cmd := &cobra.Command{
		Use:   "trails",
		Short: listTrailsDesc,
		Long:  listTrailsDesc,
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

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *listTrailsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/trails/%s/%s", global.Host, global.Org, o.flowName)

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
			"table": printTrailsListAsTable,
			"json":  output.PrintJson,
		})
}

func printTrailsListAsTable(raw string, out io.Writer, page int) error {
	var trails []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &trails)
	if err != nil {
		return err
	}

	if len(trails) == 0 {
		logger.Info("No trails were found.")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "COMPLIANCE"}
	rows := []string{}
	for _, trail := range trails {
		row := fmt.Sprintf("%s\t%s\t%s", trail["name"], trail["description"], trail["compliance_state"])
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
