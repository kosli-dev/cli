package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

const snapshotLsDesc = `List all snapshots for an environment.`

type snapshotLsOptions struct {
	json   bool
	number int64
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

	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)
	cmd.Flags().Int64VarP(&o.number, "number", "n", 5, resultLimitFlag)

	return cmd
}

func (o *snapshotLsOptions) run(out io.Writer, args []string) error {
	if o.number <= 0 {
		_, err := out.Write([]byte("No environment snapshots were requested\n"))
		if err != nil {
			return err
		}
		return nil
	}
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/log/0/%d",
		global.Host, global.Owner, args[0], o.number)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return err
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
	} else {
		var snapshots []map[string]interface{}
		err = json.Unmarshal([]byte(response.Body), &snapshots)
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
			tsFromStr := time.Unix(int64(snapshot["from"].(float64)), 0).Format(time.RFC3339)
			tsToStr := "now"
			if snapshot["to"].(float64) != 0.0 {
				tsToStr = time.Unix(int64(snapshot["to"].(float64)), 0).Format(time.RFC3339)
			}
			timeago.English.Max = 36 * timeago.Month
			timeago.English.PastSuffix = ""
			durationNs := time.Duration(int64(snapshot["duration"].(float64)) * 1e9)
			duration := timeago.English.FormatRelativeDuration(durationNs)
			index := int64(snapshot["index"].(float64))
			row := fmt.Sprintf("%d\t%s\t%s\t%s", index, tsFromStr, tsToStr, duration)
			rows = append(rows, row)
		}
		printTable(out, header, rows)
	}

	return nil
}
