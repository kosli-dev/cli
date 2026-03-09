package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotK8STestSuite struct {
	suite.Suite
	defaultKosliArguments string
	envName               string
}

func (suite *SnapshotK8STestSuite) SetupTest() {
	suite.envName = "snapshot-k8s-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateEnv(global.Org, suite.envName, "K8S", suite.T())
}

func (suite *SnapshotK8STestSuite) TestSnapshotK8SCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot K8S fails if both --namespaces and --exclude-namespaces are set",
			cmd:       fmt.Sprintf(`snapshot k8s %s --namespaces default --exclude-namespaces default %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: only one of --namespaces, --exclude-namespaces is allowed\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if no args and no --config-file",
			cmd:       fmt.Sprintf(`snapshot k8s %s`, suite.defaultKosliArguments),
			golden:    "Error: requires either a positional environment name argument or --config-file\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if --config-file and positional arg are both provided",
			cmd:       fmt.Sprintf(`snapshot k8s %s --config-file testdata/k8s-config/valid-single.yaml %s`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: cannot use '--config-file' together with a positional environment name argument\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if --config-file and --namespaces are both provided",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/valid-single.yaml --namespaces default %s`, suite.defaultKosliArguments),
			golden:    "Error: cannot use '--config-file' together with '--namespaces'\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if --config-file and --exclude-namespaces are both provided",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/valid-single.yaml --exclude-namespaces default %s`, suite.defaultKosliArguments),
			golden:    "Error: cannot use '--config-file' together with '--exclude-namespaces'\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file not found",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file /nonexistent/path.yaml %s`, suite.defaultKosliArguments),
			golden:    "Error: failed to read config file '/nonexistent/path.yaml': open /nonexistent/path.yaml: no such file or directory\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file has invalid YAML",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/invalid-yaml.yaml %s`, suite.defaultKosliArguments),
			goldenRegex: "Error: failed to parse config file.*",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file has empty environments list",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/empty-environments.yaml %s`, suite.defaultKosliArguments),
			golden:    "Error: invalid config: 'environments' list must contain at least one entry\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file has entry missing name",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/missing-name.yaml %s`, suite.defaultKosliArguments),
			golden:    "Error: invalid config: environment entry 1 is missing required field 'name'\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file has duplicate environment names",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/duplicate-names.yaml %s`, suite.defaultKosliArguments),
			golden:    "Error: invalid config: duplicate environment name 'prod-env'\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file has conflicting filters",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/conflicting-filters.yaml %s`, suite.defaultKosliArguments),
			golden:    "Error: invalid config for environment 'bad-env': cannot combine 'namespaces' with 'excludeNamespaces'\n",
		},
		{
			wantError: true,
			name:      "snapshot K8S fails if config file has invalid regex",
			cmd:       fmt.Sprintf(`snapshot k8s --config-file testdata/k8s-config/invalid-regex.yaml %s`, suite.defaultKosliArguments),
			goldenRegex: `Error: invalid config for environment 'bad-regex-env': invalid regex '\[invalid'.*`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestParseK8SSnapshotConfig(t *testing.T) {
	t.Run("valid single environment config", func(t *testing.T) {
		config, err := parseK8SSnapshotConfig("testdata/k8s-config/valid-single.yaml")
		require.NoError(t, err)
		require.Len(t, config.Environments, 1)
		assert.Equal(t, "prod-env", config.Environments[0].Name)
		assert.Equal(t, []string{"prod-ns1", "prod-ns2"}, config.Environments[0].Namespaces)
	})

	t.Run("valid multi-environment config", func(t *testing.T) {
		config, err := parseK8SSnapshotConfig("testdata/k8s-config/valid-multi.yaml")
		require.NoError(t, err)
		require.Len(t, config.Environments, 3)
		assert.Equal(t, "prod-env", config.Environments[0].Name)
		assert.Equal(t, "staging-env", config.Environments[1].Name)
		assert.Equal(t, "infra-env", config.Environments[2].Name)
		assert.Equal(t, []string{"^staging-.*"}, config.Environments[1].NamespacesRegex)
		assert.Equal(t, []string{"prod-ns1", "prod-ns2", "default"}, config.Environments[2].ExcludeNamespaces)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("/nonexistent/path.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("invalid YAML", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("testdata/k8s-config/invalid-yaml.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse config file")
	})

	t.Run("empty environments list", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("testdata/k8s-config/empty-environments.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "'environments' list must contain at least one entry")
	})

	t.Run("missing environment name", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("testdata/k8s-config/missing-name.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "environment entry 1 is missing required field 'name'")
	})

	t.Run("duplicate environment names", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("testdata/k8s-config/duplicate-names.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate environment name 'prod-env'")
	})

	t.Run("conflicting filters", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("testdata/k8s-config/conflicting-filters.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot combine 'namespaces' with 'excludeNamespaces'")
	})

	t.Run("invalid regex", func(t *testing.T) {
		_, err := parseK8SSnapshotConfig("testdata/k8s-config/invalid-regex.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid regex")
	})
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotK8STestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotK8STestSuite))
}
