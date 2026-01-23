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

const listTrailsShortDesc = `List Trails of an org.`

const listTrailsLongDesc = listTrailsShortDesc + `The list can be filtered by flow and artifact fingerprint. The results are paginated and ordered from latest to oldest.`

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

# list all trails across all flows that contain an artifact with the provided fingerprint (in JSON):
kosli list trails \
	--fingerprint yourArtifactFingerprint \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json \
`

type listTrailsOptions struct {
	listOptions
	flowName    string
	fingerprint string
}

type Trail struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	ComplianceState string `json:"compliance_state"`
}

type Pagination struct {
	Page      float64 `json:"page"`
	PageCount float64 `json:"page_count"`
	Total     float64 `json:"total"`
}

type listTrailsResponse struct {
	Data       []Trail    `json:"data"`
	Pagination Pagination `json:"pagination"`
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

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlagOptional)
	cmd.Flags().StringVarP(&o.fingerprint, "fingerprint", "F", "", fingerprintInTrailsFlag)
	// We set the defauly page limit to 0 so that all results are returned if the flag is not provided
	addListFlags(cmd, &o.listOptions, 20)

	return cmd
}

func (o *listTrailsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/trails/%s?per_page=%d&page=%d", global.Host, global.Org, o.pageLimit, o.pageNumber)
	if o.flowName != "" {
		url += fmt.Sprintf("&flow=%s", o.flowName)
	}
	if o.fingerprint != "" {
		url += fmt.Sprintf("&fingerprint=%s", o.fingerprint)
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
	response := &listTrailsResponse{}
	trails := []Trail{}

	err := json.Unmarshal([]byte(raw), &trails)
	if err != nil {
		err = json.Unmarshal([]byte(raw), &response)
		if err != nil {
			return err
		}
		trails = response.Data
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
		row := fmt.Sprintf("%s\t%s\t%s", trail.Name, trail.Description, trail.ComplianceState)
		rows = append(rows, row)
	}
	if len(response.Data) > 0 {
		pagination := response.Pagination
		paginationInfo := fmt.Sprintf("\nShowing page %.0f of %.0f, total %.0f items", pagination.Page, pagination.PageCount, pagination.Total)
		rows = append(rows, paginationInfo)
	}

	tabFormattedPrint(out, header, rows)

	return nil
}
