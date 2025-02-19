package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type detachPolicyOptions struct {
	payload      AttachPolicyPayload
	environments []string
}

const detachPolicyShortDesc = `Detach a policy from one or more Kosli environments.  `

const detachPolicyLongDesc = `If the environment has no more policies attached to it, then its snapshots' status will become "unknown".`

const detachPolicyExample = `
# detach policy from multiple environment:
kosli detach-policy yourPolicyName \
	--environment yourFirstEnvironmentName \
	--environment yourSecondEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newDetachPolicyCmd(out io.Writer) *cobra.Command {
	o := new(detachPolicyOptions)
	cmd := &cobra.Command{
		Use:     "detach-policy POLICY-NAME",
		Short:   detachPolicyShortDesc,
		Long:    detachPolicyLongDesc,
		Example: detachPolicyExample,
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

	cmd.Flags().StringSliceVarP(&o.environments, "environment", "e", []string{}, detachPolicyEnvFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"environment"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *detachPolicyOptions) run(args []string) error {
	var err error
	for _, env := range o.environments {
		url := fmt.Sprintf("%s/api/v2/environments/%s/%s/policies", global.Host, global.Org, env)
		o.payload.PolicyNames = []string{args[0]}
		reqParams := &requests.RequestParams{
			Method:  http.MethodDelete,
			URL:     url,
			Payload: o.payload,
			DryRun:  global.DryRun,
			Token:   global.ApiToken,
		}
		_, err = kosliClient.Do(reqParams)
		if err != nil {
			break
		}
	}
	if err == nil && !global.DryRun {
		logger.Info("policy '%s' is detached from environments: %s", args[0], o.environments)
	}
	return err
}
