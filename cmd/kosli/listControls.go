package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listControlsShortDesc = `List controls for an org.`

const listControlsLongDesc = listControlsShortDesc + `
The results are paginated; use --page and --page-limit to navigate the pages.`

const listControlsExample = `
# list the first page of controls for an org:
kosli list controls \
	--api-token yourAPIToken \
	--org yourOrgName

# list the second page of controls (10 per page) in JSON:
kosli list controls \
	--page 2 \
	--page-limit 10 \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName

# list controls whose name or identifier contains "sdlc":
kosli list controls \
	--search sdlc \
	--api-token yourAPIToken \
	--org yourOrgName

# list controls tagged framework=finos-sdlc (--tag can be repeated):
kosli list controls \
	--tag framework:finos-sdlc \
	--api-token yourAPIToken \
	--org yourOrgName

# list archived controls instead of active ones:
kosli list controls \
	--archived \
	--api-token yourAPIToken \
	--org yourOrgName
`

type listControlsOptions struct {
	listOptions
	search   string
	tags     []string
	archived bool
}

type listControlsResponse struct {
	Controls   []map[string]interface{} `json:"controls"`
	Page       int                      `json:"page"`
	TotalPages int                      `json:"total_pages"`
	TotalCount int                      `json:"total_count"`
}

func newListControlsCmd(out io.Writer) *cobra.Command {
	o := new(listControlsOptions)
	cmd := &cobra.Command{
		Use:         "controls",
		Short:       listControlsShortDesc,
		Long:        listControlsLongDesc,
		Example:     listControlsExample,
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
	cmd.Flags().StringVar(&o.search, "search", "", controlSearchFlag)
	cmd.Flags().StringArrayVar(&o.tags, "tag", []string{}, controlTagFlag)
	cmd.Flags().BoolVar(&o.archived, "archived", false, controlArchivedFlag)

	return cmd
}

func (o *listControlsOptions) run(out io.Writer) error {
	base, err := url.JoinPath(global.Host, "api/v2/controls", global.Org)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("page", strconv.Itoa(o.pageNumber))
	params.Set("per_page", strconv.Itoa(o.pageLimit))
	if o.search != "" {
		params.Set("search", o.search)
	}
	for _, tag := range o.tags {
		params.Add("tag", tag)
	}
	if o.archived {
		params.Set("archived", "true")
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
			"table": printControlsListAsTable,
			"json":  output.PrintJson,
		})
}

func printControlsListAsTable(raw string, out io.Writer, page int) error {
	response := &listControlsResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	if len(response.Controls) == 0 {
		msg := "No controls were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info(msg + ".")
		return nil
	}

	header := []string{"IDENTIFIER", "NAME", "DESCRIPTION", "CREATED AT"}
	rows := []string{}
	for _, control := range response.Controls {
		description := control["description"]
		if description == nil {
			description = ""
		}
		createdAt := ""
		if control["created_at"] != nil {
			createdAt, _ = formattedTimestamp(control["created_at"], true)
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s", control["identifier"], control["name"], description, createdAt))
	}

	rows = append(rows, fmt.Sprintf("\nShowing page %d of %d, total %d controls", response.Page, response.TotalPages, response.TotalCount))

	tabFormattedPrint(out, header, rows)

	return nil
}
