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
type ListApprovalsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName1             string
	flowName2             string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

func (suite *ListApprovalsCommandTestSuite) SetupTest() {
	suite.flowName1 = "list-approvals-empty"
	suite.flowName2 = "list-approvals"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlow(suite.flowName1, suite.T())
	CreateFlow(suite.flowName2, suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName2, suite.fingerprint, suite.artifactName, suite.T())
	CreateApproval(suite.flowName2, suite.fingerprint, suite.T())
}

func (suite *ListApprovalsCommandTestSuite) TestListApprovalsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing --flow flag arg causes an error",
			cmd:       fmt.Sprintf(`list approvals %s`, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "non-existing flow causes an error",
			cmd:       fmt.Sprintf(`list approvals --flow non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Flow named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		// TODO: the correct error is overwritten by the hack flag value check in root.go
		{
			wantError: true,
			name:      "negative page number causes an error",
			cmd:       fmt.Sprintf(`list approvals --flow foo --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "negative page limit causes an error",
			cmd:       fmt.Sprintf(`list approvals --flow foo --page-limit -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			name:   "listing approvals on an empty flow works",
			cmd:    fmt.Sprintf(`list approvals --flow %s %s`, suite.flowName1, suite.defaultKosliArguments),
			golden: "No approvals were found.\n",
		},
		{
			name:   "listing approvals on an empty flow with --output json works",
			cmd:    fmt.Sprintf(`list approvals --flow %s --output json %s`, suite.flowName1, suite.defaultKosliArguments),
			golden: "[]\n",
		},
		{
			name:       "listing approvals on a flow works",
			cmd:        fmt.Sprintf(`list approvals --flow %s %s`, suite.flowName2, suite.defaultKosliArguments),
			goldenFile: "output/list/list-approvals.txt",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListApprovalsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListApprovalsCommandTestSuite))
}
