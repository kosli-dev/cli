package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SnapshotAzureAppsTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	defaultAzureArguments string
	envName               string
}

func (suite *SnapshotAzureAppsTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_CLIENT_SECRET"})

	suite.envName = "snapshot-azure-env"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	// AZURE-CLIENT-SECRET is set as a secret and passed as env variable to tests
	suite.defaultAzureArguments = " --azure-client-id 0bfad7d5-eb5e-4144-95a0-0a0d66eb07cb --azure-tenant-id e52b5fba-43c2-4eaf-91c1-579dc6fae771 " +
		"--azure-subscription-id 96cdee58-1fa8-419d-a65a-7233b3465632 --azure-resource-group-name EnvironmentReportingExperiment"
	CreateEnv(global.Org, suite.envName, "azure-apps", suite.T())
}

func (suite *SnapshotAzureAppsTestSuite) TestSnapshotAzureAppsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "snapshot azure fails if 2 args are provided",
			cmd:       fmt.Sprintf(`snapshot azure %s xxx %s %s`, suite.envName, suite.defaultKosliArguments, suite.defaultAzureArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "snapshot azure fails if no args are set",
			cmd:       fmt.Sprintf(`snapshot azure %s %s`, suite.defaultKosliArguments, suite.defaultAzureArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "snapshot azure fails if --digests-source flag is set to invalid value",
			cmd:       fmt.Sprintf(`snapshot azure %s %s %s --digests-source ghcr `, suite.envName, suite.defaultKosliArguments, suite.defaultAzureArguments),
			golden:    "Error: invalid value for --digests-source flag. Valid values are 'acr' and 'logs'\n",
		},
		{
			name: "snapshot azure succeeds if all required flags are set",
			cmd:  fmt.Sprintf(`snapshot azure %s %s %s`, suite.envName, suite.defaultKosliArguments, suite.defaultAzureArguments),
		},
		{
			name: "snapshot azure succeeds when digests-source is set to acr if all required flags are set",
			cmd:  fmt.Sprintf(`snapshot azure %s %s %s --digests-source acr`, suite.envName, suite.defaultKosliArguments, suite.defaultAzureArguments),
		},
		{
			name: "snapshot azure succeeds when digests-source is set to logs if all required flags are set",
			cmd:  fmt.Sprintf(`snapshot azure %s %s %s --digests-source logs`, suite.envName, suite.defaultKosliArguments, suite.defaultAzureArguments),
		},

		{
			wantError: true,
			name:      "snapshot azure fails when Azure client ID is not set",
			cmd:       fmt.Sprintf(`snapshot azure %s %s --azure-client-secret xxx --azure-tenant-id xxx --azure-subscription-id xxx --azure-resource-group-name xxx`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"azure-client-id\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot azure fails when Azure tenant ID is not set",
			cmd:       fmt.Sprintf(`snapshot azure %s %s --azure-client-id xxx --azure-client-secret xxx --azure-subscription-id xxx --azure-resource-group-name xxx`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"azure-tenant-id\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot azure fails when Azure subscription ID is not set",
			cmd:       fmt.Sprintf(`snapshot azure %s %s --azure-client-id xxx --azure-client-secret xxx --azure-tenant-id xxx --azure-resource-group-name xxx`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"azure-subscription-id\" not set\n",
		},
		{
			wantError: true,
			name:      "snapshot azure fails when Azure resource group name is not set",
			cmd:       fmt.Sprintf(`snapshot azure %s %s --azure-client-id xxx --azure-client-secret xxx --azure-tenant-id xxx --azure-subscription-id xxx`, suite.envName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"azure-resource-group-name\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSnapshotAzureAppsTestSuite(t *testing.T) {
	suite.Run(t, new(SnapshotAzureAppsTestSuite))
}
