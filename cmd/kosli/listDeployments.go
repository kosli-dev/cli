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

const listDeploymentsShortDesc = `List deployments in a flow.`

const listDeploymentsLongDesc = listDeploymentsShortDesc + `
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 deployments per page.
`
const listDeploymentsExample = `
# list the last 15 deployments for a flow:
kosli list deployments \ 
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 deployments for a flow:
kosli list deployments \ 
	--flow yourFlowName \	
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 deployments for a flow (in JSON):
kosli list deployments \ 
	--flow yourFlowName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listDeploymentsOptions struct {
	listOptions
	flowName string
}

func newListDeploymentsCmd(out io.Writer) *cobra.Command {
	o := new(listDeploymentsOptions)
	cmd := &cobra.Command{
		Use:     "deployments",
		Aliases: []string{"deployment", "deploy"},
		Short:   listDeploymentsShortDesc,
		Long:    listDeploymentsLongDesc,
		Example: listDeploymentsExample,
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

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	addListFlags(cmd, &o.listOptions)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *listDeploymentsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/deployments/%s/%s?page=%d&per_page=%d",
		global.Host, global.Org, o.flowName, o.pageNumber, o.pageLimit)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printDeploymentsListAsTable,
			"json":  output.PrintJson,
		})
}

func printDeploymentsListAsTable(raw string, out io.Writer, page int) error {

	var deployments []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &deployments)
	if err != nil {
		return err
	}

	if len(deployments) == 0 {
		msg := "No deployments were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"ID", "ARTIFACT", "ENVIRONMENT", "REPORTED_AT"}
	rows := []string{}
	for _, deployment := range deployments {
		deploymentId := int(deployment["deployment_id"].(float64))
		artifactName := deployment["artifact_name"].(string)
		artifactDigest := deployment["artifact_fingerprint"].(string)
		environment := deployment["environment"].(string)
		createdAt, err := formattedTimestamp(deployment["created_at"], true)
		if err != nil {
			return err
		}
		row := fmt.Sprintf("%d\tName: %s\t%s\t%s", deploymentId, artifactName, environment, createdAt)
		rows = append(rows, row)
		row = fmt.Sprintf("\tFingerprint: %s\t\t", artifactDigest)
		rows = append(rows, row)
		rows = append(rows, "\t\t\t")
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
