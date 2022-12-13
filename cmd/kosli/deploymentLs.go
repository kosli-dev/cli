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

const deploymentLsShortDesc = `List deployments in a pipeline.`

const deploymentLsLongDesc = deploymentLsShortDesc + `
The results are paginated and ordered from latests to oldest. 
By default, the page limit is 15 deployments per page.
`
const deploymentLsExample = `
# list the last 15 deployments for a pipeline:
kosli deployment list yourPipelineName \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 deployments for a pipeline:
kosli deployment list yourPipelineName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 deployments for a pipeline (in JSON):
kosli deployment list yourPipelineName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--output json
`

type deploymentLsOptions struct {
	output     string
	pageNumber int
	pageLimit  int
}

func newDeploymentLsCmd(out io.Writer) *cobra.Command {
	o := new(deploymentLsOptions)
	cmd := &cobra.Command{
		Use:     "ls PIPELINE-NAME",
		Aliases: []string{"list"},
		Short:   deploymentLsShortDesc,
		Long:    deploymentLsLongDesc,
		Example: deploymentLsExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if o.pageNumber <= 0 {
				return ErrorBeforePrintingUsage(cmd, "page number must be a positive integer")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().IntVar(&o.pageNumber, "page", 1, pageNumberFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", 15, pageLimitFlag)

	return cmd
}

func (o *deploymentLsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/deployments/?page=%d&per_page=%d",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit)

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
		artifactDigest := deployment["artifact_sha256"].(string)
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
