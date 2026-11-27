package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"strconv"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listTrailEventsShortDesc = `List the events of a trail.`

const listTrailEventsLongDesc = listTrailEventsShortDesc + `
Each event has a ^type^, a ^timestamp^ and, when it was reported for a commit, its ^git_commit_info^.

The events are paged by cursor. A page that has a next page ends with the ^--after^ value that fetches it,
and a page that has a previous page with the ^--before^ value that fetches it.
A cursor holds only a position, so repeat the ^--sort^ and ^--page-limit^ of the command that printed it.
In JSON output the same cursors are the ^after^ and ^before^ query parameters of ^_links.next.href^ and ^_links.prev.href^.`

const listTrailEventsExample = `
# list the events of a trail:
kosli list trail-events yourTrailName \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName

# list the events of a trail, oldest first:
kosli list trail-events yourTrailName \
	--flow yourFlowName \
	--sort oldest \
	--api-token yourAPIToken \
	--org yourOrgName

# list the 10 newest events of a trail:
kosli list trail-events yourTrailName \
	--flow yourFlowName \
	--page-limit 10 \
	--api-token yourAPIToken \
	--org yourOrgName

# list the next page of events, with the cursor the previous command printed
# and the same --sort and --page-limit:
kosli list trail-events yourTrailName \
	--flow yourFlowName \
	--page-limit 10 \
	--after yourCursor \
	--api-token yourAPIToken \
	--org yourOrgName

# list the events of a trail (in JSON):
kosli list trail-events yourTrailName \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listTrailEventsOptions struct {
	output    string
	flowName  string
	sort      string
	pageLimit int
	after     string
	before    string
}

func (o *listTrailEventsOptions) validate(cmd *cobra.Command) error {
	if o.sort != "newest" && o.sort != "oldest" {
		return ErrorBeforePrintingUsage(cmd, "--sort must be one of: newest, oldest")
	}
	if o.pageLimit < 1 || o.pageLimit > 100 {
		return ErrorBeforePrintingUsage(cmd, "--page-limit must be between 1 and 100")
	}
	if err := MuXRequiredFlags(cmd, []string{"after", "before"}, false); err != nil {
		return ErrorBeforePrintingUsage(cmd, err.Error())
	}
	return nil
}

func newListTrailEventsCmd(out io.Writer) *cobra.Command {
	o := new(listTrailEventsOptions)
	cmd := &cobra.Command{
		Use:     "trail-events TRAIL-NAME",
		Short:   listTrailEventsShortDesc,
		Long:    listTrailEventsLongDesc,
		Example: listTrailEventsExample,
		Args:    cobra.ExactArgs(1),
		Annotations: map[string]string{
			docgen.JSONOutputAPIAnnotation: "trails/list-trail-events",
		},
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

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().StringVar(&o.sort, "sort", "newest", trailEventsSortFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", 50, trailEventsPageLimitFlag)
	cmd.Flags().StringVar(&o.after, "after", "", trailEventsAfterFlag)
	cmd.Flags().StringVar(&o.before, "before", "", trailEventsBeforeFlag)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *listTrailEventsOptions) run(out io.Writer, args []string) error {
	base, err := neturl.JoinPath(global.Host, "api/v2/trails", global.Org, o.flowName, args[0], "events")
	if err != nil {
		return err
	}
	q := neturl.Values{}
	q.Set("sort", o.sort)
	q.Set("per_page", strconv.Itoa(o.pageLimit))
	if o.after != "" {
		q.Set("after", o.after)
	}
	if o.before != "" {
		q.Set("before", o.before)
	}

	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    base + "?" + q.Encode(),
		Token:  global.ApiToken,
	})
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printTrailEventsAsTable,
			"json":  output.PrintJson,
		})
}

type trailEvent struct {
	Type                  string      `json:"type"`
	Timestamp             json.Number `json:"timestamp"`
	TemplateReferenceName string      `json:"template_reference_name"`
	AttestationName       string      `json:"attestation_name"`
	ArtifactName          string      `json:"artifact_name"`
	GitCommitInfo         *struct {
		Sha1 string `json:"sha1"`
	} `json:"git_commit_info"`
}

// name is what the event is about: a template name where the event type has
// one, since that is the name the trail's compliance is reported under.
func (e trailEvent) name() string {
	for _, name := range []string{e.TemplateReferenceName, e.AttestationName, e.ArtifactName} {
		if name != "" {
			return name
		}
	}
	return ""
}

func (e trailEvent) shortCommit() string {
	if e.GitCommitInfo == nil {
		return ""
	}
	return fmt.Sprintf("%.7s", e.GitCommitInfo.Sha1)
}

type listTrailEventsResponse struct {
	Events []trailEvent `json:"events"`
	Links  map[string]struct {
		Href string `json:"href"`
	} `json:"_links"`
}

// cursor returns the token of the page that link leads to, which the API
// publishes only as the param query parameter of the link's href.
func (r *listTrailEventsResponse) cursor(link, param string) string {
	href, err := neturl.Parse(r.Links[link].Href)
	if err != nil {
		return ""
	}
	return href.Query().Get(param)
}

// cursorHints returns the lines that tell a table reader how to reach the
// pages either side, the first led by a blank line, or none at the only page.
func (r *listTrailEventsResponse) cursorHints() []string {
	hints := []string{}
	if next := r.cursor("next", "after"); next != "" {
		hints = append(hints, "Next page: --after "+next)
	}
	if prev := r.cursor("prev", "before"); prev != "" {
		hints = append(hints, "Previous page: --before "+prev)
	}
	if len(hints) > 0 {
		hints[0] = "\n" + hints[0]
	}
	return hints
}

func printTrailEventsAsTable(raw string, out io.Writer, page int) error {
	response := &listTrailEventsResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	if len(response.Events) == 0 {
		logger.Info("No events were found.")
		tabFormattedPrint(out, []string{}, response.cursorHints())
		return nil
	}

	header := []string{"TIMESTAMP", "TYPE", "NAME", "COMMIT"}
	rows := []string{}
	for _, event := range response.Events {
		timestamp, err := formattedTimestamp(event.Timestamp, true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s", timestamp, event.Type, event.name(), event.shortCommit()))
	}
	tabFormattedPrint(out, header, append(rows, response.cursorHints()...))

	return nil
}
