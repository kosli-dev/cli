package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const environmentEventsLogShortDesc = `List environment events. `

const environmentEventsLogLongDesc = environmentEventsLogShortDesc + `The results are paginated and ordered from latests to oldest. 
By default, the page limit is 15 events per page.
`

const environmentEventsLogExample = `
# list the last 15 events for an environment:
kosli environment log yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 events for an environment:
kosli environment log yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName

# list the last 30 events for an environment (in JSON):
kosli environment log yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--owner yourOrgName \
	--output json
`

type environmentEventsLogOptions struct {
	output     string
	long       bool
	pageNumber int
	pageLimit  int
	reverse    bool
}

func newEnvironmentEventsLogCmd(out io.Writer) *cobra.Command {
	o := new(environmentEventsLogOptions)
	cmd := &cobra.Command{
		Use:     "log ENV_NAME [INTERVAL]",
		Short:   environmentEventsLogShortDesc,
		Long:    environmentEventsLogLongDesc,
		Example: environmentEventsLogExample,
		Args:    cobra.MatchAll(cobra.MaximumNArgs(2), cobra.MinimumNArgs(1)),
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
	cmd.Flags().BoolVarP(&o.long, "long", "l", false, longFlag)
	cmd.Flags().IntVar(&o.pageNumber, "page", 1, pageNumberFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", 15, pageLimitFlag)
	cmd.Flags().BoolVar(&o.reverse, "reverse", false, reverseFlag)

	return cmd
}

func (o *environmentEventsLogOptions) run(out io.Writer, args []string) error {
	interval := ""
	if len(args) > 1 {
		interval = args[1]
	}

	if o.long {
		url := fmt.Sprintf("%s/api/v1/environments/%s/%s/events/?page=%d&per_page=%d&interval=%s&reverse=%t",
			global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit, url.QueryEscape(interval), o.reverse)

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
				"table": printEnvironmentEventsLogAsTable,
				"json":  output.PrintJson,
			})
	} else {
		err := o.getSnapshotsList(out, args)
		if err != nil {
			return err
		}
		return nil
	}
}

func printEnvironmentEventsLogAsTable(raw string, out io.Writer, page int) error {

	var events []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &events)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		logger.Info("no environment events were found at page number %d", page)
		return nil
	}
	header := []string{"SNAPSHOT", "EVENT", "PIPELINE", "DEPLOYMENTS"}
	rows := []string{}
	for _, event := range events {
		snapshotIndex := int(event["snapshot_index"].(float64))
		artifactName := event["artifact_name"].(string)
		sha256 := event["sha256"].(string)
		description := event["description"].(string)
		reportedAt, err := formattedTimestamp(event["reported_at"], true)
		if err != nil {
			return err
		}
		pipeline := event["pipeline"].(string)
		deploymentsList := event["deployments"].([]interface{})
		deployments := ""
		for _, deployment := range deploymentsList {
			deployments += fmt.Sprintf("#%d ", int64(deployment.(float64)))
		}

		row := fmt.Sprintf("#%d\tArtifact: %s\t%s\t%s", snapshotIndex, artifactName, pipeline, deployments)
		rows = append(rows, row)
		row = fmt.Sprintf("\tFingerprint: %s\t\t", sha256)
		rows = append(rows, row)
		row = fmt.Sprintf("\tDescription: %s\t\t", description)
		rows = append(rows, row)
		row = fmt.Sprintf("\tReported at: %s\t\t", reportedAt)
		rows = append(rows, row)
		rows = append(rows, "\t\t\t") // These tabs are required for alignment
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
