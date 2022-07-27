package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

const environmentDiffDesc = `Diff snapshots.`

type environmentDiffOptions struct {
	// long bool
	json bool
}

type EnvironmentDiffPayload struct {
	Snappish1 string `json:"snappish1"`
	Snappish2 string `json:"snappish2"`
}

type EnvironmentDiffResponse struct {
	Sha256              string   `json:"sha256"`
	Name                string   `json:"name"`
	CommitUrl           string   `json:"commit_url"`
	MostRecentTimestamp int64    `json:"most_recent_timestamp"`
	Pods                []string `json:"pods"`
}

func newEnvironmentDiffCmd(out io.Writer) *cobra.Command {
	o := new(environmentDiffOptions)
	cmd := &cobra.Command{
		Use:   "diff SNAPPISH_1 SNAPPISH_2",
		Short: environmentDiffDesc,
		Long:  environmentDiffDesc,
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

	return cmd
}

func (o *environmentDiffOptions) run(out io.Writer, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("two snappish required")
	}

	payload := new(EnvironmentDiffPayload)
	payload.Snappish1 = args[0]
	payload.Snappish2 = args[1]

	url := fmt.Sprintf("%s/api/v1/env-diff/%s/", global.Host, global.Owner)
	// response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
	// 	global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	response, err := requests.SendPayload(payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodGet, log)

	if err != nil {
		return err
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
		return nil
	}

	var diffs map[string][]EnvironmentDiffResponse
	err = json.Unmarshal([]byte(response.Body), &diffs)
	if err != nil {
		return err
	}

	removalCount := len(diffs["-"])
	additionCount := len(diffs["+"])

	if removalCount > 0 {
		for _, entry := range diffs["-"] {
			printDiffEntry(true, entry)
		}
	}

	if removalCount > 0 && additionCount > 0 {
		fmt.Println()
	}

	if additionCount > 0 {
		for _, entry := range diffs["+"] {
			printDiffEntry(false, entry)
		}
	}
	return nil
}

func printDiffEntry(removal bool, entry EnvironmentDiffResponse) {
	colorRed := "\033[31m%s\033[0m"
	colorGreen := "\033[32m%s\033[0m"
	color := colorGreen
	sign := "+"
	if removal {
		color = colorRed
		sign = "-"
	}

	fmt.Printf(color, sign+" Name: ")
	fmt.Printf("  %s\n", entry.Name)
	fmt.Printf(color, "  Sha256: ")
	fmt.Printf("%s\n", entry.Sha256)
	if entry.CommitUrl != "" {
		fmt.Printf(color, "  Commit: ")
		fmt.Printf("%s\n", entry.CommitUrl)
	}
	if len(entry.Pods) > 0 {
		fmt.Printf(color, "  Pods: ")
		fmt.Printf("  %s\n", entry.Pods)
	}
	fmt.Printf(color, "  Started: ")
	timestamp := time.Unix(entry.MostRecentTimestamp, 0)
	timeago.English.Max = 36 * timeago.Month
	since := timeago.English.Format(timestamp)
	fmt.Printf("%s \u2022 %s\n", time.Unix(entry.MostRecentTimestamp, 0).Format(time.RFC822), since)
}
