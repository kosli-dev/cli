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
		// {
		// 	name: "get workflow works",
		// 	cmd: fmt.Sprintf(`get workflow %s --audit-trail %s %s`,
		// 		suite.workflowID, suite.auditTrailName, suite.workflowOrgKosliArguments),
		// 	goldenFile: "output/get/get-workflow.txt",
		// },
		// {
		// 	wantError: true,
		// 	name:      "get workflow fails when there is no workflow",
		// 	cmd:       fmt.Sprintf(`get workflow non-existing --audit-trail %s %s`, suite.auditTrailName, suite.workflowOrgKosliArguments),
		// 	golden:    "Error: Workflow with ID 'non-existing' does not exist for audit trail testGetWorkflow in organization 'get-workflow-org'\n",
		// },
		{
			wantError: true,
			name:      "get workflow fails when there is no workflow",
			cmd:       fmt.Sprintf(`get workflow non-existing --audit-trail %s %s`, suite.auditTrailName, suite.workflowOrgKosliArguments),
			golden:    "Error: The audit trail feature is in beta. You can enable the feature by running the following Kosli CLI command (version 2.3.2 or later):\n$ kosli enable beta\nJoin our Slack community for more information: https://www.kosli.com/community/\n",
		},
		// {
		// 	name: "get workflow works with --output json when there is workflow",
		// 	cmd: fmt.Sprintf(`get workflow %s --audit-trail %s --output json %s`,
		// 		suite.workflowID, suite.auditTrailName, suite.workflowOrgKosliArguments),
		// },
		// {
		// 	wantError: true,
		// 	name:      "get workflow fails with --output json when there is no workflow",
		// 	cmd: fmt.Sprintf(`get workflow non-existing --audit-trail %s --output json %s`,
		// 		suite.auditTrailName, suite.workflowOrgKosliArguments),
		// 	golden: "Error: Workflow with ID 'non-existing' does not exist for audit trail testGetWorkflow in organization 'get-workflow-org'\n",
		// },
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetWorkflowCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetWorkflowCommandTestSuite))
}
