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

const listReposDesc = `List repos for an org.`

type listReposOptions struct {
	listOptions
	name     string
	provider string
	repoID   string
}

func newListReposCmd(out io.Writer) *cobra.Command {
	o := new(listReposOptions)
	cmd := &cobra.Command{
		Use:    "repos",
		Hidden: true,
		Short:  listReposDesc,
		Long:   listReposDesc,
		Args:   cobra.NoArgs,
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
	cmd.Flags().StringVar(&o.name, "name", "", "[optional] The repo name to filter by.")
	cmd.Flags().StringVar(&o.provider, "provider", "", "[optional] The VCS provider to filter repos by (e.g. github, gitlab).")
	cmd.Flags().StringVar(&o.repoID, "repo-id", "", "[optional] The external repo ID to filter repos by.")

	return cmd
}

func (o *listReposOptions) run(out io.Writer) error {
	params := neturl.Values{}
	params.Set("page", fmt.Sprintf("%d", o.pageNumber))
	params.Set("per_page", fmt.Sprintf("%d", o.pageLimit))
	if o.name != "" {
		params.Set("name", o.name)
	}
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

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printReposListAsTable,
			"json":  output.PrintJson,
		})
}

func printReposListAsTable(raw string, out io.Writer, page int) error {
	var repos []map[string]any
	var response struct {
		Embedded struct {
			Repos []map[string]any `json:"repos"`
		} `json:"_embedded"`
	}

	err := json.Unmarshal([]byte(raw), &response)
	if err != nil {
		return err
	}
	repos = response.Embedded.Repos

	if len(repos) == 0 {
		logger.Info("No repos were found.")
		return nil
	}

	header := []string{"NAME", "URL", "PROVIDER", "LAST_ACTIVITY"}
	rows := []string{}
	for _, repo := range repos {
		row := fmt.Sprintf("%s\t%s\t%s\t%s", repo["name"], repo["url"], repo["provider"], repo["latest_activity"])
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
