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

const listReposShortDesc = `List repos for an org.`

const listReposLongDesc = listReposShortDesc + ` The results are always paginated:
by default the first page is returned with 15 repos per page. Use --page to select
a page and --page-limit to change the page size (maximum 50).
The list can be filtered by name with --name (exact match), by name substring with
--search (case-insensitive, mutually exclusive with --name), by VCS provider with
--provider, by external repo ID with --repo-id, and by tags with --tag.
Results are sorted by repo name; use --sort-direction to choose asc or desc.`

const listReposExample = `
# list repos for an org (first page, 15 per page):
kosli list repos \
	--api-token yourAPIToken \
	--org yourOrgName

# list repos filtered by name (exact match on the full repo name):
kosli list repos \
	--name my-org/my-repo \
	--api-token yourAPIToken \
	--org yourOrgName

# list repos whose name contains a substring (case-insensitive):
kosli list repos \
	--search cli \
	--api-token yourAPIToken \
	--org yourOrgName

# list repos filtered by VCS provider (in JSON):
kosli list repos \
	--provider github \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json

# list repos tagged with team=platform:
kosli list repos \
	--tag team:platform \
	--api-token yourAPIToken \
	--org yourOrgName

# list repos sorted by name, Z–A:
kosli list repos \
	--sort-direction desc \
	--api-token yourAPIToken \
	--org yourOrgName

# show the second page of repos (25 per page):
kosli list repos \
	--page-limit 25 \
	--page 2 \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listReposOptions struct {
	listOptions
	name          string
	search        string
	provider      string
	repoID        string
	sortDirection string
	tags          []string
}

func newListReposCmd(out io.Writer) *cobra.Command {
	o := new(listReposOptions)
	cmd := &cobra.Command{
		Use:     "repos",
		Short:   listReposShortDesc,
		Long:    listReposLongDesc,
		Example: listReposExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return o.validate(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	addListFlags(cmd, &o.listOptions)
	cmd.Flags().StringVar(&o.name, "name", "", "[optional] The repo name to filter by (exact match).")
	cmd.Flags().StringVar(&o.search, "search", "", "[optional] Filter repos whose name contains this substring (case-insensitive). Mutually exclusive with --name.")
	cmd.Flags().StringVar(&o.provider, "provider", "", "[optional] The VCS provider to filter repos by (e.g. github, gitlab).")
	cmd.Flags().StringVar(&o.repoID, "repo-id", "", "[optional] The external repo ID to filter repos by.")
	cmd.Flags().StringVar(&o.sortDirection, "sort-direction", "", "[optional] The direction to sort repos by name. Valid values are: [asc, desc]. (defaults to asc)")
	cmd.Flags().StringSliceVar(&o.tags, "tag", []string{}, "[optional] Only list repos that have this tag, given as 'key' or 'key:value'. Can be repeated to match more than one tag.")
	cmd.MarkFlagsMutuallyExclusive("name", "search")

	return cmd
}

func (o *listReposOptions) run(out io.Writer) error {
	params := neturl.Values{}
	params.Set("page", fmt.Sprintf("%d", o.pageNumber))
	params.Set("per_page", fmt.Sprintf("%d", o.pageLimit))
	if o.name != "" {
		params.Set("name", o.name)
	}
	if o.search != "" {
		params.Set("search", o.search)
	}
	if o.provider != "" {
		params.Set("provider", o.provider)
	}
	if o.repoID != "" {
		params.Set("repo_id", o.repoID)
	}
	if o.sortDirection != "" {
		params.Set("sort_direction", o.sortDirection)
	}
	for _, tag := range o.tags {
		params.Add("tag", tag)
	}
	base, err := neturl.JoinPath(global.Host, "api/v2/repos", global.Org)
	if err != nil {
		return err
	}
	reqURL := base + "?" + params.Encode()

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printReposListAsTable,
			"json":  output.PrintJson,
		})
}

type listReposResponse struct {
	Repos      []map[string]any `json:"repos"`
	Page       int              `json:"page"`
	TotalPages int              `json:"total_pages"`
	TotalCount int              `json:"total_count"`
}

func printReposListAsTable(raw string, out io.Writer, page int) error {
	response := &listReposResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	// both the empty-list message and the footer read the page from the
	// response envelope (the server echoes the requested page), so the two
	// paths never disagree on which page is being reported
	if len(response.Repos) == 0 {
		msg := "No repos were found"
		if response.Page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, response.Page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"NAME", "URL", "PROVIDER", "TAGS"}
	rows := []string{}
	for _, repo := range response.Repos {
		row := fmt.Sprintf("%v\t%v\t%v\t%s", repo["name"], repo["url"], repo["provider"], formatRepoTags(repo["tags"]))
		rows = append(rows, row)
	}

	rows = append(rows, fmt.Sprintf("\nShowing page %d of %d, total %d repos", response.Page, response.TotalPages, response.TotalCount))

	tabFormattedPrint(out, header, rows)

	return nil
}
