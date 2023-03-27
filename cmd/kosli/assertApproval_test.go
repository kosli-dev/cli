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
type AssertApprovalCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

type assertApprovalTestConfig struct {
	createApproval bool
}

func (suite *AssertApprovalCommandTestSuite) SetupTest() {
	suite.flowName = "assert-approval"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.T())
}

func (suite *AssertApprovalCommandTestSuite) TestAssertApprovalCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing --org fails",
			cmd:       fmt.Sprintf(`assert approval --fingerprint 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c  --flow %s --api-token secret`, suite.flowName),
			golden:    "Error: --org is not set\nUsage: kosli assert approval [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "asserting approval for a non existing artifact fails",
			cmd:       fmt.Sprintf(`assert approval --fingerprint 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c  --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c' does not exist in flow 'assert-approval' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "asserting an existing artifact that does not have an approval (using --fingerprint) works and exits with non-zero code",
			cmd:       fmt.Sprintf(`assert approval --fingerprint %s --flow %s %s`, suite.fingerprint, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: artifact with fingerprint fcf33337634c2577a5d86fd7ecb0a25a7c1bb5d89c14fd236f546a5759252c02 has no approvals created\n",
		},
		{
			wantError: true,
			name:      "asserting approval of an existing artifact that does not have an approval (using --artifact-type) works and exits with non-zero code",
			cmd:       fmt.Sprintf(`assert approval %s --artifact-type file --flow %s %s`, suite.artifactPath, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: artifact with fingerprint fcf33337634c2577a5d86fd7ecb0a25a7c1bb5d89c14fd236f546a5759252c02 has no approvals created\n",
		},
		{
			name:   "asserting approval of an existing artifact that has an approval (using --artifact-type) works and exits with zero code",
			cmd:    fmt.Sprintf(`assert approval %s --artifact-type file --flow %s %s`, suite.artifactPath, suite.flowName, suite.defaultKosliArguments),
			golden: "artifact with fingerprint fcf33337634c2577a5d86fd7ecb0a25a7c1bb5d89c14fd236f546a5759252c02 is approved (approval no. [1])\n",
			additionalConfig: assertApprovalTestConfig{
				createApproval: true,
			},
		},
		{
			wantError: true,
			name:      "not providing --fingerprint nor --artifact-type fails",
			cmd:       fmt.Sprintf(`assert approval --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: docker image name or file/dir path is required when --fingerprint is not provided\nUsage: kosli assert approval [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		// TODO: this test case does not pass as the validation does not check for it
		// {
		// 	wantError: true,
		// 	name:      "providing both --fingerprint and --artifact-type fails",
		// 	cmd:       fmt.Sprintf(`assert approval --artifact-type file --fingerprint %s --flow %s %s`, suite.fingerprint, suite.flowName, suite.defaultKosliArguments),
		// 	golden:    "COMPLIANT\n",
		// },
		{
			wantError: true,
			name:      "missing --flow fails",
			cmd:       fmt.Sprintf(`assert approval --fingerprint %s  %s`, suite.fingerprint, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
	}

	for _, t := range tests {
		if t.additionalConfig != nil && t.additionalConfig.(assertApprovalTestConfig).createApproval {
			CreateApproval(suite.flowName, suite.fingerprint, suite.T())
		}
		runTestCmd(suite.T(), []cmdTestCase{t})
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAssertApprovalCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AssertApprovalCommandTestSuite))
}
