package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListWorkflowsCommandTestSuite struct {
	suite.Suite
	atrOrgKosliArguments      string
	atrEmptyOrgKosliArguments string
	auditTrailName            string
	workflowID                string
}

func (suite *ListWorkflowsCommandTestSuite) SetupTest() {
	suite.auditTrailName = "testAuditTrail"
	suite.workflowID = "testExternalId"
	global = &GlobalOpts{
		ApiToken: "z8qw5f3Vf1TXz10LruL8QjHrnya3Un1-InOm0jsdmUWVQuvBfNs2Yo2Whr7KA4DHn4mTiVjURBc0V9ZZ9fVEG1GVSI7YWriBJTg-7RK7a3zakymorXhiNi-6Z2M-nCXB0qdl8f1ECTfj7V0oN_JzWEREX-64_fNBbhRF97PZtiI",
		Org:      "workflows-org",
		Host:     "http://localhost:8001",
	}
	suite.atrOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	// CreateAuditTrail(suite.auditTrailName, suite.T())                         // create an audit trail for the workflows-org
	// CreateWorkflow(suite.auditTrailName, suite.workflowID, suite.T())         // create workflow for the workflows-org
	// CreateWorkflowEvidence(suite.auditTrailName, suite.workflowID, suite.T()) // create workflow evidence for the workflows-org

	global.Org = "workflows-empty-org"
	global.ApiToken = "Fmbyc_Obhwna69rxvZVeOUS_8r-57ZCdqCK2QRfy1Q2hNzgPNjcOO1aaXmMlRT4Bts7kapjg1MXvVXwJmrCBkAx3RUtrgLvdLZZ5wZ46xRdRb0yvGrmXi08fcWqU8l9cET0oHk6TeAnK3iHq-SzP7D3_gjmZf1H9nKiEoIfsIIw"
	suite.atrEmptyOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	// CreateAuditTrail(suite.auditTrailName, suite.T()) // create an audit trail for the workflows-empty-org
}

func (suite *ListWorkflowsCommandTestSuite) TestListWorkflowsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "listing workflows works when there are no workflows",
			cmd:       fmt.Sprintf(`list workflows --audit-trail %s %s`, suite.auditTrailName, suite.atrEmptyOrgKosliArguments),
			golden:    "Command \"workflows\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: You don't have access to the audit trails feature. This feature has been deprecated in favour of Flows and Trails.\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListWorkflowsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListWorkflowsCommandTestSuite))
}
