package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const updateControlShortDesc = `Update a Kosli control.`

const updateControlLongDesc = updateControlShortDesc + `

Only the flags you provide are changed; omitted fields are left untouched.
Providing ^--link^ replaces all of the control's existing links.`

const updateControlExample = `
# update a control's name:
kosli update control yourControlIdentifier \
	--name "New control name" \
	--api-token yourAPIToken \
	--org yourOrgName

# update a control's description and links:
kosli update control yourControlIdentifier \
	--description "what this control checks" \
	--link runbook=https://example.com/runbook \
	--api-token yourAPIToken \
	--org yourOrgName
`

type updateControlOptions struct {
	name        string
	description string
	links       map[string]string
}

func newUpdateControlCmd(out io.Writer) *cobra.Command {
	o := new(updateControlOptions)
	cmd := &cobra.Command{
		Use:         "control CONTROL-IDENTIFIER",
		Short:       updateControlShortDesc,
		Long:        updateControlLongDesc,
		Example:     updateControlExample,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return RequireAtLeastOneOfFlags(cmd, []string{"name", "description", "link"})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd, args)
		},
	}

	cmd.Flags().StringVarP(&o.name, "name", "n", "", updateControlNameFlag)
	cmd.Flags().StringVarP(&o.description, "description", "d", "", controlDescriptionFlag)
	cmd.Flags().StringToStringVar(&o.links, "link", map[string]string{}, controlLinkFlag)

	addDryRunFlag(cmd)

	return cmd
}

func (o *updateControlOptions) run(cmd *cobra.Command, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/controls", global.Org, args[0])
	if err != nil {
		return err
	}

	// Only send the fields the user explicitly set, so unset flags leave the
	// corresponding values unchanged (the server treats an omitted field as
	// "no change").
	payload := map[string]interface{}{}
	if cmd.Flags().Changed("name") {
		payload["name"] = o.name
	}
	if cmd.Flags().Changed("description") {
		payload["description"] = o.description
	}
	if cmd.Flags().Changed("link") {
		payload["links"] = o.links
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     url,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("control %s was updated", args[0])
	}
	return err
}
