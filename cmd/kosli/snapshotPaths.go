package main

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const snapshotPathsShortDesc = `Report a snapshot of artifacts running in specific filesystem paths to Kosli.  `

const pathSpecFileDesc = `Paths files can be in YAML, JSON or TOML formats.
They specify a list of artifacts to fingerprint. For each artifact, the file specifies a base path to look for the artifact in 
and (optionally) a list of paths to exclude. Excluded paths are relative to the artifact path(s) and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

` + kosliIgnoreDesc + `

This is an example YAML paths spec file:
` +
	"```yaml\n" +
	`version: 1
artifacts:
  artifact_name_a:
    path: dir1
    exclude: [subdir1, **/log]` +
	"\n```"

const snapshotPathsLongDesc = snapshotPathsShortDesc + `
You can report directory or file artifacts in one or more filesystem paths. 
Artifacts names and the paths to include and exclude when fingerprinting them can be 
defined in a paths file which can be provided using ^--paths-file^.

` + pathSpecFileDesc

const snapshotPathsExample = `
# report one or more artifacts running in a filesystem using a path spec file:
kosli snapshot paths yourEnvironmentName \
	--paths-file path/to/your/paths/file \
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

	cmd.Flags().StringVar(&o.pathSpecFile, "paths-file", "", pathsSpecFileFlag)
	addDryRunFlag(cmd)

	if err := RequireFlags(cmd, []string{"paths-file"}); err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotPathsOptions) run(args []string) error {
	envName := args[0]

	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/server", global.Host, global.Org, envName)

	// load path spec from file
	ps, err := processPathSpecFile(o.pathSpecFile)
	if err != nil {
		return err
	}

	artifacts, err := server.CreatePathsArtifactsData(ps, logger)
	if err != nil {
		return err
	}
	payload := &server.ServerEnvRequest{
		Artifacts: artifacts,
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     url,
		Payload: payload,
		DryRun:  global.DryRun,
		Token:   global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("[%d] artifacts were reported to environment %s", len(payload.Artifacts), envName)
	}
	return err
}

func processPathSpecFile(pathsSpecFile string) (*server.PathsSpec, error) {
	var ps *server.PathsSpec
	v := viper.New()
	dir, file := filepath.Split(pathsSpecFile)
	file = strings.TrimSuffix(file, filepath.Ext(file))

	// Set the base name of the pathspec file, without the file extension.
	v.SetConfigName(file)

	// Set the dir path where viper should look for the
	// pathspec file. By default, we are looking in the current working directory.
	if dir == "" {
		dir = "."
	}
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		return ps, fmt.Errorf("failed to parse path spec file [%s] : %v", pathsSpecFile, err)
	}

	if err := v.UnmarshalExact(&ps); err != nil {
		return ps, fmt.Errorf("failed to unmarshal path spec file [%s] : %v", pathsSpecFile, err)
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(ps); err != nil {
		return ps, fmt.Errorf("path spec file [%s] is invalid: %v", pathsSpecFile, err)
	}

	return ps, nil
}
