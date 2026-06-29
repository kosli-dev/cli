package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListFlowsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListFlowsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	// create flows with deterministic names so the --name / --case-sensitive tests
	// can assert match / no-match behaviour reliably
	CreateFlow("list-flows-search-target", suite.T())

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListFlowsCommandTestSuite) TestListFlowsCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing flows works when there are flows",
			cmd:    fmt.Sprintf(`list flows %s`, suite.defaultKosliArguments),
			golden: "", // some flows exist from other tests
		},
		{
			name:   "listing flows works when there are no flows",
			cmd:    fmt.Sprintf(`list flows %s`, suite.acmeOrgKosliArguments),
			golden: "No flows were found.\n",
		},
		{
			name:       "listing flows with --output json works when there are flows",
			cmd:        fmt.Sprintf(`list flows --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "non-empty"}},
		},
		{
			name:       "listing flows with --output json works when there are no flows",
			cmd:        fmt.Sprintf(`list flows --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"data", "[]"}},
		},
		{
			name:       "--name matches flows whose name contains the substring",
			cmd:        fmt.Sprintf(`list flows --name list-flows-search-target --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "non-empty"}, {"data.[0].name", "list-flows-search-target"}},
		},
		{
			name:       "--name with no matching substring returns an empty list",
			cmd:        fmt.Sprintf(`list flows --name no-such-flow-substring-xyz --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "[]"}},
		},
		{
			name:       "--name matching is case sensitive by default so a wrong-case substring does not match",
			cmd:        fmt.Sprintf(`list flows --name LIST-FLOWS-SEARCH-TARGET --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "[]"}},
		},
		{
			name:       "--ignore-case makes a wrong-case substring match",
			cmd:        fmt.Sprintf(`list flows --name LIST-FLOWS-SEARCH-TARGET --ignore-case --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "non-empty"}, {"data.[0].name", "list-flows-search-target"}},
		},
		{
			name:       "short flags -N and -i work like --name and --ignore-case",
			cmd:        fmt.Sprintf(`list flows -N LIST-FLOWS-SEARCH-TARGET -i --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "non-empty"}, {"data.[0].name", "list-flows-search-target"}},
		},
		{
			name:       "pagination metadata is returned on page 1",
			cmd:        fmt.Sprintf(`list flows --page-limit 1 --page 1 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"data", "non-empty"}, {"pagination.page", float64(1)}},
		},
		{
			name:       "can page through flows with --page",
			cmd:        fmt.Sprintf(`list flows --page-limit 1 --page 2 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"pagination.page", float64(2)}},
		},
		{
			name:   "an empty page reports the page number",
			cmd:    fmt.Sprintf(`list flows --page 99 %s`, suite.defaultKosliArguments),
			golden: "No flows were found at page number 99.\n",
		},
		{
			wantError: true,
			name:      "negative page limit causes an error",
			cmd:       fmt.Sprintf(`list flows --page-limit -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "negative page number causes an error",
			cmd:       fmt.Sprintf(`list flows --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "zero page limit causes an error",
			cmd:       fmt.Sprintf(`list flows --page-limit 0 %s`, suite.defaultKosliArguments),
			golden:    "Error: page limit must be a positive integer\nUsage: kosli list flows [flags]\n",
		},
		{
			wantError: true,
			name:      "zero page number causes an error",
			cmd:       fmt.Sprintf(`list flows --page 0 %s`, suite.defaultKosliArguments),
			golden:    "Error: page number must be a positive integer\nUsage: kosli list flows [flags]\n",
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list flows xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list flows\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListFlowsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListFlowsCommandTestSuite))
}
