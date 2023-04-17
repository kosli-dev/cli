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
type ArtifactEvidencePRAzureCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	flowName              string
}

func (suite *ArtifactEvidencePRAzureCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_AZURE_TOKEN"})

	suite.flowName = "azure-pr"
	suite.artifactFingerprint = "847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	CreateArtifact(suite.flowName, suite.artifactFingerprint, "foobar", suite.T())
}

func (suite *ArtifactEvidencePRAzureCommandTestSuite) TestArtifactEvidencePRAzureCmd() {
	tests := []cmdTestCase{
		{
			name: "report Azure PR evidence works with new flags (fingerprint, name ...)",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "azure pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --org is missing",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a --api-token foo --host bar`,
			golden: "Error: --org is not set\n" +
				"Usage: kosli report evidence artifact pullrequest azure [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --name is missing",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --azure-org-url is missing",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"azure-org-url\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --project is missing",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"project\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --repository is missing",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --commit is missing",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when neither --fingerprint nor --artifact-type are set",
			cmd: `report evidence artifact pullrequest azure artifactNameArg --name az-pr --flow ` + suite.flowName + `
					  --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: either --artifact-type or --fingerprint must be specified\n" +
				"Usage: kosli report evidence artifact pullrequest azure [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			name: "report Azure PR evidence does not fail when commit does not exist, empty evidence is reported instead.",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: fmt.Sprintf("azure pull request evidence is reported to artifact: %s\n", suite.artifactFingerprint),
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --assert is used and commit has no PRs",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
					  --assert
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "Error: no pull requests found for the given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n",
		},
		{
			name: "report Azure PR evidence does not fail when commit has no PRs",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 9bca2c44eaf221a79fb18a1a11bdf2997adaf870` + suite.defaultKosliArguments,
			golden: "no pull requests found for given commit: 9bca2c44eaf221a79fb18a1a11bdf2997adaf870\n" +
				"azure pull request evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when the artifact does not exist in the server",
			cmd: `report evidence artifact pullrequest azure testdata/file1 --artifact-type file --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: Artifact with fingerprint '7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'azure-pr' belonging to organization 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --artifact-type is unsupported",
			cmd: `report evidence artifact pullrequest azure testdata/file1 --artifact-type unsupported --name az-pr --flow ` + suite.flowName + `
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: unsupported is not a supported artifact type\n",
		},
		{
			wantError: true,
			name:      "report Azure PR evidence fails when --user-data is not found",
			cmd: `report evidence artifact pullrequest azure --fingerprint ` + suite.artifactFingerprint + ` --name az-pr --flow ` + suite.flowName + `
					  --user-data non-existing.json
			          --build-url example.com --azure-org-url https://dev.azure.com/kosli --project kosli-azure --repository cli --commit 5f61be8f00a01c84e491922a630c9a418c684c7a` + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidencePRAzureCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidencePRAzureCommandTestSuite))
}
