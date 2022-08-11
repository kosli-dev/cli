package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

const snapshotLsDesc = `List all snapshots for an environment.`

type snapshotLsOptions struct {
	output     string
	pageNumber int
	pageLimit  int
}

func newSnapshotLsCmd(out io.Writer) *cobra.Command {
	o := new(snapshotLsOptions)
	cmd := &cobra.Command{
		Use:     "ls ENVIRONMENT-NAME",
		Aliases: []string{"list"},
		Short:   snapshotLsDesc,
		Long:    snapshotLsDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "environment name argument is required")
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

func (o *snapshotLsOptions) run(out io.Writer, args []string) error {
	if o.pageNumber <= 0 || o.pageLimit <= 0 {
		fmt.Fprint(out, "No environment snapshots were requested\n")
		return nil
	}
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/snapshots/%d/%d",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit)
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodGet, log)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
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
		_, err := out.Write([]byte("No environment snapshots were found\n"))
		if err != nil {
			return err
		}
		return nil
	}

	header := []string{"SNAPSHOT", "FROM", "TO", "DURATION"}
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
		index := int64(snapshot["index"].(float64))
		row := fmt.Sprintf("%d\t%s\t%s\t%s", index, tsFromStr, tsToStr, duration)
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
