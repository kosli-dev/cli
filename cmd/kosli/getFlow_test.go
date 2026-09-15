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
type GetFlowCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
}

func (suite *GetFlowCommandTestSuite) SetupTest() {
	suite.flowName = "get-flow"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
}

func (suite *GetFlowCommandTestSuite) TestGetFlowCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "getting a non existing flow fails",
			cmd:       fmt.Sprintf(`get flow non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Flow named 'non-existing' does not exist for organization 'docs-cmd-test-user'\n",
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get flow non-existing xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			name: "getting an existing flow works",
			cmd:  fmt.Sprintf(`get flow %s %s`, suite.flowName, suite.defaultKosliArguments),
		},
		{
			name: "getting an existing flow with --output json works",
			cmd:  fmt.Sprintf(`get flow %s --output json %s`, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetFlowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetFlowCommandTestSuite))
}

func TestPrintFlowAsTableOmitsVisibility(t *testing.T) {
	// visibility is a legacy per-flow field with no effect on access, so the
	// table must not surface it even when the server still returns it
	raw := `{"name":"backend","description":"Backend service","visibility":"private","template":"artifact pull-request","last_deployment_at":1700000000,"tags":{"team":"platform"}}`
	var buf bytes.Buffer
	require.NoError(t, printFlowAsTable(raw, &buf, 0))
	out := buf.String()
	require.NotContains(t, out, "Visibility:")
	require.NotContains(t, out, "private")
	require.Contains(t, out, "Name:")
	require.Contains(t, out, "Description:")
	require.Contains(t, out, "Template:")
	require.Contains(t, out, "Last Deployment At:")
	require.Contains(t, out, "Tags:")
	require.Contains(t, out, "backend")
	require.Contains(t, out, "Backend service")
	require.Contains(t, out, "[team=platform]")
}
