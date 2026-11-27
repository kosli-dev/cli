package main

import (
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
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.flowName, "flow", "f", "", flowNameFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "json", outputFlag)

	err := RequireFlags(cmd, []string{"flow"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *listTrailEventsOptions) run(out io.Writer, args []string) error {
	url, err := neturl.JoinPath(global.Host, "api/v2/trails", global.Org, o.flowName, args[0], "events")
	if err != nil {
		return err
	}

	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    url,
		Token:  global.ApiToken,
	})
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"json": output.PrintJson,
		})
}
