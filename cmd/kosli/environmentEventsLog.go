package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const environmentEventsLogDesc = `List a number of environment events.`

type environmentEventsLogOptions struct {
	output     string
	pageNumber int
	pageLimit  int
}

func newEnvironmentEventsLogCmd(out io.Writer) *cobra.Command {
	o := new(environmentEventsLogOptions)
	cmd := &cobra.Command{
		Use:   "log SNAPPISH_1 [SNAPPISH_2]",
		Short: environmentEventsLogDesc,
		Long:  environmentEventsLogDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "SNAPPISH_1 argument is required")
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

	return cmd
}

func (o *environmentEventsLogOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/log/%d/%d?from=%s&to=%s",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit, url.QueryEscape(args[0]), url.QueryEscape(args[1]))
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
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
		deploymentsList := event["deployments"].([]int64)
		deployments := ""
		for _, deployment := range deploymentsList {
			deployments += fmt.Sprintf("#%d ", deployment)
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
