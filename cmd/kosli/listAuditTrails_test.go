package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListAuditTrailsCommandTestSuite struct {
	suite.Suite
	acmeOrgKosliArguments string
	iuOrgKosliArguments   string
}

func (suite *ListAuditTrailsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c",
		Org:      "acme-org",
		Host:     "http://localhost:8001",
	}
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	// CreateAuditTrail("testAuditTrail", suite.T()) // create an audit trail for the acme-org

	global.Org = "iu-org"
	global.ApiToken = "qM9u2_grv6pJLbACwsMMMT5LIQy82tQj2k1zjZnlXti1smnFaGwCKW4jzk0La7ae9RrSYvEwCXSsXknD6YZqd-onLaaIUUKtEn6-B6yh53vWIe9EC5u85FCbKZjFbaicp_d0Me0Zcqq_KcCgrAZRX9xggl_pBb2oaCsNdllqNjk"
	suite.iuOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListAuditTrailsCommandTestSuite) TestListAuditTrailsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "listing audit trails works when there are audit trails",
			cmd:       fmt.Sprintf(`list audit-trails %s`, suite.acmeOrgKosliArguments),
			golden:    "Command \"audit-trails\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: You don't have access to the audit trails feature. This feature has been deprecated in favour of Flows and Trails.\n",
		},
		{
			wantError: true,
			name:      "listing audit trails works when there are no audit trails",
			cmd:       fmt.Sprintf(`list audit-trails %s`, suite.iuOrgKosliArguments),
			golden:    "Command \"audit-trails\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: You don't have access to the audit trails feature. This feature has been deprecated in favour of Flows and Trails.\n",
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list audit-trails xxx %s`, suite.acmeOrgKosliArguments),
			golden:    "Command \"audit-trails\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: unknown command \"xxx\" for \"kosli list audit-trails\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListAuditTrailsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListAuditTrailsCommandTestSuite))
}
