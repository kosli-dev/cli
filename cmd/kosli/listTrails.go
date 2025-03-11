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

const listTrailsShortDesc = `List Trails for a Flow in an org.`

const listTrailsLongDesc = listTrailsShortDesc + `The results are ordered from latest to oldest.  
If the ^page-limit^ flag is provided, the results will be paginated, otherwise all results will be 
returned.  
If ^page-limit^ is set to 0, all results will be returned.`

const listTrailsExample = `
# list all trails for a flow:
kosli list trails \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

#list the most recent 30 trails for a flow:
kosli list trails \
	--flow yourFlowName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

#show the second page of trails for a flow:
kosli list trails \
	--flow yourFlowName \
	--page-limit 30 \
	--page 2 \
	--api-token yourAPIToken \
	--org yourOrgName

# list all trails for a flow (in JSON):
kosli list trails \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listTrailsOptions struct {
	listOptions
	flowName string
}

func newListTrailsCmd(out io.Writer) *cobra.Command {
	o := new(listTrailsOptions)
	cmd := &cobra.Command{
		Use:     "trails",
		Short:   listTrailsShortDesc,
		Long:    listTrailsLongDesc,
		Example: listTrailsExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return o.validateForListTrails(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().IntVar(&o.pageNumber, "page", 1, pageNumberFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", 0, pageLimitListTrailsFlag)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *listTrailsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/trails/%s/%s", global.Host, global.Org, o.flowName)

	// For backward compatibility, we need to return all of the results if no pagination
	// flags are provided - i.e. by not passing a pageLimit parameter to the API.
	if o.pageLimit != 0 {
		url = fmt.Sprintf("%s?per_page=%d&page=%d", url, o.pageLimit, o.pageNumber)
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

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printTrailsListAsTable,
			"json":  output.PrintJson,
		})
}

func printTrailsListAsTable(raw string, out io.Writer, page int) error {
	var trails []map[string]interface{}
	var response map[string]interface{}

	err := json.Unmarshal([]byte(raw), &trails)
	if err != nil {
		err = json.Unmarshal([]byte(raw), &response)
		if err != nil {
			return err
		}
		// This is a little ridiculous but seems to be the easiest way to get the
		// trails list from paginated results and put it into the trails variable defined above.
		trails_json, err := json.Marshal(response["data"])
		if err != nil {
			return err
		}
		err = json.Unmarshal(trails_json, &trails)
		if err != nil {
			return err
		}
	}

	if len(trails) == 0 {
		msg := "No trails were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "COMPLIANCE"}
	rows := []string{}
	for _, trail := range trails {
		row := fmt.Sprintf("%s\t%s\t%s", trail["name"], trail["description"], trail["compliance_state"])
		rows = append(rows, row)
	}
	if pagination, ok := response["pagination"].(map[string]interface{}); ok {
		paginationInfo := fmt.Sprintf("\nShowing page %.0f of %.0f, total %.0f items", pagination["page"], pagination["page_count"], pagination["total"])
		rows = append(rows, paginationInfo)
	}

	tabFormattedPrint(out, header, rows)

	return nil
}
