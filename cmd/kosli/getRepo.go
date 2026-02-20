package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getRepoShortDesc = `Get a repo for an org.`

const getRepoLongDesc = getRepoShortDesc + `
The name of the repo is specified as an argument (e.g. "my-org/my-repo").
Use --provider or --repo-id to narrow down the result when multiple repos
match the given name.`

const getRepoExample = `
# get a repo
kosli get repo my-org/my-repo \
	--api-token yourAPIToken \
	--org KosliOrgName

# get a repo filtering by provider
kosli get repo my-org/my-repo \
	--provider github \
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
		Use:     "repo REPO-NAME",
		Hidden:  true,
		Short:   getRepoShortDesc,
		Long:    getRepoLongDesc,
		Example: getRepoExample,
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
	cmd.Flags().StringVar(&o.provider, "provider", "", "[optional] The VCS provider to filter repos by (e.g. github, gitlab).")
	cmd.Flags().StringVar(&o.repoID, "repo-id", "", "[optional] The external repo ID to filter repos by.")

	return cmd
}

func (o *getRepoOptions) run(out io.Writer, args []string) error {
	params := neturl.Values{}
	params.Set("name", args[0])
	if o.provider != "" {
		params.Set("provider", o.provider)
	}
	if o.repoID != "" {
		params.Set("repo_id", o.repoID)
	}
	reqURL := fmt.Sprintf("%s/api/v2/repos/%s?%s", global.Host, global.Org, params.Encode())

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	var parsed struct {
		Embedded struct {
			Repos []map[string]any `json:"repos"`
		} `json:"_embedded"`
	}
	if err := json.Unmarshal([]byte(response.Body), &parsed); err != nil {
		return err
	}
	if len(parsed.Embedded.Repos) > 1 {
		return fmt.Errorf("found %d repos matching %q. Use --provider or --repo-id to narrow down the search", len(parsed.Embedded.Repos), args[0])
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printRepoAsTable,
			"json":  output.PrintJson,
		})
}

func printRepoAsTable(raw string, out io.Writer, page int) error {
	var response struct {
		Embedded struct {
			Repos []map[string]any `json:"repos"`
		} `json:"_embedded"`
	}

	err := json.Unmarshal([]byte(raw), &response)
	if err != nil {
		return err
	}

	repos := response.Embedded.Repos
	if len(repos) == 0 {
		logger.Info("Repo was not found.")
		return nil
	}

	repo := repos[0]
	rows := []string{
		fmt.Sprintf("Name:\t%s", repo["name"]),
		fmt.Sprintf("URL:\t%s", repo["url"]),
		fmt.Sprintf("Provider:\t%s", repo["provider"]),
		fmt.Sprintf("Latest Activity:\t%s", repo["latest_activity"]),
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}
