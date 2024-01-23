package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetWorkflowCommandTestSuite struct {
	suite.Suite
	workflowOrgKosliArguments string
	auditTrailName            string
	workflowID                string
}

func (suite *GetWorkflowCommandTestSuite) SetupTest() {
	suite.auditTrailName = "testGetWorkflow"
	suite.workflowID = "testExternalId"
	global = &GlobalOpts{
		ApiToken: "edzkHOnV23d5h6n23wspTKzv3VSdmpjQscZXNbRMA63ym4TAIiaOGeS9AEkAXO_x7RkVGvvHB_LsZUIQWtsuij1k_1nY7AuqjNrcZTWj3ww8m8E78ZvbuT-dEUjjPhYsaqr055cL8osdR1tGWEB-l-eTy4AAI0Tl0Q1nA1DY-ig",
		Org:      "get-workflow-org",
		Host:     "http://localhost:8001",
	}

	suite.workflowOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	// CreateAuditTrail(suite.auditTrailName, suite.T())                         // create an audit trail for the get-workflow-org
	// CreateWorkflow(suite.auditTrailName, suite.workflowID, suite.T())         // create workflow for the get-workflow-org
	// CreateWorkflowEvidence(suite.auditTrailName, suite.workflowID, suite.T()) // create workflow evidence for the get-workflow-org
}

func (suite *GetWorkflowCommandTestSuite) TestGetWorkflowCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "get workflow fails when there is no workflow",
			cmd:       fmt.Sprintf(`get workflow non-existing --audit-trail %s %s`, suite.auditTrailName, suite.workflowOrgKosliArguments),
			golden:    "Command \"workflow\" is deprecated, Audit trails are deprecated. Please use Flows and Trail instead.\nError: You don't have access to the audit trails feature. This feature has been deprecated in favour of Flows and Trails.\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetWorkflowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetWorkflowCommandTestSuite))
}
