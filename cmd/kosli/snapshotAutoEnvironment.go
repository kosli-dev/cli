package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// autoEnvOptions holds the values of the group-wide auto-environment flags that
// are shared across all `kosli snapshot` subcommands.
type autoEnvOptions struct {
	autoEnvironment bool
	description     string
	includeScaling  bool
	excludeScaling  bool
}

// snapshotAutoEnv is the shared destination for the auto-environment flags.
// The flags are registered once as persistent flags on the parent `snapshot`
// command, mirroring how --dry-run writes to the package-level `global`.
var snapshotAutoEnv = &autoEnvOptions{}

const defaultAutoEnvDescription = "Auto-created by kosli snapshot"

// addAutoEnvironmentFlags registers the auto-environment flags on the snapshot
// command group. It is called on the parent `snapshot` command so that every
// subcommand inherits the flags. `--auto-env` is accepted as an alias for
// `--auto-environment`.
func addAutoEnvironmentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVarP(&snapshotAutoEnv.autoEnvironment, "auto-environment", "A", false, autoEnvironmentFlag)
	cmd.PersistentFlags().StringVar(&snapshotAutoEnv.description, "environment-description", "", envDescriptionFlag)
	cmd.PersistentFlags().BoolVar(&snapshotAutoEnv.includeScaling, "include-scaling", false, includeScalingFlag)
	cmd.PersistentFlags().BoolVar(&snapshotAutoEnv.excludeScaling, "exclude-scaling", false, excludeScalingFlag)

	// Accept --auto-env as an alias for --auto-environment. SetGlobalNormalizationFunc
	// propagates to all subcommands of the snapshot group.
	cmd.SetGlobalNormalizationFunc(func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		if name == "auto-env" {
			name = "auto-environment"
		}
		return pflag.NormalizedName(name)
	})
}

// ensureEnvironment auto-creates the target environment before a snapshot is
// reported, when --auto-environment is set. It is a no-op when the flag is not
// set or when the environment already exists with a matching type. envType is
// inferred from the snapshot subcommand (e.g. "docker", "K8S").
func ensureEnvironment(envName, envType string) error {
	o := snapshotAutoEnv
	optionalFlagsSet := o.description != "" || o.includeScaling || o.excludeScaling

	if !o.autoEnvironment {
		if optionalFlagsSet {
			logger.Warn("--environment-description, --include-scaling and --exclude-scaling are ignored unless --auto-environment is set")
		}
		return nil
	}

	if o.includeScaling && o.excludeScaling {
		return fmt.Errorf("only one of --include-scaling, --exclude-scaling is allowed")
	}

	if global.DryRun {
		logger.Info("dry-run: environment %s would be created with type %s if it does not exist", envName, envType)
		return nil
	}

	exists, existingType, err := getEnvironmentTypeIfExists(envName)
	if err != nil {
		return err
	}
	if exists {
		if strings.EqualFold(existingType, "logical") {
			return fmt.Errorf("cannot report a snapshot to the logical environment %s", envName)
		}
		if !strings.EqualFold(existingType, envType) {
			return fmt.Errorf("environment %s already exists with type %s, which does not match the snapshot type %s", envName, existingType, envType)
		}
		if optionalFlagsSet {
			logger.Warn("environment %s already exists; --environment-description, --include-scaling and --exclude-scaling are ignored", envName)
		}
		return nil
	}

	return createEnvironmentForSnapshot(envName, envType)
}

// getEnvironmentTypeIfExists fetches the environment's metadata and returns
// whether it exists along with its type. A failed GET (most commonly a 404 for
// an environment that has not been created yet) is treated as "does not exist";
// any genuine error (auth, network) resurfaces from the subsequent create or
// report request.
func getEnvironmentTypeIfExists(envName string) (bool, string, error) {
	reqURL, err := url.JoinPath(global.Host, "api/v2/environments", global.Org, envName)
	if err != nil {
		return false, "", err
	}

	reqParams := &requests.RequestParams{
		Method: http.MethodGet,
		URL:    reqURL,
		Token:  global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		logger.Debug("could not fetch environment %s, assuming it does not exist: %v", envName, err)
		return false, "", nil
	}

	var env map[string]interface{}
	if err := json.Unmarshal([]byte(response.Body), &env); err != nil {
		return false, "", err
	}
	envType, _ := env["type"].(string)
	return true, envType, nil
}

// createEnvironmentForSnapshot creates a physical environment of the inferred
// type, applying the optional --environment-description and scaling flags. It
// reuses the same payload and endpoint as `kosli create environment`.
func createEnvironmentForSnapshot(envName, envType string) error {
	reqURL, err := url.JoinPath(global.Host, "api/v2/environments", global.Org)
	if err != nil {
		return err
	}

	description := snapshotAutoEnv.description
	if description == "" {
		description = defaultAutoEnvDescription
	}

	payload := CreateEnvironmentPayload{
		Name:        envName,
		Type:        envType,
		Description: description,
	}
	if snapshotAutoEnv.includeScaling {
		myTrue := true
		payload.IncludeScaling = &myTrue
	}
	if snapshotAutoEnv.excludeScaling {
		myFalse := false
		payload.IncludeScaling = &myFalse
	}

	reqParams := &requests.RequestParams{
		Method:  http.MethodPut,
		URL:     reqURL,
		Payload: payload,
		Token:   global.ApiToken,
	}
	if _, err := kosliClient.Do(reqParams); err != nil {
		return fmt.Errorf("failed to auto-create environment %s: %v", envName, err)
	}
	logger.Info("environment %s was created", envName)
	return nil
}
