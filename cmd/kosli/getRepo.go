package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getRepoShortDesc = `Get a repo for an org.`

const getRepoLongDesc = getRepoShortDesc + `
The repo is identified either by its name, specified as an argument
(e.g. "my-org/my-repo"), or unambiguously by its internal ID via --repo-id.
The output includes the repo's internal ID, which is the identifier used
to tag the repo (see: kosli tag).
Use --provider to disambiguate when multiple repos share the same name
across VCS providers.`

const getRepoExample = `
# get a repo
kosli get repo my-org/my-repo \
	--api-token yourAPIToken \
	--org KosliOrgName

# get a repo whose name exists across multiple VCS providers
kosli get repo my-org/my-repo \
	--provider github \
	--api-token yourAPIToken \
	--org KosliOrgName

# get a repo by its internal ID
kosli get repo --repo-id yourRepoID \
	--api-token yourAPIToken \
	--org KosliOrgName`

type getRepoOptions struct {
	output   string
	provider string
	repoID   string
}

func newGetRepoCmd(out io.Writer) *cobra.Command {
	o := new(getRepoOptions)
	cmd := &cobra.Command{
		Use:     "repo [REPO-NAME]",
		Short:   getRepoShortDesc,
		Long:    getRepoLongDesc,
		Example: getRepoExample,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if (len(args) == 1) == (o.repoID != "") {
				return ErrorBeforePrintingUsage(cmd, "exactly one of the REPO-NAME argument or --repo-id must be provided")
			}
			if o.provider != "" && o.repoID != "" {
				return ErrorBeforePrintingUsage(cmd, "--provider cannot be combined with --repo-id")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().StringVar(&o.provider, "provider", "", "[optional] The VCS provider of the repo (e.g. github, gitlab). Required when multiple repos share the same name across providers.")
	cmd.Flags().StringVar(&o.repoID, "repo-id", "", "[optional] The repo's internal ID (as shown in the repo output). Identifies the repo unambiguously; cannot be combined with the REPO-NAME argument.")

	return cmd
}

func (o *getRepoOptions) run(out io.Writer, args []string) error {
	// the endpoint's path is the repo name; when fetching by internal id the
	// server ignores the path, so the id doubles as the path segment
	pathName := o.repoID
	if len(args) == 1 {
		pathName = args[0]
	}
	reqURL, err := neturl.JoinPath(global.Host, "api/v2/repos", global.Org, pathName)
	if err != nil {
		return err
	}
	params := neturl.Values{}
	if o.repoID != "" {
		params.Set("id", o.repoID)
	} else if o.provider != "" {
		params.Set("provider", o.provider)
	}
	if len(params) > 0 {
		reqURL += "?" + params.Encode()
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printRepoAsTable,
			"json":  output.PrintJson,
		})
}

func printRepoAsTable(raw string, out io.Writer, page int) error {
	var repo map[string]any

	err := json.Unmarshal([]byte(raw), &repo)
	if err != nil {
		return err
	}

	tagsOutput := formatRepoTags(repo["tags"])
	if tagsOutput == "" {
		tagsOutput = "None"
	}
	rows := []string{
		fmt.Sprintf("Name:\t%v", repo["name"]),
		fmt.Sprintf("ID:\t%v", repo["id"]),
		fmt.Sprintf("URL:\t%v", repo["url"]),
		fmt.Sprintf("Provider:\t%v", repo["provider"]),
		fmt.Sprintf("Tags:\t%s", tagsOutput),
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}

// formatRepoTags renders a repo's tags map as sorted "key=value" pairs,
// or "" when there are no tags.
func formatRepoTags(rawTags any) string {
	tags, ok := rawTags.(map[string]any)
	if !ok || len(tags) == 0 {
		return ""
	}
	keys := make([]string, 0, len(tags))
	for key := range tags {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(tags))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, tags[key]))
	}
	return strings.Join(pairs, ", ")
}
