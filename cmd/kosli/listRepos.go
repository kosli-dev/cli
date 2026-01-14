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

const listReposDesc = `List repos for an org.`

type listReposOptions struct {
	listOptions
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

	return cmd
}

func (o *listReposOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/api/v2/repos/%s?page=%d&per_page=%d", global.Host, global.Org, o.pageNumber, o.pageLimit)

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.listOptions.output, out, o.pageNumber,
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

	header := []string{"NAME", "URL", "LAST_ACTIVITY"}
	rows := []string{}
	for _, repo := range repos {
		row := fmt.Sprintf("%s\t%s\t%s", repo["name"], repo["url"], repo["latest_activity"])
		rows = append(rows, row)
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
