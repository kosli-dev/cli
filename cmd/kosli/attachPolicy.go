package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type AttachPolicyPayload struct {
	PolicyNames []string `json:"policy_names"`
}

type attachPolicyOptions struct {
	payload      AttachPolicyPayload
	environments []string
}

const attachPolicyShortDesc = `Attach a policy to one or more Kosli environments.  `

const attachPolicyExample = `
# attach previously created policy to multiple environment:
kosli attach-policy yourPolicyName \
	--environment yourFirstEnvironmentName \
	--environment yourSecondEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newAttachPolicyCmd(out io.Writer) *cobra.Command {
	o := new(attachPolicyOptions)
	cmd := &cobra.Command{
		Use:     "attach-policy POLICY-NAME",
		Short:   attachPolicyShortDesc,
		Long:    attachPolicyShortDesc,
		Example: attachPolicyExample,
		Hidden:  true,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringSliceVarP(&o.environments, "environment", "e", []string{}, attachPolicyEnvFlag)

	err := RequireFlags(cmd, []string{"environment"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *attachPolicyOptions) run(args []string) error {
	var err error
	for _, env := range o.environments {
		url := fmt.Sprintf("%s/api/v2/environments/%s/%s/policies", global.Host, global.Org, env)
		o.payload.PolicyNames = []string{args[0]}
		reqParams := &requests.RequestParams{
			Method:   http.MethodPost,
			URL:      url,
			Payload:  o.payload,
			DryRun:   global.DryRun,
			Password: global.ApiToken,
		}
		_, err = kosliClient.Do(reqParams)
		if err != nil {
			break
		}
	}
	if err == nil && !global.DryRun {
		logger.Info("policy '%s' is attached to environments: %s", args[0], o.environments)
	}
	return err
}
