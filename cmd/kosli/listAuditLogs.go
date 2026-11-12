package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listAuditLogsShortDesc = `List audit log events for an org.`

const listAuditLogsLongDesc = listAuditLogsShortDesc + `
Each event records a change to an org member's role: a member being added, changing role, or being removed.
Listing the audit log requires org admin rights.
The list can be filtered by user, event type, role and source. When multiple filters are provided, only events matching all of them are returned.
The results are paginated and ordered from latest to oldest by default; use --page and --page-limit to navigate the pages.`

const listAuditLogsExample = `
# list the first page of audit log events for an org:
kosli list audit-logs \
	--api-token yourAPIToken \
	--org yourOrgName

# list the second page of audit log events (30 per page):
kosli list audit-logs \
	--page 2 \
	--page-limit 30 \
	--api-token yourAPIToken \
	--org yourOrgName

# list audit log events in JSON:
kosli list audit-logs \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName

# list the role changes that affected or were made by users whose name contains "alice":
kosli list audit-logs \
	--search alice \
	--event-type changed \
	--api-token yourAPIToken \
	--org yourOrgName

# list the events that made someone an admin or a member through SCIM (--role and --source can be repeated):
kosli list audit-logs \
	--role admin \
	--role member \
	--source scim \
	--api-token yourAPIToken \
	--org yourOrgName

# list audit log events sorted by user name in ascending order:
kosli list audit-logs \
	--sort-by user \
	--sort-direction asc \
	--api-token yourAPIToken \
	--org yourOrgName
`

var (
	auditLogEventTypes     = []string{"added", "changed", "removed"}
	auditLogSortColumns    = []string{"when", "user", "role", "changed_by", "source"}
	auditLogSortDirections = []string{"asc", "desc"}
)

type listAuditLogsOptions struct {
	listOptions
	search        string
	eventType     string
	roles         []string
	sources       []string
	sortBy        string
	sortDirection string
}

func (o *listAuditLogsOptions) validate(cmd *cobra.Command) error {
	if err := o.listOptions.validate(cmd); err != nil {
		return err
	}
	for _, flag := range []struct {
		name   string
		value  string
		values []string
	}{
		{"event-type", o.eventType, auditLogEventTypes},
		{"sort-by", o.sortBy, auditLogSortColumns},
		{"sort-direction", o.sortDirection, auditLogSortDirections},
	} {
		if flag.value != "" && !slices.Contains(flag.values, flag.value) {
			return ErrorBeforePrintingUsage(cmd, fmt.Sprintf("invalid --%s %q: valid values are [%s]",
				flag.name, flag.value, strings.Join(flag.values, ", ")))
		}
	}
	return nil
}

type auditLogEvent struct {
	User         string `json:"user"`
	ChangedBy    string `json:"changed_by"`
	Role         string `json:"role"`
	PreviousRole string `json:"previous_role"`
	Source       string `json:"source"`
	CreatedAt    any    `json:"created_at"`
}

type listAuditLogsResponse struct {
	Events     []auditLogEvent `json:"events"`
	Page       int             `json:"page"`
	TotalPages int             `json:"total_pages"`
	TotalCount int             `json:"total_count"`
}

func newListAuditLogsCmd(out io.Writer) *cobra.Command {
	o := new(listAuditLogsOptions)
	cmd := &cobra.Command{
		Use:     "audit-logs",
		Aliases: []string{"al", "als", "audit-log"},
		Short:   listAuditLogsShortDesc,
		Long:    listAuditLogsLongDesc,
		Example: listAuditLogsExample,
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
	cmd.Flags().StringVar(&o.search, "search", "", auditLogSearchFlag)
	cmd.Flags().StringVar(&o.eventType, "event-type", "", auditLogEventTypeFlag)
	cmd.Flags().StringArrayVar(&o.roles, "role", []string{}, auditLogRoleFlag)
	cmd.Flags().StringArrayVar(&o.sources, "source", []string{}, auditLogSourceFlag)
	cmd.Flags().StringVar(&o.sortBy, "sort-by", "", auditLogSortByFlag)
	cmd.Flags().StringVar(&o.sortDirection, "sort-direction", "", auditLogSortDirectionFlag)

	return cmd
}

func (o *listAuditLogsOptions) run(out io.Writer) error {
	base, err := url.JoinPath(global.Host, "api/v2/membership-audit-log", global.Org)
	if err != nil {
		return err
	}

	params := url.Values{}
	params.Set("page", strconv.Itoa(o.pageNumber))
	params.Set("per_page", strconv.Itoa(o.pageLimit))
	if o.search != "" {
		params.Set("search", o.search)
	}
	if o.eventType != "" {
		params.Set("event_type", o.eventType)
	}
	for _, role := range o.roles {
		params.Add("role", role)
	}
	for _, source := range o.sources {
		params.Add("source", source)
	}
	if o.sortBy != "" {
		params.Set("sort_by", o.sortBy)
	}
	if o.sortDirection != "" {
		params.Set("sort_dir", o.sortDirection)
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
			"table": printAuditLogsListAsTable,
			"json":  output.PrintJson,
		})
}

func printAuditLogsListAsTable(raw string, out io.Writer, page int) error {
	response := &listAuditLogsResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	if len(response.Events) == 0 {
		msg := "No audit log events were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		logger.Info("%s.", msg)
		return nil
	}

	header := []string{"WHEN", "USER", "ROLE", "PREVIOUS ROLE", "CHANGED BY", "SOURCE"}
	rows := []string{}
	for _, event := range response.Events {
		when, err := formattedTimestamp(event.CreatedAt, true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s",
			when, event.User, event.Role, event.PreviousRole, event.ChangedBy, event.Source))
	}

	rows = append(rows, fmt.Sprintf("\nShowing page %d of %d, total %d events", response.Page, response.TotalPages, response.TotalCount))

	tabFormattedPrint(out, header, rows)

	return nil
}
