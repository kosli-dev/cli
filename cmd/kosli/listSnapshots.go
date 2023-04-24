package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

const listSnapshotsShortDesc = `List environment snapshots.`

const listSnapshotsLongDesc = listSnapshotsShortDesc + `
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 snapshots per page.

You can optionally specify an INTERVAL between two snapshot expressions with [expression]..[expression]. 

Expressions can be:
* ~N   N'th behind the latest snapshot  
* N    snapshot number N  
* NOW  the latest snapshot  

Either expression can be omitted to default to NOW.
`

const listSnapshotsExample = `
# list the last 15 snapshots for an environment:
kosli list snapshots yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 snapshots for an environment:
kosli list snapshots yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 snapshots for an environment (in JSON):
kosli list snapshots yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listSnapshotsOptions struct {
	listOptions
	reverse  bool
	interval string
}

func newListSnapshotsCmd(out io.Writer) *cobra.Command {
	o := new(listSnapshotsOptions)
	cmd := &cobra.Command{
		Use:     "snapshots ENV_NAME",
		Short:   listSnapshotsShortDesc,
		Long:    listSnapshotsLongDesc,
		Example: listSnapshotsExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return o.validate(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.interval, "interval", "i", "", intervalFlag)
	addListFlags(cmd, &o.listOptions)
	cmd.Flags().BoolVar(&o.reverse, "reverse", false, reverseFlag)

	return cmd
}

func (o *listSnapshotsOptions) run(out io.Writer, args []string) error {
	envName := args[0]
	return o.getSnapshotsList(out, envName, o.interval)

}

func (o *listSnapshotsOptions) getSnapshotsList(out io.Writer, envName, interval string) error {
	url := fmt.Sprintf("%s/api/v2/snapshots/%s/%s?page=%d&per_page=%d&interval=%s&reverse=%t",
		global.Host, global.Org, envName, o.pageNumber, o.pageLimit, url.QueryEscape(interval), o.reverse)

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
			"table": printSnapshotsListAsTable,
			"json":  output.PrintJson,
		})
}

func printSnapshotsListAsTable(raw string, out io.Writer, page int) error {
	var snapshots []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &snapshots)
	if err != nil {
		return err
	}

	if len(snapshots) == 0 {
		msg := "No environment snapshots were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"SNAPSHOT", "FROM", "TO", "DURATION", "COMPLIANT"}
	rows := []string{}
	for _, snapshot := range snapshots {
		tsFromStr, err := formattedTimestamp(snapshot["from"], true)
		if err != nil {
			return err
		}
		tsToStr := "now"
		if snapshot["to"].(float64) != 0.0 {
			tsToStr, err = formattedTimestamp(snapshot["to"], true)
			if err != nil {
				return err
			}
		}

		timeago.English.Max = 36 * timeago.Month
		timeago.English.PastSuffix = ""
		durationNs := time.Duration(int64(snapshot["duration"].(float64)) * 1e9)
		duration := timeago.English.FormatRelativeDuration(durationNs)
		compliance := snapshot["compliant"].(bool)
		index := int64(snapshot["index"].(float64))
		row := fmt.Sprintf("%d\t%s\t%s\t%s\t%t", index, tsFromStr, tsToStr, duration, compliance)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}

func printEnvironmentEventsLogAsTable(raw string, out io.Writer, page int) error {
	var events []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &events)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		msg := "No environment events were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}
	header := []string{"SNAPSHOT", "EVENT", "FLOW", "DEPLOYMENTS"}
	rows := []string{}
	for _, event := range events {
		snapshotIndex := int(event["snapshot_index"].(float64))
		artifactName := event["artifact_name"].(string)
		fingerprint := event["sha256"].(string)
		description := event["description"].(string)
		reportedAt, err := formattedTimestamp(event["reported_at"], true)
		if err != nil {
			return err
		}
		flow := event["pipeline"].(string)
		deploymentsList := event["deployments"].([]interface{})
		deployments := ""
		for _, deployment := range deploymentsList {
			deployments += fmt.Sprintf("#%d ", int64(deployment.(float64)))
		}

		row := fmt.Sprintf("#%d\tArtifact: %s\t%s\t%s", snapshotIndex, artifactName, flow, deployments)
		rows = append(rows, row)
		row = fmt.Sprintf("\tFingerprint: %s\t\t", fingerprint)
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
