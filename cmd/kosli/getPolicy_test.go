package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetPolicyCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	policyName            string
}

func (suite *GetPolicyCommandTestSuite) SetupTest() {
	suite.policyName = "get-policy"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreatePolicy(global.Org, suite.policyName, suite.T())
}

func (suite *GetPolicyCommandTestSuite) TestGetPolicyCmd() {
	tests := []cmdTestCase{
		{
			name: "can get a policy that exists",
			cmd:  fmt.Sprintf("get policy %s %s", suite.policyName, suite.defaultKosliArguments),
		},
		{
			name: "can get a policy that exists in json format",
			cmd:  fmt.Sprintf("get policy %s %s --output json", suite.policyName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "trying to get non-existing policy fails",
			cmd:       "get policy non-existing" + suite.defaultKosliArguments,
			golden:    "Error: Policy 'non-existing' does not exist in organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "fails when no argument (policy name) provided",
			cmd:       "get policy " + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetPolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetPolicyCommandTestSuite))
}

func TestPrintPolicyAsTableErrorsOnEmptyVersions(t *testing.T) {
	// a policy with no versions has no content to print; that must be an
	// error, not an index-out-of-range panic
	raw := `{"name":"p1","description":"d","created_at":1452902400,"consuming_envs":[],"versions":[]}`
	var buf bytes.Buffer
	require.NotPanics(t, func() {
		require.Error(t, printPolicyAsTable(raw, &buf, 0))
	})
}

func TestPrintPolicyAsTableRendersLatestVersion(t *testing.T) {
	raw := `{"name":"p1","description":"d","created_at":1452902400,"consuming_envs":["prod","staging"],
	  "versions":[{"policy_yaml":"old: 1"},{"policy_yaml":"_type: policy\nrules: []"}]}`
	var buf bytes.Buffer
	require.NoError(t, printPolicyAsTable(raw, &buf, 0))
	out := buf.String()
	require.Regexp(t, `Versions:\s+2\n`, out)
	require.Regexp(t, `Attached to environments:\s+\[prod staging\]\n`, out)
	require.Regexp(t, `Policy content:\n\s+_type: policy\n\s+rules: \[\]\n`, out)
	require.NotContains(t, out, "old: 1")
}
