package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const snapshotPathShortDesc = `Report a snapshot of a single artifact running in a specific filesystem path to Kosli.  `

const snapshotPathLongDesc = snapshotPathsShortDesc + `
You can report a directory or file artifact. For reporting multiple artifacts in one go, use "kosli snapshot paths".
You can exclude certain paths or patterns from the artifact fingerprint using ^--exclude^.
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

` + kosliIgnoreDesc

const snapshotPathExample = `
# report one artifact running in a specific path in a filesystem:
kosli snapshot path yourEnvironmentName \
	--path path/to/your/artifact/dir/or/file \
	--name yourArtifactDisplayName \
	--api-token yourAPIToken \
	--org yourOrgName

# report one artifact running in a specific path in a filesystem AND exclude certain path patterns:
kosli snapshot path yourEnvironmentName \
	--path path/to/your/artifact/dir \
	--name yourArtifactDisplayName \
	--exclude **/log,unwanted.txt,path/**/output.txt
	--api-token yourAPIToken \
	--org yourOrgName
`

type snapshotPathOptions struct {
	path         string
	artifactName string
	exclude      []string
}

func newSnapshotPathCmd(out io.Writer) *cobra.Command {
	o := new(snapshotPathOptions)
	cmd := &cobra.Command{
		Use:     "path ENVIRONMENT-NAME",
		Short:   snapshotPathShortDesc,
		Long:    snapshotPathLongDesc,
		Args:    cobra.ExactArgs(1),
		Example: snapshotPathExample,
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

	cmd.Flags().StringVar(&o.path, "path", "", snapshotPathPathFlag)
	cmd.Flags().StringVar(&o.artifactName, "name", "", snapshotPathArtifactNameFlag)
	cmd.Flags().StringSliceVarP(&o.exclude, "exclude", "x", []string{}, snapshotPathExcludeFlag)
	addDryRunFlag(cmd)

	if err := RequireFlags(cmd, []string{"path", "name"}); err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotPathOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/server", global.Host, global.Org, envName)

	// load path spec from flags
	ps := &server.PathsSpec{
		Version: 1,
		Artifacts: map[string]server.ArtifactPathSpec{
			o.artifactName: {
				Path:    o.path,
				Exclude: o.exclude,
			},
		},
	}

	artifacts, err := server.CreatePathsArtifactsData(ps, logger)
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
	if err == nil && global.DryRun == "false" {
		logger.Info("[%d] artifacts were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}
