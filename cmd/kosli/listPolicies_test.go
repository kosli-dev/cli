package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListPoliciesCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListPoliciesCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreatePolicy(global.Org, "test-policy", suite.T())

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListPoliciesCommandTestSuite) TestListPoliciesCmd() {
	tests := []cmdTestCase{
		{
			name: "listing policies works when there are envs",
			cmd:  fmt.Sprintf(`list policies %s`, suite.defaultKosliArguments),
		},
		{
			name:   "listing policies works when there are no envs",
			cmd:    fmt.Sprintf(`list policies %s`, suite.acmeOrgKosliArguments),
			golden: "No environment policies were found.\n",
		},
		{
			name:       "listing policies with --output json works when there are polices",
			cmd:        fmt.Sprintf(`list policies --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"", "non-empty"}},
		},
		{
			name:       "listing policies with --output json works when there are no polices",
			cmd:        fmt.Sprintf(`list policies --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"", "[]"}},
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list policies xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list policies\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListPoliciesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListPoliciesCommandTestSuite))
}
