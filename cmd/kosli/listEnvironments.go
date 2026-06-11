package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listEnvironmentsShortDesc = `List environments for an org.`

const listEnvironmentsLongDesc = listEnvironmentsShortDesc + `
By default, all environments are returned in one response.
When --page or --page-limit is set, the results are paginated and the response includes pagination metadata.
The list can be filtered by name, type, space and tags, and sorted with --sort and --sort-direction.`

const listEnvironmentsExample = `
# list all environments for an org:
kosli list environments \
	--api-token yourAPIToken \
	--org yourOrgName

# show the second page of environments, 25 per page:
kosli list environments \
	--page 2 \
	--page-limit 25 \
	--api-token yourAPIToken \
	--org yourOrgName

# list environments whose name contains a substring (in JSON):
kosli list environments \
	--name prod \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName

# list K8S and ECS environments tagged with team=platform:
kosli list environments \
	--type K8S \
	--type ECS \
	--tag team:platform \
	--api-token yourAPIToken \
	--org yourOrgName

# list environments sorted by when they last changed, newest first:
kosli list environments \
	--sort last_changed_at \
	--sort-direction desc \
	--api-token yourAPIToken \
	--org yourOrgName
`

type environmentLsOptions struct {
	listOptions
	// withPagination is true when the user explicitly set --page or --page-limit.
	// Without it, no pagination params are sent and the API returns all environments.
	withPagination bool
	name           string
	envTypes       []string
	spaceIDs       []string
	tags           []string
	sort           string
	sortDirection  string
}

type paginatedEnvsResponse struct {
	Page         int64                    `json:"page"`
	PerPage      int64                    `json:"per_page"`
	TotalPages   int64                    `json:"total_pages"`
	TotalCount   int64                    `json:"total_count"`
	Environments []map[string]interface{} `json:"environments"`
}

func newListEnvironmentsCmd(out io.Writer) *cobra.Command {
	o := new(environmentLsOptions)
	cmd := &cobra.Command{
		Use:     "environments",
		Aliases: []string{"env", "envs"},
		Short:   listEnvironmentsShortDesc,
		Long:    listEnvironmentsLongDesc,
		Example: listEnvironmentsExample,
		Args:    cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			o.withPagination = cmd.Flags().Changed("page") || cmd.Flags().Changed("page-limit")
			if o.withPagination {
				return o.validate(cmd)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVar(&o.name, "name", "", envSearchNameFlag)
	cmd.Flags().StringSliceVar(&o.envTypes, "type", []string{}, envTypeFilterFlag)
	cmd.Flags().StringSliceVar(&o.spaceIDs, "space-id", []string{}, envSpaceIDFilterFlag)
	cmd.Flags().StringSliceVar(&o.tags, "tag", []string{}, envTagFilterFlag)
	cmd.Flags().StringVar(&o.sort, "sort", "", envSortFlag)
	cmd.Flags().StringVar(&o.sortDirection, "sort-direction", "", envSortDirectionFlag)
	addListFlags(cmd, &o.listOptions)

	return cmd
}

func (o *environmentLsOptions) run(out io.Writer, args []string) error {
	base, err := url.JoinPath(global.Host, "api/v2/environments", global.Org)
	if err != nil {
		return err
	}

	params := url.Values{}
	if o.withPagination {
		params.Set("page", strconv.Itoa(o.pageNumber))
		params.Set("per_page", strconv.Itoa(o.pageLimit))
	}
	if o.name != "" {
		params.Set("name", o.name)
	}
	for _, envType := range o.envTypes {
		params.Add("type", envType)
	}
	for _, spaceID := range o.spaceIDs {
		params.Add("space_id", spaceID)
	}
	for _, tag := range o.tags {
		params.Add("tag", tag)
	}
	if o.sort != "" {
		params.Set("sort", o.sort)
	}
	if o.sortDirection != "" {
		params.Set("sort_direction", o.sortDirection)
	}
	reqURL := base
	if encoded := params.Encode(); encoded != "" {
		reqURL = base + "?" + encoded
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

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printEnvListAsTable,
			"json":  output.PrintJson,
		})
}

func printEnvListAsTable(raw string, out io.Writer, page int) error {
	// the API returns a plain array when no pagination params are sent,
	// and a wrapped object with pagination metadata when they are
	var envs []map[string]interface{}
	var paginated *paginatedEnvsResponse
	if err := json.Unmarshal([]byte(raw), &envs); err != nil {
		paginated = &paginatedEnvsResponse{}
		if err := json.Unmarshal([]byte(raw), paginated); err != nil {
			return err
		}
		envs = paginated.Environments
	}

	if len(envs) == 0 {
		msg := "No environments were found"
		if page > 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"NAME", "TYPE", "LAST REPORT", "LAST MODIFIED", "TAGS", "POLICIES"}
	rows := []string{}
	for _, env := range envs {
		last_reported_str := ""
		last_reported_at := env["last_reported_at"]
		if last_reported_at != nil {
			last_reported_str = time.Unix(int64(last_reported_at.(float64)), 0).Format(time.RFC3339)
		}
		last_modified_str := ""
		last_modified_at := env["last_modified_at"]
		if last_modified_at != nil {
			last_modified_str = time.Unix(int64(last_modified_at.(float64)), 0).Format(time.RFC3339)
		}

		tagsOutput := ""
		if tags, ok := env["tags"].(map[string]interface{}); ok {
			for key, value := range tags {
				tagsOutput += fmt.Sprintf("[%s=%s], ", key, value)
			}
			tagsOutput = strings.TrimSuffix(tagsOutput, ", ")
		}

		var policies []interface{}
		if env["policies"] != nil {
			policies = env["policies"].([]interface{})
		} else {
			policies = []interface{}{}
		}

		row := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s", env["name"], env["type"], last_reported_str, last_modified_str, tagsOutput, policies)
		rows = append(rows, row)
	}
	if paginated != nil {
		rows = append(rows, fmt.Sprintf("\nShowing page %d of %d, total %d items", paginated.Page, paginated.TotalPages, paginated.TotalCount))
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
