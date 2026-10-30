package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listRegoPoliciesShortDesc = `List rego policies for an org.`

const listRegoPoliciesLongDesc = listRegoPoliciesShortDesc + `
The results are paginated; use ^--page^ and ^--page-limit^ to navigate the pages.
By default, the policies changed most recently are listed first.`

const listRegoPoliciesExample = `
# list the first page of rego policies for an org:
kosli list rego-policies \
	--api-token yourAPIToken \
	--org yourOrgName

# list the second page of rego policies (50 per page) in JSON:
kosli list rego-policies \
	--page 2 \
	--page-limit 50 \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName

# list rego policies whose name or identifier contains "approved":
kosli list rego-policies \
	--search approved \
	--api-token yourAPIToken \
	--org yourOrgName

# list rego policies in identifier order:
kosli list rego-policies \
	--sort identifier \
	--sort-direction asc \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listRegoPoliciesOptions struct {
	listOptions
	search        string
	sort          string
	sortDirection string
}

type listRegoPoliciesResponse struct {
	RegoPolicies []map[string]any `json:"rego_policies"`
	Page         int              `json:"page"`
	TotalPages   int              `json:"total_pages"`
	TotalCount   int              `json:"total_count"`
}

func newListRegoPoliciesCmd(out io.Writer) *cobra.Command {
	o := new(listRegoPoliciesOptions)
	cmd := &cobra.Command{
		Use:         "rego-policies",
		Short:       listRegoPoliciesShortDesc,
		Long:        listRegoPoliciesLongDesc,
		Example:     listRegoPoliciesExample,
		Args:        cobra.NoArgs,
		Annotations: map[string]string{betaCLIAnnotation: ""},
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
	cmd.Flags().StringVar(&o.search, "search", "", regoPolicySearchFlag)
	cmd.Flags().StringVar(&o.sort, "sort", "", regoPolicySortFlag)
	cmd.Flags().StringVar(&o.sortDirection, "sort-direction", "", regoPolicySortDirectionFlag)

	return cmd
}

func (o *listRegoPoliciesOptions) run(out io.Writer) error {
	base, err := regopolicies.ListURL(global.Host, global.Org)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("page", strconv.Itoa(o.pageNumber))
	params.Set("per_page", strconv.Itoa(o.pageLimit))
	if o.search != "" {
		params.Set("search", o.search)
	}
	if o.sort != "" {
		params.Set("sort_by", o.sort)
	}
	if o.sortDirection != "" {
		params.Set("sort_direction", o.sortDirection)
	}

	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    base + "?" + params.Encode(),
		Token:  global.ApiToken,
	})
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printRegoPoliciesListAsTable,
			"json":  output.PrintJson,
		})
}

func printRegoPoliciesListAsTable(raw string, out io.Writer, page int) error {
	response := &listRegoPoliciesResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	if len(response.RegoPolicies) == 0 {
		msg := "No rego policies were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info("%s.", msg)
		return nil
	}

	header := []string{"IDENTIFIER", "NAME", "DESCRIPTION", "CREATED BY", "LAST UPDATED"}
	rows := []string{}
	for _, policy := range response.RegoPolicies {
		description := policy["description"]
		if description == nil {
			description = ""
		}
		updatedAt, err := formattedTimestamp(policy["updated_at"], true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s",
			policy["identifier"], policy["name"], description, policy["created_by"], updatedAt))
	}

	rows = append(rows, fmt.Sprintf("\nShowing page %d of %d, total %d rego policies",
		response.Page, response.TotalPages, response.TotalCount))

	tabFormattedPrint(out, header, rows)

	return nil
}
