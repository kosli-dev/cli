package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	output string
}

type SearchResponse struct {
	ResolvedTo              ResolvedToBody           `json:"resolved_to"`
	ArtifactsForCommit      []map[string]interface{} `json:"artifacts_for_commit"`
	ArtifactsForFingerprint []map[string]interface{} `json:"artifacts_for_fingerprint"`
	EnvironmentEvents       []map[string]interface{} `json:"environment_events_for_no_provenance_artifacts"`
	Allowlist               []map[string]interface{} `json:"allowlist"`
}

type ResolvedToBody struct {
	FullMatch string `json:"full_match"`
	Type      string `json:"type"`
}

type HistoryEvent struct {
	Description string
	Timestamp   float64
}

// const artifactCreationExample = `
// # Report to a Kosli pipeline that a file type artifact has been created
// kosli pipeline artifact report creation FILE.tgz \
// 	--api-token yourApiToken \
// 	--artifact-type file \
// 	--build-url https://exampleci.com \
// 	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--owner yourOrgName \
// 	--pipeline yourPipelineName

// # Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
// kosli pipeline artifact report creation \
// 	--api-token yourApiToken \
// 	--build-url https://exampleci.com \
// 	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
// 	--owner yourOrgName \
// 	--pipeline yourPipelineName \
// 	--sha256 yourSha256
// `

func newSearchCmd(out io.Writer) *cobra.Command {
	o := new(searchOptions)
	cmd := &cobra.Command{
		Use:   "search GIT-COMMIT|FINGERPRINT",
		Short: "Search for a git commit or artifact fingerprint in Kosli.",
		// Example: artifactCreationExample,
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "git commit or artifact fingerprint argument is required")
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

func (o *searchOptions) run(out io.Writer, args []string) error {
	var err error
	search_value := args[0]

	url := fmt.Sprintf("%s/api/v1/search/%s/sha/%s", global.Host, global.Owner, search_value)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, log)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printSearchAsTableWrapper,
			"json":  output.PrintJson,
		})
}

func printSearchAsTableWrapper(responseRaw string, out io.Writer, pageNumber int) error {
	var searchResult SearchResponse
	err := json.Unmarshal([]byte(responseRaw), &searchResult)
	if err != nil {
		return err
	}
	fullMatch := searchResult.ResolvedTo.FullMatch
	if searchResult.ResolvedTo.Type == "git_commit" {
		fmt.Fprintf(out, "Search result resolved to commit %s\n", fullMatch)
	} else {
		fmt.Fprintf(out, "Search result resolved to artifact with fingerprint %s\n", fullMatch)
	}
	if len(searchResult.ArtifactsForCommit) > 0 {
		fmt.Fprintf(out, "Found the following artifact(s) for commit\n")
		err = printArtifactsJsonAsTable(searchResult.ArtifactsForCommit, out, pageNumber)
		if err != nil {
			return err
		}
	}
	if len(searchResult.ArtifactsForFingerprint) > 0 {
		fmt.Fprintf(out, "Found the following artifact\n")
		err = printArtifactsJsonAsTable(searchResult.ArtifactsForFingerprint, out, pageNumber)
		if err != nil {
			return err
		}
	}

	historyEvents := []HistoryEvent{}
	printHistory := false

	if len(searchResult.EnvironmentEvents) > 0 {
		fmt.Fprintf(out, "Artifact has no provenance\n")
		// err = printEventsForSingleFingerprintAsTable(searchResult.EnvironmentEvents, out)
		events, err := getHistoryEventsForSingleFingerprint(searchResult.EnvironmentEvents)
		if err != nil {
			return err
		}
		historyEvents = append(historyEvents, events...)
		printHistory = true
	}
	if len(searchResult.Allowlist) > 0 {
		// err = printAllowlistAsTable(searchResult.Allowlist, out)
		events, err := getHistoryEventsForAllowlist(searchResult.Allowlist)
		if err != nil {
			return err
		}
		historyEvents = append(historyEvents, events...)
		printHistory = true
	}

	if printHistory {
		fmt.Fprintf(out, "Found the following environment events for artifact:\n")
		header := []string{}
		rows := []string{}

		sort.Slice(historyEvents, func(i, j int) bool {
			return historyEvents[i].Timestamp < historyEvents[j].Timestamp
		})

		for _, event := range historyEvents {
			createdAt, err := formattedTimestamp(event.Timestamp, true)
			if err != nil {
				createdAt = "bad timestamp"
			}
			rows = append(rows, fmt.Sprintf("    %s\t%s", event.Description, createdAt))
		}
		tabFormattedPrint(out, header, rows)
	}

	return nil
}

