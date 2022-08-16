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

const environmentEventsLogDesc = `List a number of environment events.`

type environmentEventsLogOptions struct {
	output     string
	pageNumber int
	pageLimit  int
	reverse    bool
}

func newEnvironmentEventsLogCmd(out io.Writer) *cobra.Command {
	o := new(environmentEventsLogOptions)
	cmd := &cobra.Command{
		Use:   "log ENV_NAME [INTERVAL]",
		Short: environmentEventsLogDesc,
		Long:  environmentEventsLogDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "ENV_NAME argument is required")
			}
			if len(args) > 2 {
				return ErrorBeforePrintingUsage(cmd, "command accepts maximum 2 arguments")
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
	cmd.Flags().IntVarP(&o.pageNumber, "page-number", "n", 1, pageNumberFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "l", 15, pageLimitFlag)
	cmd.Flags().BoolVar(&o.reverse, "reverse", false, reverseFlag)

	return cmd
}

func (o *environmentEventsLogOptions) run(out io.Writer, args []string) error {
	interval := ""
	if len(args) > 1 {
		interval = args[1]
	}
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/events/?page=%d&per_page=%d&interval=%s&reverse=%t",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit, url.QueryEscape(interval), o.reverse)
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodGet, log)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printEnvironmentEventsLogAsTable,
			"json":  output.PrintJson,
		})
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
		_, err := out.Write([]byte(msg + ".\n"))
		if err != nil {
			return err
		}
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
		row = fmt.Sprintf("\tSHA256: %s\t\t", sha256)
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
