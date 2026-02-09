package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const logEnvironmentShortDesc = `List environment events.`

const logEnvironmentLongDesc = logEnvironmentShortDesc + `
The results are paginated and ordered from latest to oldest.
By default, the page limit is 15 events per page.

You can optionally specify an INTERVAL between two snapshot expressions with [expression]..[expression]. 

Expressions can be:
* ~N   N'th behind the latest snapshot  
* N    snapshot number N  
* NOW  the latest snapshot  

Either expression can be omitted to default to NOW.
`

const logEnvironmentExample = `
# list the last 15 events for an environment:
kosli log environment yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 events for an environment:
kosli log environment yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# list the last 30 events for an environment (in JSON):
kosli log environment yourEnvironmentName \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json

# list events for an environment filtered by repo:
kosli log environment yourEnvironmentName \
	--repo yourOrg/yourRepo \
	--api-token yourAPIToken \
	--org yourOrgName

# list events for an environment filtered by multiple repos:
kosli log environment yourEnvironmentName \
	--repo yourOrg/yourRepo1 \
	--repo yourOrg/yourRepo2 \
	--api-token yourAPIToken \
	--org yourOrgName
`

type logEnvironmentOptions struct {
	listOptions
	reverse  bool
	interval string
	repos    []string
}

func newLogEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(logEnvironmentOptions)
	cmd := &cobra.Command{
		Use:     "environment ENV_NAME",
		Aliases: []string{"env"},
		Short:   logEnvironmentShortDesc,
		Long:    logEnvironmentLongDesc,
		Example: logEnvironmentExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return o.validate(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.interval, "interval", "i", "", intervalFlag)
	cmd.Flags().StringSliceVar(&o.repos, "repo", []string{}, repoNameFlag)
	addListFlags(cmd, &o.listOptions)
	cmd.Flags().BoolVar(&o.reverse, "reverse", false, reverseFlag)

	return cmd
}

func (o *logEnvironmentOptions) run(out io.Writer, args []string) error {
	envName := args[0]

	return o.getEnvironmentEvents(out, envName, o.interval)

}

// events

func (o *logEnvironmentOptions) getEnvironmentEvents(out io.Writer, envName, interval string) error {
	baseURL := fmt.Sprintf("%s/api/v2/environments/%s/%s/events", global.Host, global.Org, envName)
	u, err := url.Parse(baseURL)
	if err != nil {
		return fmt.Errorf("failed to parse events URL: %w", err)
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(o.pageNumber))
	q.Set("per_page", strconv.Itoa(o.pageLimit))
	q.Set("interval", interval)
	q.Set("reverse", strconv.FormatBool(o.reverse))
	for _, repo := range o.repos {
		if repo != "" {
			q.Add("repo_name", repo)
		}
	}
	u.RawQuery = q.Encode()

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    u.String(),
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}
	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printEnvironmentEventsLogAsTable,
			"json":  output.PrintJson,
		})
}
