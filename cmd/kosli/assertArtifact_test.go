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
	flowName1             string
	flowName2             string
	envName               string
	policyName1           string
	policyName2           string
	artifactName1         string
	artifact1Path         string
	fingerprint1          string
	artifactName2         string
	artifact2Path         string
	fingerprint2          string
}

func (suite *AssertArtifactCommandTestSuite) SetupTest() {
	suite.flowName1 = "assert-artifact-one"
	suite.flowName2 = "assert-artifact-two"
	suite.envName = "assert-artifact-environment"
	suite.policyName1 = "assert-artifact-policy-1"
	suite.policyName2 = "assert-artifact-policy-2"
	suite.artifactName1 = "arti-for-AssertArtifactCommandTestSuite"
	suite.artifact1Path = "testdata/artifacts/AssertArtifactCommandTestSuiteArtifact1.txt"
	suite.artifactName2 = "arti-for-AssertArtifactCommandTestSuite2"
	suite.artifact2Path = "testdata/artifacts/AssertArtifactCommandTestSuiteArtifact2.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName1, suite.Suite.T())
	CreateFlow(suite.flowName2, suite.Suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	CreateEnv(global.Org, suite.envName, "server", suite.Suite.T())
	CreatePolicy(global.Org, suite.policyName1, suite.Suite.T())
	CreatePolicy(global.Org, suite.policyName2, suite.Suite.T())
	var err error
	suite.fingerprint1, err = GetSha256Digest(suite.artifact1Path, fingerprintOptions, logger)
	require.NoError(suite.Suite.T(), err)
	CreateArtifact(suite.flowName1, suite.fingerprint1, suite.artifactName1, suite.Suite.T())
	suite.fingerprint2, err = GetSha256Digest(suite.artifact2Path, fingerprintOptions, logger)
	require.NoError(suite.Suite.T(), err)
	CreateArtifact(suite.flowName1, suite.fingerprint2, suite.artifactName2, suite.Suite.T())
	CreateArtifact(suite.flowName2, suite.fingerprint2, suite.artifactName1, suite.Suite.T())
}

func (suite *AssertArtifactCommandTestSuite) TestAssertArtifactCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "01 missing --org fails",
			cmd:       fmt.Sprintf(`assert artifact --fingerprint 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c  --flow %s --api-token secret`, suite.flowName1),
			golden:    "Error: --org is not set\nUsage: kosli assert artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "02 asserting a non existing artifact fails",
			cmd:       fmt.Sprintf(`assert artifact --fingerprint 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c  --flow %s %s`, suite.flowName1, suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c' does not exist in flow 'assert-artifact-one' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:        "03 asserting a single existing compliant artifact (using --fingerprint) results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact --fingerprint %s %s`, suite.fingerprint1, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*Attestation-name.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact-one/artifacts/0089a849fce9c7c9128cd13a2e8b1c0757bdb6a7bad0fdf2800e38c19055b7fc(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name: "04 json output of asserting a single existing compliant artifact (using --fingerprint) results in OK and zero exit",
			cmd:  fmt.Sprintf(`assert artifact --output json --fingerprint %s %s`, suite.fingerprint1, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "flow"},
				{"flows", "length:1"},
			},
		},
		{
			name:        "05 asserting a single existing compliant artifact (using --fingerprint) for an environment results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact --fingerprint %s --environment %s %s`, suite.fingerprint1, suite.envName, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*Policy-name.*Attestation-name.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact-one/artifacts/0089a849fce9c7c9128cd13a2e8b1c0757bdb6a7bad0fdf2800e38c19055b7fc(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name: "06 json output of asserting a single existing compliant artifact (using --fingerprint) for an environment results in OK and zero exit",
			cmd:  fmt.Sprintf(`assert artifact --output json --fingerprint %s --environment %s %s`, suite.fingerprint1, suite.envName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "environment"},
				{"environment", suite.envName},
				{"flows", "length:1"},
			},
		},
		{
			name:        "07 asserting a single existing compliant artifact (using --fingerprint) for an policy results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact --fingerprint %s --policy %s %s`, suite.fingerprint1, suite.policyName1, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*Policy-name.*assert-artifact-policy-1.*Attestation-name.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact-one/artifacts/0089a849fce9c7c9128cd13a2e8b1c0757bdb6a7bad0fdf2800e38c19055b7fc(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n.*",
		},
		{
			name: "08 json output of asserting a single existing compliant artifact (using --fingerprint) for an policy results in OK and zero exit\"",
			cmd:  fmt.Sprintf(`assert artifact --output json --fingerprint %s --policy %s %s`, suite.fingerprint1, suite.policyName1, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "policy"},
				{"flows", "length:1"},
			},
		},
		{
			name:        "09 asserting a single existing compliant artifact (using --artifact-type) results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact %s --artifact-type file %s`, suite.artifact1Path, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact-one/artifacts/0089a849fce9c7c9128cd13a2e8b1c0757bdb6a7bad0fdf2800e38c19055b7fc?(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name: "10 json output of asserting a single existing compliant artifact (using --artifact-type) results in OK and zero exit",
			cmd:  fmt.Sprintf(`assert artifact %s --output json --artifact-type file %s`, suite.artifact1Path, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "flow"},
				{"flows", "length:1"},
			},
		},
		{
			name: "11 json output of asserting a single existing compliant artifact (using --fingerprint) for a policy results in correct json",
			cmd:  fmt.Sprintf(`assert artifact %s --output json --artifact-type file --policy=%s --policy=%s %s`, suite.artifact1Path, suite.policyName1, suite.policyName2, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "policy"},
				{"policy_evaluations", "length:2"},
				{"policy_evaluations.[0].policy_name", suite.policyName1},
				{"policy_evaluations.[1].policy_name", suite.policyName2}},
		},
		{
			name:        "12 asserting a multi existing compliant artifact (using --fingerprint) results in OK and zero exit",
			cmd:         fmt.Sprintf(`assert artifact --fingerprint %s %s`, suite.fingerprint2, suite.defaultKosliArguments),
			goldenRegex: "(?s)^COMPLIANT\n.*Flow: assert-artifact-one.*Flow: assert-artifact-two.*Attestation-name.*See more details at http://localhost(:8001)?/docs-cmd-test-user/flows/assert-artifact-one/artifacts/130fabe054d8d90b5d899833cfc769253a39b38854eb0c64214b68b276ef07e8(?:\\?artifact_id=[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{8})?\n",
		},
		{
			name: "13 json output of asserting a multi existing compliant artifact (using --fingerprint) results in OK and zero exit",
			cmd:  fmt.Sprintf(`assert artifact --output json --fingerprint %s %s`, suite.fingerprint2, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"compliant", true},
				{"scope", "flow"},
				{"flows", "length:2"},
			},
		},
		{
			wantError: true,
			name:      "14 not providing --fingerprint nor --artifact-type fails",
			cmd:       fmt.Sprintf(`assert artifact --flow %s %s`, suite.flowName1, suite.defaultKosliArguments),
			golden:    "Error: docker image name or file/dir path is required when --fingerprint is not provided\nUsage: kosli assert artifact [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "15 providing both --environment and --polices fails",
			cmd:       fmt.Sprintf(`assert artifact --fingerprint %s --environment %s --policy %s %s`, suite.fingerprint1, suite.envName, suite.policyName1, suite.defaultKosliArguments),
			golden:    "Error: Cannot specify both 'environment_name' and 'policy_name' at the same time\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertArtifactCommandTestSuite))
}
