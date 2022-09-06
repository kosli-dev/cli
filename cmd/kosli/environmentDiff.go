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
	snappish1 := args[0]
	snappish2 := args[1]
	url := fmt.Sprintf("%s/api/v1/env-diff/%s/?snappish1=%s&snappish2=%s",
		global.Host, global.Owner, url.QueryEscape(snappish1), url.QueryEscape(snappish2))

	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return err
	}

	wrapper := func(raw string, out io.Writer, page int) error {
		return printEnvironmentDiffAsTable(snappish1, snappish2, raw, out, page)
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": wrapper,
			"json":  output.PrintJson,
		})
}

func printEnvironmentDiffAsTable(snappish1, snappish2, raw string, out io.Writer, page int) error {
	var diffs map[string][]EnvironmentDiffResponse
	err := json.Unmarshal([]byte(raw), &diffs)
	if err != nil {
		return err
	}

	s1Count := len(diffs[snappish1])
	s2Count := len(diffs[snappish2])

	if s1Count > 0 {
		fmt.Printf("%s only\n", snappish1)
		for _, entry := range diffs[snappish1] {
			err := printOnlyEntry(entry)
			if err != nil {
				return err
			}
		}
	}

	if s1Count > 0 && s2Count > 0 {
		fmt.Println()
	}

	if s2Count > 0 {
		fmt.Printf("%s only\n", snappish2)
		for _, entry := range diffs[snappish2] {
			err := printOnlyEntry(entry)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func printOnlyEntry(entry EnvironmentDiffResponse) error {
	fmt.Printf("  Name: %s\n", entry.Name)

	fmt.Printf("  Sha256: %s\n", entry.Sha256)

	if entry.Pipeline != "" {
		fmt.Printf("  Pipeline: %s\n", entry.Pipeline)
	} else {
		fmt.Printf("  Pipeline: Unknown\n")
	}

	if entry.CommitUrl != "" {
		fmt.Printf("  Commit: %s\n", entry.CommitUrl)
	} else {
		fmt.Printf("  Commit: Unknown\n")
	}

	if len(entry.Pods) > 0 {
		fmt.Printf("  Pods: %s\n", entry.Pods)
	}

	timestamp, err := formattedTimestamp(entry.MostRecentTimestamp, false)
	if err != nil {
		return err
	}
	fmt.Printf("  Started: %s\n", timestamp)

	return nil
}