func timestampToFloat64(timestamp interface{}) (float64, error) {
	var floatTimestamp float64
	switch t := timestamp.(type) {
	case int64:
		floatTimestamp = float64(timestamp.(int64))
	case float64:
		floatTimestamp = timestamp.(float64)
	case string:
		var err error
		floatTimestamp, err = strconv.ParseFloat(timestamp.(string), 64)
		if err != nil {
			return 0.0, err
		}
	case nil:
		return 0.0, nil
	default:
		return 0.0, fmt.Errorf("unsupported timestamp type %s", t)
	}
	return floatTimestamp, nil
}

func getHistoryEventsForSingleFingerprint(events []map[string]interface{}) ([]HistoryEvent, error) {
	historyEvents := []HistoryEvent{}
	for _, event := range events {
		env_name := fmt.Sprintf("%s#%d", event["environment_name"], int(event["snapshot_index"].(float64)))
		description := event["description"]
		timestamp, err := timestampToFloat64(event["reported_at"])
		if err != nil {
			return nil, err
		}

		historyEvent := HistoryEvent{
			Description: fmt.Sprintf("%s in %s", description, env_name),
			Timestamp:   timestamp,
		}
		historyEvents = append(historyEvents, historyEvent)
	}
	return historyEvents, nil
}

func getHistoryEventsForAllowlist(events []map[string]interface{}) ([]HistoryEvent, error) {
	historyEvents := []HistoryEvent{}
	for _, event := range events {
		env_name := event["env_name"]
		user_name := event["user_name"]
		// description := event["description"]
		timestamp, err := timestampToFloat64(event["created_at"])
		if err != nil {
			return nil, err
		}
		action := ""
		if event["active"] == true {
			action = "Allow listed"
		} else {
			action = "Revoked allow listing"
		}

		historyEvent := HistoryEvent{
			Description: fmt.Sprintf("%s by %s in %s", action, user_name, env_name),
			Timestamp:   timestamp,
		}
		historyEvents = append(historyEvents, historyEvent)
	}
	return historyEvents, nil
}

// func printEventsForSingleFingerprintAsTable(events []map[string]interface{}, out io.Writer) error {
// 	if len(events) == 0 {
// 		fmt.Fprintf(out, "	No events were found")
// 		return nil
// 	}
// 	header := []string{"ENVIRONMENT", "EVENT", "REPORTED AT"}
// 	rows := []string{}
// 	for _, event := range events {
// 		env_name := fmt.Sprintf("%s#%d", event["environment_name"], int(event["snapshot_index"].(float64)))
// 		description := event["description"]
// 		reportedAt, err := formattedTimestamp(event["reported_at"], true)
// 		if err != nil {
// 			return err
// 		}
// 		rows = append(rows, fmt.Sprintf("%s\t%s\t%s", env_name, description, reportedAt))
// 	}
// 	rows = append(rows, "\t\t") // These tabs are required for alignment
// 	tabFormattedPrint(out, header, rows)
// 	return nil
// }

// func printAllowlistAsTable(allowlist []map[string]interface{}, out io.Writer) error {
// 	if len(allowlist) == 0 {
// 		fmt.Fprintf(out, "	No artifacts in allowlist were found")
// 		return nil
// 	}

// 	header := []string{"STATE", "ENVIRONMENT", "DESCRIPTION", "USER", "CREATED AT"}
// 	rows := []string{}
// 	for _, artifact := range allowlist {
// 		state := "Active"
// 		if !artifact["active"].(bool) {
// 			state = "Revoked"
// 		}
// 		env_name := artifact["env_name"]
// 		description := artifact["description"]
// 		user := artifact["user_name"]
// 		createdAt, err := formattedTimestamp(artifact["timestamp"], true)
// 		if err != nil {
// 			return err
// 		}
// 		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", state, env_name, description, user, createdAt))
// 	}
// 	rows = append(rows, "\t\t\t\t") // These tabs are required for alignment
// 	tabFormattedPrint(out, header, rows)
// 	return nil
// }
