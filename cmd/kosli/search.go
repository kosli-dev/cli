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
	Flow            string                   `json:"pipeline"`
	Commit          string                   `json:"git_commit"`
	HasProvenance   bool                     `json:"has_provenance"`
	CommitURL       string                   `json:"commit_url"`
	BuildURL        string                   `json:"build_url"`
	ArtifactURL     string                   `json:"html_url"`
	ComplianceState string                   `json:"compliance_state"`
	RunningIn       []string                 `json:"running_in"`
	ExitedFrom      []string                 `json:"exited_from"`
	History         []map[string]interface{} `json:"history"`
}

type ResolvedToBody struct {
	Type         string               `json:"type"`
	FullMatch    string               `json:"full_match"`
	Fingerprints ResolvedFingerprints `json:"fingerprints"`
	Commits      ResolvedCommits      `json:"commits"`
}

type ResolvedFingerprints struct {
	Type    string   `json:"type"`
	Count   int      `json:"count"`
	Matches []string `json:"matches"`
}

type ResolvedCommits struct {
	Type    string   `json:"type"`
	Count   int      `json:"count"`
	Matches []string `json:"matches"`
}

const searchExample = `
# Search for a git commit in Kosli
kosli search YOUR_GIT_COMMIT \
	--api-token yourApiToken \
	--org yourOrgName

# Search for an artifact fingerprint in Kosli
kosli search YOUR_ARTIFACT_FINGERPRINT \
	--api-token yourApiToken \
	--org yourOrgName
`

const searchShortDesc = `Search for a git commit or an artifact fingerprint in Kosli.  `

const searchLongDesc = searchShortDesc + ` 
You can use short git commit or artifact fingerprint shas, but you must provide at least 5 characters.`

func newSearchCmd(out io.Writer) *cobra.Command {
	o := new(searchOptions)
	cmd := &cobra.Command{
		Use:     "search {GIT-COMMIT | FINGERPRINT}",
		Short:   searchShortDesc,
		Long:    searchLongDesc,
		Example: searchExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
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

	url := fmt.Sprintf("%s/api/v2/search/%s/sha/%s", global.Host, global.Org, search_value)

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
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

	countFingerprints := searchResult.ResolvedTo.Fingerprints.Count
	countCommits := searchResult.ResolvedTo.Commits.Count
	fullMatch := searchResult.ResolvedTo.FullMatch
	if searchResult.ResolvedTo.Type == "commit" {
		logger.Info("Search result resolved to commit %s", fullMatch)
	} else if searchResult.ResolvedTo.Type == "fingerprint" {
		logger.Info("Search result resolved to artifact with fingerprint %s", fullMatch)
	} else {
		logger.Info("Search result resolved to %d fingerprint(s) and %d commit(s) across %d artifacts\n", countFingerprints, countCommits, len(searchResult.Artifacts))
	}

	rows := []string{}
	for _, artifact := range searchResult.Artifacts {
		rows = append(rows, fmt.Sprintf("Name:\t%s", artifact.Name))
		rows = append(rows, fmt.Sprintf("Fingerprint:\t%s", artifact.Fingerprint))
		rows = append(rows, fmt.Sprintf("Has provenance:\t%t", artifact.HasProvenance))
		if artifact.HasProvenance {
			rows = append(rows, fmt.Sprintf("Flow:\t%s", artifact.Flow))
			rows = append(rows, fmt.Sprintf("Git commit:\t%s", artifact.Commit))
			rows = append(rows, fmt.Sprintf("Commit URL:\t%s", artifact.CommitURL))
			rows = append(rows, fmt.Sprintf("Build URL:\t%s", artifact.BuildURL))
			rows = append(rows, fmt.Sprintf("Artifact URL:\t%s", artifact.ArtifactURL))
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
		rows = append(rows, "\n")
	}

	tabFormattedPrint(out, []string{}, rows)

	return nil
}
