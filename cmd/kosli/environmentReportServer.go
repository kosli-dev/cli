package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const environmentReportServerShortDesc = `Report artifacts running in a server environment to Kosli.`

const environmentReportServerLongDesc = environmentReportServerShortDesc + `
You can report directory or file artifacts in one or more server paths.`

const environmentReportServerExample = `
# report directory artifacts running in a server at a list of paths:
kosli environment report server yourEnvironmentName \
	--paths a/b/c,e/f/g \
	--api-token yourAPIToken \
	--owner yourOrgName  `

type environmentReportServerOptions struct {
	paths []string
	id    string
}

func newEnvironmentReportServerCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportServerOptions)
	cmd := &cobra.Command{
		Use:     "server ENVIRONMENT-NAME",
		Short:   environmentReportServerShortDesc,
		Long:    environmentReportServerLongDesc,
		Aliases: []string{"directories"},
		Args:    cobra.ExactArgs(1),
		Example: environmentReportServerExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringSliceVarP(&o.paths, "paths", "p", []string{}, pathsFlag)

	err := RequireFlags(cmd, []string{"paths"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportServerOptions) run(args []string) error {
	envName := args[0]
	if o.id == "" {
		o.id = envName
	}

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)

	artifacts, err := server.CreateServerArtifactsData(o.paths, logger)
	if err != nil {
		return err
	}
	payload := &server.ServerEnvRequest{
		Artifacts: artifacts,
		Type:      "server",
		Id:        o.id,
	}

	reqParams := &requests.RequestParams{
		Method:   http.MethodPut,
		URL:      url,
		Payload:  payload,
		DryRun:   global.DryRun,
		Password: global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] artifacts were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err

}
