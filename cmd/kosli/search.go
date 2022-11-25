package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type searchOptions struct {
	output string
}

type SearchResponse struct {
	ResolvedTo ResolvedToBody   `json:"resolved_to"`
	Artifacts  []SearchArtifact `json:"artifacts"`
}

type SearchArtifact struct {
	Fingerprint     string                   `json:"fingerprint"`
	Name            string                   `json:"name"`
	Pipeline        string                   `json:"pipeline"`
	Commit          string                   `json:"git_commit"`
	HasProvenance   bool                     `json:"has_provenance"`
	CommitURL       string                   `json:"commit_url"`
	BuildURL        string                   `json:"build_url"`
	ComplianceState string                   `json:"compliance_state"`
	RunningIn       []string                 `json:"running_in"`
	ExitedFrom      []string                 `json:"exited_from"`
	History         []map[string]interface{} `json:"history"`
}

type ResolvedToBody struct {
	FullMatch string `json:"full_match"`
	Type      string `json:"type"`
}

const searchExample = `
# Search for a git commit in Kosli
kosli search YOUR_GIT_COMMIT \
	--api-token yourApiToken \
	--owner yourOrgName

# Search for an artifact fingerprint in Kosli
kosli search YOUR_FINGERPRINT \
	--api-token yourApiToken \
	--owner yourOrgName
`

const searchShortDesc = `Search for a git commit or artifact fingerprint in Kosli.`

const searchLongDesc = searchShortDesc + ` 
You can use short git commit or artifact fingerprint shas, but you must provide at least 5 characters.
`

func newSearchCmd(out io.Writer) *cobra.Command {
	o := new(searchOptions)
	cmd := &cobra.Command{
		Use:     "search GIT-COMMIT|FINGERPRINT",
		Short:   searchShortDesc,
		Long:    searchLongDesc,
		Example: searchExample,
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
	if searchResult.ResolvedTo.Type == "commit" {
		fmt.Fprintf(out, "Search result resolved to commit %s\n", fullMatch)
	} else {
		fmt.Fprintf(out, "Search result resolved to artifact with fingerprint %s\n", fullMatch)
	}

	rows := []string{}
	for _, artifact := range searchResult.Artifacts {
		rows = append(rows, fmt.Sprintf("Name:\t%s", artifact.Name))
		rows = append(rows, fmt.Sprintf("Fingerprint:\t%s", artifact.Fingerprint))
		rows = append(rows, fmt.Sprintf("Has provenance:\t%t", artifact.HasProvenance))
		if artifact.HasProvenance {
			rows = append(rows, fmt.Sprintf("Pipeline:\t%s", artifact.Pipeline))
			rows = append(rows, fmt.Sprintf("Git commit:\t%s", artifact.Commit))
			rows = append(rows, fmt.Sprintf("Commit URL:\t%s", artifact.CommitURL))
			rows = append(rows, fmt.Sprintf("Build URL:\t%s", artifact.BuildURL))
			rows = append(rows, fmt.Sprintf("Compliance state:\t%s", artifact.ComplianceState))
		}

		rows = append(rows, fmt.Sprintf("Running in:\t[ %s ]", strings.Join(artifact.RunningIn, ", ")))
		rows = append(rows, fmt.Sprintf("Exited from:\t[ %s ]", strings.Join(artifact.ExitedFrom, ", ")))
		rows = append(rows, "History:")
		for _, event := range artifact.History {
			timestampHuman, err := formattedTimestamp(event["timestamp"], true)
			if err != nil {
				timestampHuman = "bad timestamp"
			}
			rows = append(rows, fmt.Sprintf("    %s\t%s", event["event"], timestampHuman))
		}
	}

	tabFormattedPrint(out, []string{}, rows)

	return nil
}
