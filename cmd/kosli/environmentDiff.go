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

const environmentDiffDesc = `Diff snapshots.`

type environmentDiffOptions struct {
	output string
}

type EnvironmentDiffResponse struct {
	Sha256              string   `json:"sha256"`
	Pipeline            string   `json:"pipeline"`
	Name                string   `json:"name"`
	CommitUrl           string   `json:"commit_url"`
	MostRecentTimestamp int64    `json:"most_recent_timestamp"`
	InstancesS1         int64    `json:"instances_s1"`
	InstancesS2         int64    `json:"instances_s2"`
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
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 2 {
				return ErrorBeforePrintingUsage(cmd, "two snappishes required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *environmentDiffOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/env-diff/%s/?snappish1=%s&snappish2=%s",
		global.Host, global.Owner, url.QueryEscape(args[0]), url.QueryEscape(args[1]))

	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printEnvironmentDiffAsTable,
			"json":  output.PrintJson,
		})
}

func printEnvironmentDiffAsTable(raw string, out io.Writer, page int) error {
	var diffs map[string][]EnvironmentDiffResponse
	err := json.Unmarshal([]byte(raw), &diffs)
	if err != nil {
		return err
	}

	colorRed := "\033[31m%s\033[0m"
	colorGreen := "\033[32m%s\033[0m"
	noColor := "%s"

	removalCount := len(diffs["-"])
	additionCount := len(diffs["+"])
	changedCount := len(diffs["0"])

	if removalCount > 0 {
		for _, entry := range diffs["-"] {
			err := printDiffEntry(colorRed, "-", entry)
			if err != nil {
				return err
			}
		}
	}

	if removalCount > 0 && additionCount > 0 {
		fmt.Println()
	}

	if additionCount > 0 {
		for _, entry := range diffs["+"] {
			err := printDiffEntry(colorGreen, "+", entry)
			if err != nil {
				return err
			}
		}
	}

	if changedCount > 0 && (additionCount > 0 || removalCount > 0) {
		fmt.Println()
	}

	if changedCount > 0 {
		for _, entry := range diffs["0"] {
			err := printDiffEntry(noColor, " ", entry)
			if err != nil {
				return err
			}
			fmt.Printf("  Instances: ")
			fmt.Printf("scaled from %d to %d\n", entry.InstancesS1, entry.InstancesS2)
		}
	}

	return nil
}

func printDiffEntry(color string, sign string, entry EnvironmentDiffResponse) error {
	fmt.Printf(color, sign+" Name: ")
	fmt.Printf("  %s\n", entry.Name)
	fmt.Printf(color, "  Sha256: ")
	fmt.Printf("%s\n", entry.Sha256)
	if entry.Pipeline != "" {
		fmt.Printf(color, "  Pipeline: ")
		fmt.Printf("%s\n", entry.Pipeline)
	} else {
		fmt.Printf(color, "  Pipeline: ")
		fmt.Printf("Unknown\n")
	}
	if entry.CommitUrl != "" {
		fmt.Printf(color, "  Commit: ")
		fmt.Printf("%s\n", entry.CommitUrl)
	} else {
		fmt.Printf(color, "  Commit: ")
		fmt.Printf("Unknown\n")
	}
	if len(entry.Pods) > 0 {
		fmt.Printf(color, "  Pods: ")
		fmt.Printf("  %s\n", entry.Pods)
	}
	fmt.Printf(color, "  Started: ")
	timestamp, err := formattedTimestamp(entry.MostRecentTimestamp, false)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", timestamp)
	return nil
}
