package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UpdateDefaultOrgCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *UpdateDefaultOrgCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --api-token %s", global.Host, global.ApiToken)
}

func (suite *UpdateDefaultOrgCommandTestSuite) TestUpdateDefaultOrgCmd() {
	tests := []cmdTestCase{
		{
			name:   "can set default organization",
			cmd:    fmt.Sprintf(`update default-org docs-cmd-test-user %s`, suite.defaultKosliArguments),
			golden: "default organization is set to: docs-cmd-test-user\n",
		},
		{
			wantError:   false,
			name:        "dry-run builds the right url without contacting the server",
			cmd:         fmt.Sprintf(`update default-org docs-cmd-test-user --dry-run %s`, suite.defaultKosliArguments),
			goldenRegex: `the request would have been sent to: .*api/v2/user/docs-cmd-test-user`,
		},
		{
			wantError: true,
			name:      "setting default org fails when no args are provided",
			cmd:       fmt.Sprintf(`update default-org %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "setting default org fails when 2 args are provided",
			cmd:       fmt.Sprintf(`update default-org org1 org2 %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "setting default org fails when org name is empty",
			cmd:       fmt.Sprintf(`update default-org "" %s`, suite.defaultKosliArguments),
			golden:    "Error: ORG-NAME argument is required\nUsage: kosli update default-org ORG-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "setting default org fails for non-existing org",
			cmd:       fmt.Sprintf(`update default-org non-existing-org-abc123 %s`, suite.defaultKosliArguments),
			golden:    "Error: Access denied to /api/v2/user/non-existing-org-abc123\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestUpdateDefaultOrgCommandTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateDefaultOrgCommandTestSuite))
}
