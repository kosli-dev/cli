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
	Snappish1 DiffItem `json:"snappish1"`
	Snappish2 DiffItem `json:"snappish2"`
	Changed   DiffItem `json:"changed"`
}

type DiffItem struct {
	SnapshotID string         `json:"snapshot_id"`
	Artifacts  []DiffArtifact `json:"artifacts"`
}

type DiffArtifact struct {
	Sha256              string   `json:"sha256"`
	Pipeline            string   `json:"pipeline"`
	Name                string   `json:"name"`
	CommitUrl           string   `json:"commit_url"`
	MostRecentTimestamp int64    `json:"most_recent_timestamp"`
	S1InstanceCount     int64    `json:"s1_instance_count"`
	S2InstanceCount     int64    `json:"s2_instance_count"`
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
	var diffs EnvironmentDiffResponse
	err := json.Unmarshal([]byte(raw), &diffs)
	if err != nil {
		return err
	}

	s1Artifacts := diffs.Snappish1.Artifacts
	s2Artifacts := diffs.Snappish2.Artifacts
	changedArtifacts := diffs.Changed.Artifacts
	s1Count := len(s1Artifacts)
	s2Count := len(s2Artifacts)
	changedCount := len(changedArtifacts)

	if s1Count > 0 {
		if snappish1 == diffs.Snappish1.SnapshotID {
			fmt.Printf("Only present in %s\n", snappish1)
		} else {
			fmt.Printf("Only present in %s (snapshot: %s)\n", snappish1, diffs.Snappish1.SnapshotID)
		}
		for _, entry := range s1Artifacts {
			err := printOnlyEntry(entry, out)
			if err != nil {
				return err
			}
		}
	}

	if s1Count > 0 && s2Count > 0 {
		fmt.Println()
	}

	if s2Count > 0 {
		if snappish2 == diffs.Snappish2.SnapshotID {
			fmt.Printf("Only present in %s\n", snappish2)
		} else {
			fmt.Printf("Only present in %s (snapshot: %s)\n", snappish2, diffs.Snappish2.SnapshotID)
		}
		for _, entry := range s2Artifacts {
			err := printOnlyEntry(entry, out)
			if err != nil {
				return err
			}
		}
	}

	if changedCount > 0 && (s1Count > 0 || s2Count > 0) {
		fmt.Println()
	}

	if changedCount > 0 {
		fmt.Printf("%s -> %s scaling\n", snappish1, snappish2)
		for _, entry := range changedArtifacts {
			err := printOnlyEntry(entry, out)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func printOnlyEntry(entry DiffArtifact, out io.Writer) error {
	rows := []string{}
	rows = append(rows, "\t\t")
	rows = append(rows, fmt.Sprintf("\tName:\t%s", entry.Name))
	rows = append(rows, fmt.Sprintf("\tFingerprint:\t%s", entry.Sha256))

	if entry.Pipeline != "" {
		rows = append(rows, fmt.Sprintf("\tPipeline:\t%s", entry.Pipeline))
	} else {
		rows = append(rows, "\tPipeline:\tUnknown")
	}

	if entry.CommitUrl != "" {
		rows = append(rows, fmt.Sprintf("\tCommit URL:\t%s", entry.CommitUrl))
	} else {
		rows = append(rows, "\tCommit URL:\tUnknown")
	}

	if len(entry.Pods) > 0 {
		rows = append(rows, fmt.Sprintf("\tPods:\t%s", entry.Pods))
	}

	timestamp, err := formattedTimestamp(entry.MostRecentTimestamp, false)
	if err != nil {
		return err
	}
	rows = append(rows, fmt.Sprintf("\tStarted:\t%s", timestamp))

	if entry.S1InstanceCount != 0 && entry.S2InstanceCount != 0 {
		rows = append(rows, fmt.Sprintf("\tInstances:\tscaled from %d to %d", entry.S1InstanceCount, entry.S2InstanceCount))
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}
