package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListTrailsCommandTestSuite struct {
	suite.Suite
	flowName              string
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListTrailsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}

	suite.flowName = "list-trails"
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --host %s --org %s --api-token %s", suite.flowName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail("trail-name", suite.flowName, "", suite.T())

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --flow %s --host %s --org %s --api-token %s", suite.flowName, global.Host, global.Org, global.ApiToken)
}

func (suite *ListTrailsCommandTestSuite) TestListTrailsCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing trails works when there are trails",
			cmd:    fmt.Sprintf(`list trails %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "listing trails works when there are no trails",
			cmd:    fmt.Sprintf(`list trails %s`, suite.acmeOrgKosliArguments),
			golden: "No trails were found.\n",
		},
		{
			name:   "listing trails with --output json works when there are trails",
			cmd:    fmt.Sprintf(`list trails --output json %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "listing trails with --output json works when there are no trails",
			cmd:    fmt.Sprintf(`list trails --output json %s`, suite.acmeOrgKosliArguments),
			golden: "[]\n",
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list trails xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list trails\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListTrailsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListTrailsCommandTestSuite))
}
