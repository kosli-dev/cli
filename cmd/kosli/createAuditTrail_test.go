package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CreateAuditTrailCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateAuditTrailCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateAuditTrailCommandTestSuite) TestCreateAuditTrailCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       "create audit-trail newAuditTrail xxx" + suite.defaultKosliArguments,
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when name is considered invalid by the server",
			cmd:       "create audit-trail foo_bar" + suite.defaultKosliArguments,
			golden:    "Error: Input payload validation failed: map[name:'foo_bar' does not match '^[a-zA-Z0-9\\\\-]+$']\n",
		},
		{
			name:   "can create an audit trail",
			cmd:    "create audit-trail newAuditTrail --description \"my new audit trail\" " + suite.defaultKosliArguments,
			golden: "flow 'newAuditTrail' was created\n",
		},
		{
			name:   "re-creating a flow updates its metadata",
			cmd:    "create audit-trail newAuditTrail --description \"changed description\" " + suite.defaultKosliArguments,
			golden: "audit trail 'newAuditTrail' was created\n",
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd:       "create audit-trail newAuditTrail --description \"my new audit trail\" -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --org is not set\nUsage: kosli create audit-trail [AUDIT-TRAIL-NAME] [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "create audit-trail newAuditTrail --description \"my new audit trail\" --org cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli create audit-trail [AUDIT-TRAIL-NAME] [flags]\n",
		},
		{
			wantError: true,
			name:      "missing name argument fails",
			cmd:       "create audit-trail --description \"my new flow\" -H http://localhost:8001 --org cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: audit trail name must be provided as an argument\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateAuditTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAuditTrailCommandTestSuite))
}
