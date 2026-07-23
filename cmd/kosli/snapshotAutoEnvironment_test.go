package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// SnapshotAutoEnvironmentTestSuite exercises the group-wide --auto-environment
// flag shared across all `kosli snapshot` subcommands. The `snapshot paths`
// subcommand is used as a representative because its output is deterministic.
type SnapshotAutoEnvironmentTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	pathsFileArg          string
	// pre-existing environments
	existingServerEnv string
	existingDockerEnv string
	logicalEnv        string
	// unique per-process names for the auto-creation cases
	newEnv    string
	aliasEnv  string
	dryRunEnv string
}

func (suite *SnapshotAutoEnvironmentTestSuite) SetupSuite() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	suite.pathsFileArg = "--paths-file testdata/paths-files/valid-pathsfile.yml"

	// Unique-per-process names so re-running against a long-lived server still
	// exercises the "does not exist yet" creation path.
	pid := os.Getpid()
	suite.existingServerEnv = fmt.Sprintf("autoenv-existing-server-%d", pid)
	suite.existingDockerEnv = fmt.Sprintf("autoenv-existing-docker-%d", pid)
	suite.logicalEnv = fmt.Sprintf("autoenv-logical-%d", pid)
	suite.newEnv = fmt.Sprintf("autoenv-new-%d", pid)
	suite.aliasEnv = fmt.Sprintf("autoenv-alias-%d", pid)
	suite.dryRunEnv = fmt.Sprintf("autoenv-dryrun-%d", pid)

	CreateEnv(global.Org, suite.existingServerEnv, "server", suite.T())
	CreateEnv(global.Org, suite.existingDockerEnv, "docker", suite.T())
	suite.createLogicalEnv(suite.logicalEnv, []string{suite.existingServerEnv})
}

func (suite *SnapshotAutoEnvironmentTestSuite) createLogicalEnv(name string, included []string) {
	suite.T().Helper()
	o := &createEnvOptions{
		payload: CreateEnvironmentPayload{
			Name:                 name,
			Type:                 "logical",
			Description:          "test logical env",
			IncludedEnvironments: included,
		},
	}
	require.NoError(suite.T(), o.run([]string{name}), "logical env should be created without error")
}

func (suite *SnapshotAutoEnvironmentTestSuite) TestSnapshotAutoEnvironment() {
	tests := []cmdTestCase{
		{
			name:   "auto-creates the environment with the inferred type when it does not exist",
			cmd:    fmt.Sprintf(`snapshot paths %s --auto-environment %s %s`, suite.newEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden: fmt.Sprintf("environment %s was created\n[1] artifacts were reported to environment %s\n", suite.newEnv, suite.newEnv),
		},
		{
			name:   "--auto-env alias auto-creates the environment",
			cmd:    fmt.Sprintf(`snapshot paths %s --auto-env %s %s`, suite.aliasEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden: fmt.Sprintf("environment %s was created\n[1] artifacts were reported to environment %s\n", suite.aliasEnv, suite.aliasEnv),
		},
		{
			name:   "is idempotent (no-op) when the environment already exists",
			cmd:    fmt.Sprintf(`snapshot paths %s --auto-environment %s %s`, suite.existingServerEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[1] artifacts were reported to environment %s\n", suite.existingServerEnv),
		},
		{
			wantError: true,
			name:      "errors when the existing environment has a different type",
			cmd:       fmt.Sprintf(`snapshot paths %s --auto-environment %s %s`, suite.existingDockerEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: environment %s already exists with type docker, which does not match the snapshot type server\n", suite.existingDockerEnv),
		},
		{
			wantError: true,
			name:      "errors when targeting a logical environment",
			cmd:       fmt.Sprintf(`snapshot paths %s --auto-environment %s %s`, suite.logicalEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: cannot report a snapshot to the logical environment %s\n", suite.logicalEnv),
		},
		{
			wantError: true,
			name:      "infers the type from the subcommand (docker) for the mismatch check",
			cmd:       fmt.Sprintf(`snapshot docker %s --auto-environment %s`, suite.existingServerEnv, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: environment %s already exists with type server, which does not match the snapshot type docker\n", suite.existingServerEnv),
		},
		{
			wantError: true,
			name:      "errors when both scaling flags are set",
			cmd:       fmt.Sprintf(`snapshot paths %s --auto-environment --include-scaling --exclude-scaling %s %s`, suite.newEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden:    "Error: only one of --include-scaling, --exclude-scaling is allowed\n",
		},
		{
			wantError: true,
			name:      "errors when both scaling flags are set even without --auto-environment",
			cmd:       fmt.Sprintf(`snapshot paths %s --include-scaling --exclude-scaling %s %s`, suite.existingServerEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden:    "Error: only one of --include-scaling, --exclude-scaling is allowed\n",
		},
		{
			name:   "warns and ignores optional flags when --auto-environment is not set",
			cmd:    fmt.Sprintf(`snapshot paths %s --environment-description "ignored" %s %s`, suite.existingServerEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[warning] --environment-description, --include-scaling and --exclude-scaling are ignored unless --auto-environment is set\n[1] artifacts were reported to environment %s\n", suite.existingServerEnv),
		},
		{
			name:   "warns that optional flags are ignored when the environment already exists",
			cmd:    fmt.Sprintf(`snapshot paths %s --auto-environment --environment-description "ignored" %s %s`, suite.existingServerEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			golden: fmt.Sprintf("[warning] environment %s already exists; --environment-description, --include-scaling and --exclude-scaling are ignored\n[1] artifacts were reported to environment %s\n", suite.existingServerEnv, suite.existingServerEnv),
		},
		{
			name:        "dry-run reports what would be created without creating anything",
			cmd:         fmt.Sprintf(`snapshot paths %s --auto-environment --dry-run %s %s`, suite.dryRunEnv, suite.pathsFileArg, suite.defaultKosliArguments),
			goldenRegex: fmt.Sprintf("dry-run: environment %s would be created with type server if it does not exist", suite.dryRunEnv),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotAutoEnvironmentTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotAutoEnvironmentTestSuite))
}
