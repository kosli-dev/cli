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
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when name is considered invalid by the server",
			cmd:       "create audit-trail foo_bar --steps step1,step2" + suite.defaultKosliArguments,
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: Input payload validation failed: map[name:'foo_bar' does not match '^[a-zA-Z0-9\\\\-]+$']\n",
		},
		{
			wantError: true,
			name:      "cannot create an audit trail when it is not enabled for you",
			cmd:       "create audit-trail newAuditTrail --description \"my new audit trail\" --steps step1,step2" + suite.defaultKosliArguments,
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: You don't have access to the audit trails feature. This feature has been deprecated in favour of Flows and Trails.\n",
		},
		{
			wantError: true,
			name:      "re-creating a flow updates its metadata",
			cmd:       "create audit-trail newAuditTrail --description \"changed description\" --steps step1,step2" + suite.defaultKosliArguments,
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: You don't have access to the audit trails feature. This feature has been deprecated in favour of Flows and Trails.\n",
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd:       "create audit-trail newAuditTrail --description \"my new audit trail\" --steps step1,step2 -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: --org is not set\nUsage: kosli create audit-trail AUDIT-TRAIL-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "create audit-trail newAuditTrail --description \"my new audit trail\" --steps step1,step2 --org cyber-dojo -H http://localhost:8001",
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: --api-token is not set\nUsage: kosli create audit-trail AUDIT-TRAIL-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing name argument fails",
			cmd:       "create audit-trail --description \"my new flow\" --steps step1,step2" + suite.defaultKosliArguments,
			golden:    "Command \"audit-trail\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: audit trail name must be provided as an argument\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateAuditTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAuditTrailCommandTestSuite))
}
