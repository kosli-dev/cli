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
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListAuditTrailsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListAuditTrailsCommandTestSuite) TestListAuditTrailsCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing audit trails works when there are audit trails",
			cmd:    fmt.Sprintf(`list audit-trails %s`, suite.defaultKosliArguments),
			golden: "", // some audit trails exist from other tests
		},
		{
			name:   "listing audit trails works when there are no audit trails",
			cmd:    fmt.Sprintf(`list audit-trails %s`, suite.acmeOrgKosliArguments),
			golden: "No audit trails were found.\n",
		},
		{
			name:   "listing audit trails with --output json works when there are audit trails",
			cmd:    fmt.Sprintf(`list audit-trails --output json %s`, suite.defaultKosliArguments),
			golden: "", // some audit trails exist from other tests
		},
		{
			name:   "listing audit trails with --output json works when there are no audit trails",
			cmd:    fmt.Sprintf(`list audit-trails --output json %s`, suite.acmeOrgKosliArguments),
			golden: "[]\n",
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list audit-trails xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list audit-trails\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListAuditTrailsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListAuditTrailsCommandTestSuite))
}
