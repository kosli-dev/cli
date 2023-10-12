package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type WorkflowReportTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	auditTrailName        string
}

func (suite *WorkflowReportTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	suite.auditTrailName = "a-process"

	EnableBeta(suite.T())
	CreateAuditTrail(suite.auditTrailName, suite.T())
}

func (suite *WorkflowReportTestSuite) TestWorkflowReportCmd() {
	tests := []cmdTestCase{
		{
			name:   "report workflow without description",
			cmd:    `report workflow --audit-trail ` + suite.auditTrailName + ` --id example-31` + suite.defaultKosliArguments,
			golden: fmt.Sprintf("workflow was created in audit-trail '%s' with ID '%s'\n", suite.auditTrailName, "example-31"),
		},
		{
			name:   "report workflow with description",
			cmd:    `report workflow --audit-trail ` + suite.auditTrailName + ` --id example-32 --description "example\!32"` + suite.defaultKosliArguments,
			golden: fmt.Sprintf("workflow was created in audit-trail '%s' with ID '%s'\n", suite.auditTrailName, "example-32"),
		},
	}
	runTestCmd(suite.T(), tests)
}

func TestWorkflowReportCommandTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowReportTestSuite))
}
