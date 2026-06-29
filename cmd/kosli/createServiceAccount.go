package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createServiceAccountShortDesc = `Create a service account.`

const createServiceAccountLongDesc = createServiceAccountShortDesc + `

A service account is a non-human identity in your organization. API keys are
created separately for it with ^kosli create api-key^.`

const createServiceAccountExample = `
# create a service account:
kosli create service-account yourServiceAccountName \
	--privilege member \
	--description "CI service account" \
	--api-token yourAPIToken \
	--org yourOrgName
`

type createServiceAccountOptions struct {
	payload createServiceAccountPayload
}

type createServiceAccountPayload struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Privilege   string `json:"privilege"`
}

func newCreateServiceAccountCmd(out io.Writer) *cobra.Command {
	o := new(createServiceAccountOptions)
	cmd := &cobra.Command{
		Use:     "service-account SERVICE-ACCOUNT-NAME",
		Aliases: []string{"sa"},
		Short:   createServiceAccountShortDesc,
		Long:    createServiceAccountLongDesc,
		Example: createServiceAccountExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", serviceAccountDescriptionFlag)
	cmd.Flags().StringVar(&o.payload.Privilege, "privilege", "", serviceAccountPrivilegeFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"privilege"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *createServiceAccountOptions) run(args []string) error {
	o.payload.Name = args[0]
	url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org)
	if err != nil {
		return err
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPost,
		URL:     url,
		Payload: o.payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("service account %s was created", o.payload.Name)
	}
	return err
}
