package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetTrailCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	trailName             string
}

func (suite *GetTrailCommandTestSuite) SetupTest() {
	suite.flowName = "get-trail"
	suite.trailName = "cli-build-1"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

func (suite *GetTrailCommandTestSuite) TestGetTrailCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "getting a non existing trail fails",
			cmd:       fmt.Sprintf(`get trail foo --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Trail with name 'foo' does not exist for Organization '%s' and Flow '%s'\n", global.Org, suite.flowName),
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get trail %s xxx --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`get trail %s --flow %s --org orgX`, suite.trailName, suite.flowName),
			golden:    "Error: --api-token is not set\nUsage: kosli get trail TRAIL-NAME [flags]\n",
		},
		{
			name:       "getting an existing trail using works",
			cmd:        fmt.Sprintf(`get trail %s --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-trail.txt",
		},
		{
			name: "getting an existing trail with --output json works",
			cmd:  fmt.Sprintf(`get trail %s --flow %s --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetTrailCommandTestSuite))
}
