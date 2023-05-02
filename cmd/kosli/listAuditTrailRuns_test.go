package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListAuditTrailRunsCommandTestSuite struct {
	suite.Suite
	atrOrgKosliArguments      string
	atrEmptyOrgKosliArguments string
	auditTrailName            string
}

func (suite *ListAuditTrailRunsCommandTestSuite) SetupTest() {
	suite.auditTrailName = "testAuditTrail"
	global = &GlobalOpts{
		ApiToken: "z8qw5f3Vf1TXz10LruL8QjHrnya3Un1-InOm0jsdmUWVQuvBfNs2Yo2Whr7KA4DHn4mTiVjURBc0V9ZZ9fVEG1GVSI7YWriBJTg-7RK7a3zakymorXhiNi-6Z2M-nCXB0qdl8f1ECTfj7V0oN_JzWEREX-64_fNBbhRF97PZtiI",
		Org:      "audit-trail-runs-org",
		Host:     "http://localhost:8001",
	}
	suite.atrOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateAuditTrail(suite.auditTrailName, suite.T())       // create an audit trail for the audit-trail-runs-org
	CreateWorkflowEvidence(suite.auditTrailName, suite.T()) // create an audit trail evidence for the audit-trail-runs-org

	global.Org = "audit-trail-runs-empty-org"
	global.ApiToken = "Fmbyc_Obhwna69rxvZVeOUS_8r-57ZCdqCK2QRfy1Q2hNzgPNjcOO1aaXmMlRT4Bts7kapjg1MXvVXwJmrCBkAx3RUtrgLvdLZZ5wZ46xRdRb0yvGrmXi08fcWqU8l9cET0oHk6TeAnK3iHq-SzP7D3_gjmZf1H9nKiEoIfsIIw"
	suite.atrEmptyOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateAuditTrail(suite.auditTrailName, suite.T()) // create an audit trail for the audit-trail-runs-empty-org
}

func (suite *ListAuditTrailRunsCommandTestSuite) TestListAuditTrailRunsCmd() {
	tests := []cmdTestCase{
		{
			name:       "listing audit trail runs works when there are audit trail runs",
			cmd:        fmt.Sprintf(`list audit-trail-runs --audit-trail %s %s`, suite.auditTrailName, suite.atrOrgKosliArguments),
			goldenFile: "output/list/list-audit-trail-runs.txt",
		},
		{
			name:   "listing audit trail runs works when there are no audit trails runs",
			cmd:    fmt.Sprintf(`list audit-trail-runs --audit-trail %s %s`, suite.auditTrailName, suite.atrEmptyOrgKosliArguments),
			golden: "No audit trial runs were found.\n",
		},
		{
			name: "listing audit trail runs with --output json works when there are audit trail runs",
			cmd:  fmt.Sprintf(`list audit-trail-runs --audit-trail %s --output json %s`, suite.auditTrailName, suite.atrOrgKosliArguments),
		},
		{
			name:   "listing audit trail runs with --output json works when there are no audit trail runs",
			cmd:    fmt.Sprintf(`list audit-trail-runs --audit-trail %s --output json %s`, suite.auditTrailName, suite.atrEmptyOrgKosliArguments),
			golden: "[]\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListAuditTrailRunsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListAuditTrailRunsCommandTestSuite))
}
