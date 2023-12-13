package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type BeginTrailCommandTestSuite struct {
	flowName string
	suite.Suite
	defaultKosliArguments string
}

func (suite *BeginTrailCommandTestSuite) SetupTest() {
	suite.flowName = "begin-trail"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
}

func (suite *BeginTrailCommandTestSuite) TestBeginTrailCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("begin trail trail1 xxx --flow %s %s", suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when name is considered invalid by the server",
			cmd:       fmt.Sprintf("begin trail foo?$bar --flow %s %s", suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Input payload validation failed: map[name:'foo?$bar' does not match '^[a-zA-Z0-9][a-zA-Z0-9\\\\-_\\\\.~]*$']\n",
		},
		{
			wantError: true,
			name:      "fails when --flow is missing",
			cmd:       "begin trail test-123 --description \"my new flow\" " + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError:   true,
			name:        "beginning a trail with an invalid template fails",
			cmd:         fmt.Sprintf("begin trail test-123 --flow %s --template-file testdata/invalid_template.yml %s", suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "Error: template file is invalid 1 validation error for Template\n.*",
		},
		{
			name:   "can begin a trail with a valid template",
			cmd:    fmt.Sprintf("begin trail test-123 --flow %s --template-file testdata/valid_template.yml %s", suite.flowName, suite.defaultKosliArguments),
			golden: "trail 'test-123' was begun\n",
		},
		{
			name:   "can update a trail with a description",
			cmd:    fmt.Sprintf("begin trail test-123 --flow %s --template-file testdata/valid_template.yml --description \"my new flow\" %s", suite.flowName, suite.defaultKosliArguments),
			golden: "trail 'test-123' was updated\n",
		},
		{
			wantError: true,
			name:      "missing --org flag causes an error",
			cmd:       "begin trail test-123 --flow my-modern-flow -H http://localhost:8001 -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: --org is not set\nUsage: kosli begin trail TRAIL-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing --api-token flag causes an error",
			cmd:       "begin trail test-123 --flow my-modern-flow --org cyber-dojo -H http://localhost:8001",
			golden:    "Error: --api-token is not set\nUsage: kosli begin trail TRAIL-NAME [flags]\n",
		},
		{
			wantError: true,
			name:      "missing name argument fails",
			cmd:       "begin trail --flow my-modern-flow  -H http://localhost:8001 --org cyber-dojo -a eyJhbGciOiJIUzUxMiIsImlhdCI6MTYyNTY0NDUwMCwiZXhwIjoxNjI1NjQ4MTAwfQ.eyJpZCI6IjgzYTBkY2Q1In0.1B-xDlajF46vipL49zPbnXBRgotqGGcB3lxwpJxZ3HNce07E0p2LwO7UDYve9j2G9fQtKrKhUKvVR97SQOEFLQ",
			golden:    "Error: trail name must be provided as an argument\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestBeginTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(BeginTrailCommandTestSuite))
}
