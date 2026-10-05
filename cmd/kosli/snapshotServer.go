package main

import (
	"fmt"
	"io"

	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const snapshotServerShortDesc = `Report a snapshot of artifacts running in a server environment to Kosli.  `

const snapshotServerLongDesc = snapshotServerShortDesc + `
You can report directory or file artifacts in one or more server paths.

` + fingerprintDirSynopsis

const snapshotServerExample = `
# report directory artifacts running in a server at a list of paths:
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

type snapshotServerOptions struct {
	paths        []string
	excludePaths []string
}

func newSnapshotServerCmd(out io.Writer) *cobra.Command {
	o := new(snapshotServerOptions)
	cmd := &cobra.Command{
		Use:        "server ENVIRONMENT-NAME",
		Short:      snapshotServerShortDesc,
		Long:       snapshotServerLongDesc,
		Aliases:    []string{"directories"},
		Args:       cobra.ExactArgs(1),
		Deprecated: "use 'kosli snapshot paths' instead",
		Example:    snapshotServerExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			if len(o.paths) == 0 {
				return fmt.Errorf("required flag \"paths\" not set")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringSliceVarP(&o.paths, "paths", "p", []string{}, pathsFlag)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "exclude", "x", []string{}, serverExcludePathsFlag)
	cmd.Flags().StringSliceVarP(&o.excludePaths, "e", "e", []string{}, serverExcludePathsFlag)
	addDryRunFlag(cmd)

	err := DeprecateFlags(cmd, map[string]string{
		"e": "use -x instead",
	})
	if err != nil {
		logger.Error("failed to configure deprecated flags: %v", err)
	}

	err = RequireFlags(cmd, []string{"paths"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotServerOptions) run(args []string) error {
	envName := args[0]
	return newSnapshotReporter().report(envName, "server", func() (any, string, error) {
		artifacts, err := server.CreateServerArtifactsData(o.paths, o.excludePaths, logger)
		if err != nil {
			return nil, "", err
		}
		reported := fmt.Sprintf("[%d] artifacts were reported to environment %s", len(artifacts), envName)
		return &server.ServerEnvRequest{Artifacts: artifacts}, reported, nil
	})
}
