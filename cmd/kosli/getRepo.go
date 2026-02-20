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
The name of the repo is specified as an argument (e.g. "my-org/my-repo").`

const getRepoExample = `
# get a repo
kosli get repo my-org/my-repo \
	--api-token yourAPIToken \
	--org KosliOrgName`

type getRepoOptions struct {
	output string
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

	return cmd
}

func (o *getRepoOptions) run(out io.Writer, args []string) error {
	reqURL := fmt.Sprintf("%s/api/v2/repos/%s?name=%s", global.Host, global.Org, neturl.QueryEscape(args[0]))

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
