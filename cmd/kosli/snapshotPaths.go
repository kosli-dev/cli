package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const snapshotPathsShortDesc = `Report a snapshot of artifacts running from specific filesystem paths to Kosli.  `

const pathSpecFileDesc = `Paths spec files can be in YAML, JSON or TOML formats.
They specify a list of artifacts to fingerprint. For each artifact, the file specifies a base path to look for the artifact in 
and (optionally) a list of paths to ignore. Ignored paths are relative to the artifact path(s) and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

This is an example YAML paths spec file:

version: 1
artifacts:
  artifact_name_a:
    path: dir1
    ignore: [subdir1, **/log]`

const snapshotPathsLongDesc = snapshotPathsShortDesc + `
You can report directory or file artifacts in one or more filesystem paths. 
Artifacts names and the paths to include and ignore when fingerprinting them can be defined in a paths spec file
which can be provided using ^--path-spec^.

` + pathSpecFileDesc

const snapshotPathsExample = `
# report directory artifacts running in a filesystem at a list of paths:
kosli snapshot server yourEnvironmentName \
	--path-spec path/to/your/pathsSpec/file \
	--api-token yourAPIToken \
	--org yourOrgName  
`

type snapshotPathsOptions struct {
	pathSpecFile string
}

func newSnapshotPathsCmd(out io.Writer) *cobra.Command {
	o := new(snapshotPathsOptions)
	cmd := &cobra.Command{
		Use:     "paths ENVIRONMENT-NAME",
		Short:   snapshotPathsShortDesc,
		Long:    snapshotPathsLongDesc,
		Args:    cobra.ExactArgs(1),
		Example: snapshotPathsExample,
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

	cmd.Flags().StringVar(&o.pathSpecFile, "path-spec", "", "path to the path-spec file")
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"path-spec"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotPathsOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/server", global.Host, global.Org, envName)

	artifacts, err := server.CreatePathsArtifactsData(o.pathSpecFile, logger)
	if err != nil {
		return err
	}
	payload := &server.ServerEnvRequest{
		Artifacts: artifacts,
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
