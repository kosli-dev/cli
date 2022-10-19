package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	output string
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
	var searchResult map[string][]map[string]interface{}
	err := json.Unmarshal([]byte(responseRaw), &searchResult)
	if err != nil {
		return err
	}
	if len(searchResult["artifacts_for_commit"]) > 0 {
		fmt.Fprintf(out, "Found the following artifact(s) for commit:\n")
		err = printArtifactsJsonAsTable(searchResult["artifacts_for_commit"], out, pageNumber)
		if err != nil {
			return err
		}
	}
	if len(searchResult["artifacts_for_fingerprint"]) > 0 {
		fmt.Fprintf(out, "Found the following artifact(s) for fingerprint:\n")
		err = printArtifactsJsonAsTable(searchResult["artifacts_for_fingerprint"], out, pageNumber)
		if err != nil {
			return err
		}
	}
	if len(searchResult["environment_events_for_no_provenance_artifacts"]) > 0 {
		fmt.Fprintf(out, "Found the following artifact(s) with no provenance in environments:\n")
		// TODO: print out env name to which an event belongs to
		err = printEventsForSingleFingerprintAsTable(searchResult["environment_events_for_no_provenance_artifacts"], out)
		if err != nil {
			return err
		}
	}
	if len(searchResult["allowlist"]) > 0 {
		fmt.Fprintf(out, "Found allowlisted artifacts for fingerprint:\n")
		err = printAllowlistAsTable(searchResult["allowlist"], out)
		if err != nil {
			return err
		}
	}
	return nil
}

func printEventsForSingleFingerprintAsTable(events []map[string]interface{}, out io.Writer) error {
	if len(events) == 0 {
		fmt.Fprintf(out, "	No events were found")
		return nil
	}
	header := []string{"ENVIRONMENT", "EVENT", "REPORTED AT"}
	rows := []string{}
	for _, event := range events {
		env_name := fmt.Sprintf("%s#%d", event["environment_name"], int(event["snapshot_index"].(float64)))
		description := event["description"]
		reportedAt, err := formattedTimestamp(event["reported_at"], false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s", env_name, description, reportedAt))
	}
	rows = append(rows, "\t\t") // These tabs are required for alignment
	tabFormattedPrint(out, header, rows)
	return nil
}

func printAllowlistAsTable(allowlist []map[string]interface{}, out io.Writer) error {
	if len(allowlist) == 0 {
		fmt.Fprintf(out, "	No artifacts in allowlist were found")
		return nil
	}

	header := []string{"STATE", "ENVIRONMENT", "DESCRIPTION", "USER", "CREATED AT"}
	rows := []string{}
	for _, artifact := range allowlist {
		state := "Active"
		if !artifact["active"].(bool) {
			state = "Revoked"
		}
		env_name := artifact["env_name"]
		description := artifact["description"]
		user := artifact["user_name"]
		createdAt, err := formattedTimestamp(artifact["timestamp"], true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s", state, env_name, description, user, createdAt))
	}
	rows = append(rows, "\t\t\t\t") // These tabs are required for alignment
	tabFormattedPrint(out, header, rows)
	return nil
}
