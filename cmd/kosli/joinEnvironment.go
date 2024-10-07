package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const joinEnvironmentShortDesc = `Join a physical environment to a logical environment.`

const joinEnvironmentLongDesc = joinEnvironmentShortDesc + ``

const joinEnvironmentExample = `
# join a physical environment to a logical environment:
kosli join environment \
	--physical prod-k8 \
	--logical prod \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type joinEnvironmentOptions struct {
	payload JoinEnvironmentPayload
	logical string
}

type JoinEnvironmentPayload struct {
	Physical string `json:"physical_env_name"`
}

func newJoinEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(joinEnvironmentOptions)
	cmd := &cobra.Command{
		Use:     "environment",
		Aliases: []string{"env"},
		Short:   joinEnvironmentShortDesc,
		Long:    joinEnvironmentLongDesc,
		Example: joinEnvironmentExample,
		Args:    cobra.ExactArgs(0),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			url := fmt.Sprintf("%s/api/v2/environments/%s/%s/join", global.Host, global.Org, o.logical)

			reqParams := &requests.RequestParams{
				Method:   http.MethodPut,
				URL:      url,
				Payload:  o.payload,
				DryRun:   global.DryRun,
				Password: global.ApiToken,
			}
			_, err := kosliClient.Do(reqParams)
			if err == nil && global.DryRun == "false" {
				logger.Info("environment '%s' was joined to '%s'", o.payload.Physical, o.logical)
			}
			return err
		},
	}

	cmd.Flags().StringVar(&o.payload.Physical, "physical", "", physicalEnvFlag)
	cmd.Flags().StringVar(&o.logical, "logical", "", logicalEnvFlag)
	err := RequireFlags(cmd, []string{"physical", "logical"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	addDryRunFlag(cmd)
	return cmd
}
