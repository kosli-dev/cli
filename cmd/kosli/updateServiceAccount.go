package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const updateServiceAccountShortDesc = `Update a service account.`

const updateServiceAccountLongDesc = updateServiceAccountShortDesc + `

Only the flags you provide are changed; omitted fields are left untouched.`

const updateServiceAccountExample = `
# update a service account's description:
kosli update service-account yourServiceAccountName \
	--description "new description" \
	--api-token yourAPIToken \
	--org yourOrgName

# update a service account's privilege:
kosli update service-account yourServiceAccountName \
	--privilege member \
	--api-token yourAPIToken \
	--org yourOrgName
`

type updateServiceAccountOptions struct {
	description string
	privilege   string
}

func newUpdateServiceAccountCmd(out io.Writer) *cobra.Command {
	o := new(updateServiceAccountOptions)
	cmd := &cobra.Command{
		Use:     "service-account SERVICE-ACCOUNT-NAME",
		Aliases: []string{"sa"},
		Short:   updateServiceAccountShortDesc,
		Long:    updateServiceAccountLongDesc,
		Example: updateServiceAccountExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return RequireAtLeastOneOfFlags(cmd, []string{"description", "privilege"})
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd, args)
		},
	}

	cmd.Flags().StringVarP(&o.description, "description", "d", "", serviceAccountDescriptionFlag)
	cmd.Flags().StringVar(&o.privilege, "privilege", "", serviceAccountPrivilegeFlag)
	addDryRunFlag(cmd)

	return cmd
}

func (o *updateServiceAccountOptions) run(cmd *cobra.Command, args []string) error {
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, args[0])
	if err != nil {
		return err
	}

	// Only send the fields the user explicitly set, so unset flags leave the
	// corresponding values unchanged (the server treats an omitted field as
	// "no change").
	payload := map[string]interface{}{}
	if cmd.Flags().Changed("description") {
		payload["description"] = o.description
	}
	if cmd.Flags().Changed("privilege") {
		payload["privilege"] = o.privilege
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPatch,
		URL:     url,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("service account %s was updated", args[0])
	}
	return err
}
