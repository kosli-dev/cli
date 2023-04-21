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

const listArtifactsShortDesc = `List artifacts in a flow. `

const listArtifactsLongDesc = listArtifactsShortDesc + `The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 artifacts per page.
`
const artifactLsExample = `
# list the last 15 artifacts for a flow:
kosli list artifacts \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 artifacts for a flow:
kosli list artifacts \
	--flow yourFlowName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 artifacts for a flow (in JSON):
kosli list artifacts \
	--flow yourFlowName \	
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listArtifactsOptions struct {
	listOptions
	flowName string
}

func newListArtifactsCmd(out io.Writer) *cobra.Command {
	o := new(listArtifactsOptions)
	cmd := &cobra.Command{
		Use:     "artifacts",
		Short:   listArtifactsShortDesc,
		Long:    listArtifactsLongDesc,
		Example: artifactLsExample,
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

func (o *listArtifactsOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/artifacts/%s/%s?page=%d&per_page=%d",
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
			"table": printArtifactsListAsTable,
			"json":  output.PrintJson,
		})
}

func printArtifactsListAsTable(raw string, out io.Writer, page int) error {
	var artifacts []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &artifacts)
	if err != nil {
		return err
	}

	if len(artifacts) == 0 {
		msg := "No artifacts were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"COMMIT", "ARTIFACT", "STATE", "CREATED_AT"}
	rows := []string{}
	for _, artifact := range artifacts {

		gitCommit := artifact["git_commit"].(string)[:7]
		artifactName := artifact["filename"].(string)

		artifactDigest := artifact["fingerprint"].(string)
		artifactState := artifact["state"].(string)
		createdAt, err := formattedTimestamp(artifact["created_at"], true)
		if err != nil {
			return err
		}

		row := fmt.Sprintf("%s\tName: %s\t%s\t%s", gitCommit, artifactName, artifactState, createdAt)
		rows = append(rows, row)
		row = fmt.Sprintf("\tFingerprint: %s\t\t", artifactDigest)
		rows = append(rows, row)
		rows = append(rows, "\t\t\t")

	}
	tabFormattedPrint(out, header, rows)

	return nil
}
