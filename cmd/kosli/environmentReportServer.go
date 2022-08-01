package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const environmentReportServerDesc = `
List the artifacts deployed in a server environment and their digests 
and report them to Kosli. 
`

const environmentReportServerExample = `
# report directory artifacts running in a server at a list of paths:
kosli environment report server yourEnvironmentName \
	--paths a/b/c, e/f/g \
	--api-token yourAPIToken \
	--owner yourOrgName  
`

type environmentReportServerOptions struct {
	paths []string
	id    string
}

func newEnvironmentReportServerCmd(out io.Writer) *cobra.Command {
	o := new(environmentReportServerOptions)
	cmd := &cobra.Command{
		Use:     "server ENVIRONMENT-NAME",
		Short:   "Report directory or file artifacts data in the given list of paths to Kosli.",
		Long:    environmentReportServerDesc,
		Aliases: []string{"directories"},
		Example: environmentReportServerExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return ErrorBeforePrintingUsage(cmd, "only env-name argument is allowed")
			}
			if len(args) == 0 || args[0] == "" {
				return ErrorBeforePrintingUsage(cmd, "env-name argument is required")
			}

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
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *environmentReportServerOptions) run(args []string) error {
	envName := args[0]
	if o.id == "" {
		o.id = envName
	}

	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, envName)

	artifacts, err := server.CreateServerArtifactsData(o.paths, log)
	if err != nil {
		return err
	}
	requestBody := &server.ServerEnvRequest{
		Artifacts: artifacts,
		Type:      "server",
		Id:        o.id,
	}

	_, err = requests.SendPayload(requestBody, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
	return err

}
