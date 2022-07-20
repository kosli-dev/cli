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

const environmentLogDesc = `Show log of snapshots.`

type environmentLogOptions struct {
	// long bool
	json   bool
	number int64
}

// type EnvironmentDiffPayload struct {
// 	Snappish1 string `json:"snappish1"`
// 	Snappish2 string `json:"snappish2"`
// }

func newEnvironmentLogCmd(out io.Writer) *cobra.Command {
	o := new(environmentLogOptions)
	cmd := &cobra.Command{
		Use:   "log ENVIRONMENT-NAME",
		Short: environmentLogDesc,
		Long:  environmentLogDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	// cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)
	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)
	cmd.Flags().Int64VarP(&o.number, "number", "n", 5, environmentJsonFlag)

	return cmd
}

func (o *environmentLogOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/log/0/%d",
		global.Host, global.Owner, args[0], o.number)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return fmt.Errorf("kosli server %s is unresponsive %s", global.Host, err)
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

		fmt.Printf("%-8s  %-25s  %-25s  %s", "SNAPSHOT", "FROM", "TO", "DURATION\n")
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
			fmt.Printf("%-8d  %s  %-25s  %s\n", index, tsFromStr, tsToStr, duration)
		}

	}

	return nil
}
