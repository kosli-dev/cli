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
	"github.com/xeonx/timeago"
)

func (o *environmentEventsLogOptions) getSnapshotsList(out io.Writer, args []string) error {
	if o.pageNumber <= 0 || o.pageLimit <= 0 {
		fmt.Fprint(out, "No environment snapshots were requested\n")
		return nil
	}

	interval := ""
	if len(args) > 1 {
		interval = args[1]
	}

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/snapshots/?page=%d&per_page=%d&interval=%s&reverse=%t",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit, url.QueryEscape(interval), o.reverse)
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet)
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
		fmt.Fprintf(out, "No environment snapshots were found at page %d\n", page)
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
