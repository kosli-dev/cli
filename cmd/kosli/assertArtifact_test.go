package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AssertArtifactCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	envName               string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

func (suite *AssertArtifactCommandTestSuite) SetupTest() {
	suite.flowName = "assert-artifact"
	suite.envName = "assert-artifact-environment"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.Suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	CreateEnv(global.Org, suite.envName, "server", suite.Suite.T())
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.Suite.T(), err)
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.Suite.T())
}

func (suite *AssertArtifactCommandTestSuite) TestAssertArtifactCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing --org fails",
			cmd:       fmt.Sprintf(`assert artifact --fingerprint 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c  --flow %s --api-token secret`, suite.flowName),
			golden:    "Error: --org is not set\nUsage: kosli assert artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "asserting a non existing artifact fails",
			cmd:       fmt.Sprintf(`assert artifact --fingerprint 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c  --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c' does not exist in flow 'assert-artifact' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:        "asserting an existing compliant artifact (using --fingerprint) results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact --fingerprint %s --flow %s %s`, suite.fingerprint, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*Attestation-name.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact/artifacts/fcf33337634c2577a5d86fd7ecb0a25a7c1bb5d89c14fd236f546a5759252c02(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name:        "asserting an existing compliant artifact (using --fingerprint) for an environment results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact --fingerprint %s --flow %s --environment %s %s`, suite.fingerprint, suite.flowName, suite.envName, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*Attestation-name.*Policy-name.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact/artifacts/fcf33337634c2577a5d86fd7ecb0a25a7c1bb5d89c14fd236f546a5759252c02(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name:        "asserting an existing compliant artifact (using --artifact-type) results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact %s --artifact-type file --flow %s %s`, suite.artifactPath, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact/artifacts/fcf33337634c2577a5d86fd7ecb0a25a7c1bb5d89c14fd236f546a5759252c02?(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name: "json output of asserting an existing compliant artifact (using --artifact-type) results in OK and zero exit",
			cmd:  fmt.Sprintf(`assert artifact %s --output json --artifact-type file --flow %s %s`, suite.artifactPath, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "flow"},
			},
		},
		{
			wantError: true,
			name:      "not providing --fingerprint nor --artifact-type fails",
			cmd:       fmt.Sprintf(`assert artifact --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: docker image name or file/dir path is required when --fingerprint is not provided\nUsage: kosli assert artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "not providing --flow when the artifact has multiple instances fails with a server error",
			cmd:       fmt.Sprintf(`assert artifact --fingerprint %s  %s`, suite.fingerprint, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Fingerprint '%s' matched multiple artifacts in org '%s'. Please narrow the search to one flow.\n", suite.fingerprint, global.Org),
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertArtifactCommandTestSuite))
}
