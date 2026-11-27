package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const listTrailEventsShortDesc = `List the events of a trail.`

const listTrailEventsLongDesc = listTrailEventsShortDesc + `
Each event has a ^type^, a ^timestamp^ and, when it was reported for a commit, its ^git_commit_info^.`

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

# list the events of a trail (in JSON):
kosli list trail-events yourTrailName \
	--flow yourFlowName \
	--api-token yourAPIToken \
	--org yourOrgName \
	--output json
`

type listTrailEventsOptions struct {
	output   string
	flowName string
	sort     string
}

func (o *listTrailEventsOptions) validate(cmd *cobra.Command) error {
	if o.sort != "newest" && o.sort != "oldest" {
		return ErrorBeforePrintingUsage(cmd, "--sort must be one of: newest, oldest")
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
}

func printTrailEventsAsTable(raw string, out io.Writer, page int) error {
	response := &listTrailEventsResponse{}
	if err := json.Unmarshal([]byte(raw), response); err != nil {
		return err
	}

	if len(response.Events) == 0 {
		logger.Info("No events were found.")
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
	tabFormattedPrint(out, header, rows)

	return nil
}
