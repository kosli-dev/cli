package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/kosli-dev/cli/internal/kosli"
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

	scalingDeprecationMsg := "this flag is deprecated and will be removed in a future version. Scaling events do not trigger new snapshots."
	for _, name := range []string{"include-scaling", "exclude-scaling"} {
		if err := cmd.PersistentFlags().MarkDeprecated(name, scalingDeprecationMsg); err != nil {
			logger.Error("failed to mark %s as deprecated: %v", name, err)
		}
	}

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
func ensureEnvironment(ctx context.Context, envName, envType string) error {
	o := snapshotAutoEnv

	// Enforce the scaling-flag mutual exclusion unconditionally, mirroring
	// `kosli create environment`, so the behaviour is consistent whether or not
	// --auto-environment is set.
	if o.includeScaling && o.excludeScaling {
		return fmt.Errorf("only one of --include-scaling, --exclude-scaling is allowed")
	}

	optionalFlagsSet := o.description != "" || o.includeScaling || o.excludeScaling

	if !o.autoEnvironment {
		if optionalFlagsSet {
			logger.Warn("--environment-description, --include-scaling and --exclude-scaling are ignored unless --auto-environment is set")
		}
		return nil
	}

	if global.DryRun {
		logger.Info("dry-run: environment %s would be created with type %s if it does not exist", envName, envType)
		return nil
	}

	exists, existingType, err := getEnvironmentTypeIfExists(ctx, envName)
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

	return createEnvironmentForSnapshot(ctx, envName, envType)
}

// getEnvironmentTypeIfExists fetches the environment's metadata and returns
// whether it exists along with its type. Only a genuine HTTP 404 is treated as
// "does not exist" (exists=false, nil error); any other error (auth, network,
// 5xx after retries) is returned to the caller so it is never mistaken for a
// missing environment.
func getEnvironmentTypeIfExists(ctx context.Context, envName string) (bool, string, error) {
	response, err := kosliAPI().GetEnvironment(ctx, envName)
	if err != nil {
		// Only a genuine 404 means the environment does not exist yet. Any other
		// error (auth, network, 5xx after retries) is propagated so we never fall
		// through to a create-or-update PUT that could overwrite an existing
		// environment or bypass the type/logical safety checks.
		if kosli.IsNotFound(err) {
			logger.Debug("environment %s does not exist yet", envName)
			return false, "", nil
		}
		return false, "", err
	}

	environment, err := response.Value()
	if err != nil {
		return false, "", err
	}
	return true, environment.Type, nil
}

// createEnvironmentForSnapshot creates a physical environment of the inferred
// type, applying the optional --environment-description and scaling flags. It
// reuses the same payload and endpoint as `kosli create environment`.
func createEnvironmentForSnapshot(ctx context.Context, envName, envType string) error {
	description := snapshotAutoEnv.description
	if description == "" {
		description = defaultAutoEnvDescription
	}

	payload := kosli.CreateEnvironmentRequest{
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

	if _, err := kosliAPI().CreateEnvironment(ctx, payload); err != nil {
		return fmt.Errorf("failed to auto-create environment %s: %v", envName, err)
	}
	logger.Info("environment %s was created", envName)
	return nil
}
