package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetAuditTrailCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	auditTrailName        string
}

func (suite *GetAuditTrailCommandTestSuite) SetupTest() {
	suite.auditTrailName = "audit-trail-get-test"
	global = &GlobalOpts{
		ApiToken: "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c",
		Org:      "acme-org",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateAuditTrail(suite.auditTrailName, suite.T())
}

func (suite *GetAuditTrailCommandTestSuite) TestGetAuditTrailCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "getting a non existing audit trail fails",
			cmd:       fmt.Sprintf(`get audit-trail non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Audit Trail called 'non-existing' does not exist for organization 'acme-org'. \n",
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get audit-trail non-existing xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			name:       "getting an existing audit trail works",
			cmd:        fmt.Sprintf(`get audit-trail %s %s`, suite.auditTrailName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-audit-trail.txt",
		},
		{
			name: "getting an existing audit trail with --output json works",
			cmd:  fmt.Sprintf(`get audit-trail %s --output json %s`, suite.auditTrailName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetAuditTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetAuditTrailCommandTestSuite))
}
