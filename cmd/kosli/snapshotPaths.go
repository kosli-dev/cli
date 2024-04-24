package main

import (
	"io"

	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const snapshotPathsShortDesc = `Report a snapshot of artifacts running from specific filesystem paths to Kosli.  `

const snapshotPathsLongDesc = snapshotPathsShortDesc + `
You can report directory or file artifacts in one or more filesystem paths.

` + fingerprintDirSynopsis

const snapshotPathsExample = `
# report directory artifacts running in a filesystem at a list of paths:
kosli snapshot server yourEnvironmentName \
	--paths a/b/c,e/f/g \
	--api-token yourAPIToken \
	--org yourOrgName  
	
# exclude certain paths when reporting directory artifacts: 
# in the example below, any path matching [a/b/c/logs, a/b/c/*/logs, a/b/c/*/*/logs]
# will be skipped when calculating the fingerprint
kosli snapshot server yourEnvironmentName \
	--paths a/b/c \
	--exclude logs,"*/logs","*/*/logs"
	--api-token yourAPIToken \
	--org yourOrgName 
	
# use glob pattern to match paths to report them as directory artifacts: 
# in the example below, any path matching "*/*/src" under top-dir/ will be reported as a separate artifact.
kosli snapshot server yourEnvironmentName \
	--paths "top-dir/*/*/src" \
	--api-token yourAPIToken \
	--org yourOrgName 
`

type snapshotPathsOptions struct {
	// paths        []string
	// excludePaths []string
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

	// cmd.Flags().StringSliceVarP(&o.paths, "paths", "p", []string{}, pathsFlag)
	// cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "x", []string{}, serverExcludePathsFlag)
	// cmd.Flags().StringSliceVarP(&o.excludePaths, "e", "e", []string{}, serverExcludePathsFlag)
	cmd.Flags().StringVar(&o.pathSpecFile, "path-spec", "", "path to the path-spec file")
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"path-spec"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotPathsOptions) run(args []string) error {
	// envName := args[0]

	// url = fmt.Sprintf("%s/api/v2/environments/%s/%s/report/server", global.Host, global.Org, envName)

	_, err := server.CreatePathsArtifactsData(o.pathSpecFile, logger)
	if err != nil {
		return err
	}
	// payload := &server.ServerEnvRequest{
	// 	Artifacts: artifacts,
	// }

	// reqParams := &requests.RequestParams{
	// 	Method:   http.MethodPut,
	// 	URL:      url,
	// 	Payload:  payload,
	// 	DryRun:   global.DryRun,
	// 	Password: global.ApiToken,
	// }
	// _, err = kosliClient.Do(reqParams)
	// if err == nil && !global.DryRun {
	// 	logger.Info("[%d] artifacts were reported to environment %s", len(payload.Artifacts), envName)
	// }
	// return err
	return nil
}
