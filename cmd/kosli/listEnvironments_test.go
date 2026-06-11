package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListEnvironmentsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListEnvironmentsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	// dedicated envs with a unique name prefix so pagination/filter tests are
	// deterministic even when other test suites create envs in the same org
	CreateEnv(global.Org, "list-envs-934-a", "server", suite.T())
	CreateEnv(global.Org, "list-envs-934-b", "docker", suite.T())
	CreateEnv(global.Org, "list-envs-934-c", "K8S", suite.T())
	TagEnv("list-envs-934-a", "team", "platform", suite.T())

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListEnvironmentsCommandTestSuite) TestListEnvironmentsCmd() {
	tests := []cmdTestCase{
		{
			name: "listing environments works when there are envs",
			cmd:  fmt.Sprintf(`list environments %s`, suite.defaultKosliArguments),
		},
		{
			name:   "listing environments works when there are no envs",
			cmd:    fmt.Sprintf(`list environments %s`, suite.acmeOrgKosliArguments),
			golden: "No environments were found.\n",
		},
		{
			name:       "listing environments with --output json works when there are envs",
			cmd:        fmt.Sprintf(`list environments --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"", "non-empty"}},
		},
		{
			name:       "listing environments with --output json works when there are no envs",
			cmd:        fmt.Sprintf(`list environments --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"", "[]"}},
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list environments xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list environments\"\n",
		},
		{
			wantError: true,
			name:      "--page 0 causes an error",
			cmd:       fmt.Sprintf(`list environments --page 0 %s`, suite.defaultKosliArguments),
			golden:    "Error: page number must be a positive integer\nUsage: kosli list environments [flags]\n",
		},
		{
			wantError: true,
			name:      "--page-limit 0 causes an error",
			cmd:       fmt.Sprintf(`list environments --page-limit 0 %s`, suite.defaultKosliArguments),
			golden:    "Error: page limit must be a positive integer\nUsage: kosli list environments [flags]\n",
		},
		{
			wantError: true,
			name:      "negative --page causes an error",
			cmd:       fmt.Sprintf(`list environments --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			name:        "paginated table output shows a pagination footer",
			cmd:         fmt.Sprintf(`list environments --page 1 --page-limit 2 %s`, suite.defaultKosliArguments),
			goldenRegex: `(?s).*Showing page 1 of \d+, total \d+ items\n$`,
		},
		{
			name: "paginated json output has the paginated response shape",
			cmd:  fmt.Sprintf(`list environments --page 1 --page-limit 2 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:2"},
				{"page", float64(1)},
				{"per_page", float64(2)},
				{"total_count", "not-nil"},
				{"total_pages", "not-nil"},
			},
		},
		{
			name:   "paginating beyond the last page reports no environments at that page",
			cmd:    fmt.Sprintf(`list environments --page 99 --page-limit 50 %s`, suite.acmeOrgKosliArguments),
			golden: "No environments were found at page number 99.\n",
		},
		{
			name: "--name matches environments whose name contains the substring",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:3"},
			},
		},
		{
			name:   "--name with no matching substring returns no environments",
			cmd:    fmt.Sprintf(`list environments --name no-such-env-substring-xyz %s`, suite.defaultKosliArguments),
			golden: "No environments were found.\n",
		},
		{
			name: "--type filters environments by type",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --type docker --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:1"},
				{"environments.[0].name", "list-envs-934-b"},
			},
		},
		{
			name: "--type can be repeated to match multiple types",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --type docker --type server --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:2"},
			},
		},
		{
			name: "--tag filters environments by tag key:value",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --tag team:platform --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:1"},
				{"environments.[0].name", "list-envs-934-a"},
			},
		},
		{
			name: "--tag filters environments by tag key only",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --tag team --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:1"},
				{"environments.[0].name", "list-envs-934-a"},
			},
		},
		// TODO: re-enable once the server returns an empty list instead of a
		// 5xx for unknown space IDs: https://github.com/kosli-dev/server/issues/5858
		// {
		// 	name: "--space-id with an unknown space returns no environments",
		// 	cmd:  fmt.Sprintf(`list environments --space-id no-such-space-id --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
		// 	goldenJson: []jsonCheck{
		// 		{"environments", "[]"},
		// 	},
		// },
		{
			name: "--sort name --sort-direction desc reverses the name order",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --sort name --sort-direction desc --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:3"},
				{"environments.[0].name", "list-envs-934-c"},
				{"environments.[2].name", "list-envs-934-a"},
			},
		},
		{
			name: "--sort-direction asc keeps the name order",
			cmd:  fmt.Sprintf(`list environments --name list-envs-934 --sort-direction asc --page 1 --page-limit 50 --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"environments", "length:3"},
				{"environments.[0].name", "list-envs-934-a"},
				{"environments.[2].name", "list-envs-934-c"},
			},
		},
		{
			wantError:   true,
			name:        "an invalid --sort value surfaces the API error",
			cmd:         fmt.Sprintf(`list environments --sort no-such-field %s`, suite.defaultKosliArguments),
			goldenRegex: `^Error: .*`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListEnvironmentsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListEnvironmentsCommandTestSuite))
}
