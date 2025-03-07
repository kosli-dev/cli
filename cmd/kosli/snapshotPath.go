package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/rjeczalik/notify"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/server"
	"github.com/spf13/cobra"
)

const snapshotPathShortDesc = `Report a snapshot of a single artifact running in a specific filesystem path to Kosli.  `

const snapshotPathLongDesc = snapshotPathShortDesc + `
You can report a directory or file artifact. For reporting multiple artifacts in one go, use "kosli snapshot paths".
You can exclude certain paths or patterns from the artifact fingerprint using ^--exclude^.
The supported glob pattern syntax is documented here: https://pkg.go.dev/path/filepath#Match ,
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
	watch        bool
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
	cmd.Flags().BoolVar(&o.watch, "watch", false, pathsWatchFlag)
	addDryRunFlag(cmd)

	if err := RequireFlags(cmd, []string{"path", "name"}); err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *snapshotPathOptions) run(args []string) error {
	envName := args[0]
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

	err := reportArtifacts(ps, envName)
	if err != nil {
		return err
	}

	if o.watch {
		err := watchPath(ps, o.path, envName)
		if err != nil {
			return err
		}
	}

	return nil
}

func reportArtifacts(ps *server.PathsSpec, envName string) error {
	url := fmt.Sprintf("%s/api/v2/environments/%s/%s/report/server", global.Host, global.Org, envName)
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

func watchPath(ps *server.PathsSpec, path, envName string) error {
	events := make(chan notify.EventInfo, 1)
	if err := watchRecursive(path, events); err != nil {
		return fmt.Errorf("error setting up watcher: %v", err)
	}
	defer notify.Stop(events)

	logger.Info("watching for file changes in [%s]. Press Ctrl+C to exit...", path)

	// Handle system interrupts (Ctrl+C)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	for {
		select {
		case event := <-events:
			logger.Info("event detected: %s on %s", event.Event().String(), event.Path())

			err := reportArtifacts(ps, envName)
			if err != nil {
				return err
			}

			// If a new directory is created, start watching it
			// github.com/rjeczalik/notify does not automatically watch new subdirs
			if event.Event() == notify.Create {
				info, err := os.Stat(event.Path())
				if err == nil && info.IsDir() {
					logger.Debug("new directory detected: %s. Adding to watcher.", event.Path())
					err = notify.Watch(event.Path()+"/...", events, notify.All)
					if err != nil {
						return err
					}
				}
			}
		case <-stop:
			logger.Info("\nstopping file watcher for [%s] ...", path)
			return nil
		}
	}
}

// Watches a directory recursively and detects new subdirectories
func watchRecursive(root string, events chan notify.EventInfo) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Watch all directories, including existing and newly created ones
		if info.IsDir() {
			if err := notify.Watch(path+"/...", events, notify.All); err != nil {
				return fmt.Errorf("failed to watch %s: %v", path, err)
			} else {
				logger.Debug("watching: %s", path)
			}
		}
		return nil
	})
}
