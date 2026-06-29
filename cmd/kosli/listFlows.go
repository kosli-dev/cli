package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listFlowsShortDesc = `List flows for an org. `

const listFlowsLongDesc = listFlowsShortDesc + `The results are paginated and ordered from latest to oldest.
By default, the page limit is 20 flows per page.
The list can be filtered by name with --name (and --ignore-case for case-insensitive matching).`

const listFlowsExample = `
# list the first page of flows (20 per page):
kosli list flows \
	--api-token yourAPIToken \
	--org yourOrgName

# list the first 30 flows:
kosli list flows \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# show the second page of flows:
kosli list flows \
	--page-limit 30 \
	--page 2 \
	--api-token yourAPIToken \
	--org yourOrgName

# list flows whose name contains "backend" (in JSON):
kosli list flows \
	--name backend \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listFlowsOptions struct {
	listOptions
	name       string
	ignoreCase bool
}

func newListFlowsCmd(out io.Writer) *cobra.Command {
	o := new(listFlowsOptions)
	cmd := &cobra.Command{
		Use:     "flows",
		Short:   listFlowsShortDesc,
		Long:    listFlowsLongDesc,
		Example: listFlowsExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return o.validate(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	addListFlags(cmd, &o.listOptions, 20)
	cmd.Flags().StringVarP(&o.name, "name", "N", "", searchByNameFlag)
	cmd.Flags().BoolVarP(&o.ignoreCase, "ignore-case", "i", false, ignoreCaseFlag)

	return cmd
}

func (o *listFlowsOptions) run(out io.Writer) error {
	base, err := url.JoinPath(global.Host, "api/v2/flows", global.Org)
	if err != nil {
		return err
	}

	params := url.Values{}
	// sending per_page switches the endpoint to the paginated envelope response
	params.Set("page", strconv.Itoa(o.pageNumber))
	params.Set("per_page", strconv.Itoa(o.pageLimit))
	if o.name != "" {
		params.Set("search_by_name", o.name)
		// case_sensitive only affects search, so only send it alongside a search term
		if o.ignoreCase {
			params.Set("case_sensitive", "false")
		}
	}
	reqURL := base + "?" + params.Encode()

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printFlowsListAsTable,
			"json":  output.PrintJson,
		})
}

type listFlowsResponse struct {
	Data       []map[string]interface{} `json:"data"`
	Pagination Pagination               `json:"pagination"`
}

func printFlowsListAsTable(raw string, out io.Writer, page int) error {
	response := &listFlowsResponse{}
	err := json.Unmarshal([]byte(raw), response)
	if err != nil {
		return err
	}
	flows := response.Data

	if len(flows) == 0 {
		msg := "No flows were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"NAME", "DESCRIPTION", "VISIBILITY", "TAGS"}
	rows := []string{}
	for _, flow := range flows {
		tags := flow["tags"].(map[string]interface{})
		tagsOutput := ""
		for key, value := range tags {
			tagsOutput += fmt.Sprintf("[%s=%s], ", key, value)
		}
		tagsOutput = strings.TrimSuffix(tagsOutput, ", ")
		row := fmt.Sprintf("%s\t%s\t%s\t%s", flow["name"], flow["description"], flow["visibility"], tagsOutput)
		rows = append(rows, row)
	}
	pagination := response.Pagination
	paginationInfo := fmt.Sprintf("\nShowing page %.0f of %.0f, total %.0f items", pagination.Page, pagination.PageCount, pagination.Total)
	rows = append(rows, paginationInfo)

	tabFormattedPrint(out, header, rows)

	return nil
}
