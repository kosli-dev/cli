package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListControlsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListControlsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateControl(global.Org, "list-control-1", "First control", suite.T())
	CreateControl(global.Org, "list-control-2", "Second control", suite.T())

	// fixtures for filtering: a tagged control and an archived control with
	// unique identifiers/tags so assertions are robust to shared server state.
	CreateControl(global.Org, "tagged-control", "Tagged control", suite.T())
	TagControl(global.Org, "tagged-control", map[string]string{"slice2b": "tagtest"}, suite.T())
	CreateControl(global.Org, "archived-control", "Archived control", suite.T())
	ArchiveControl(global.Org, "archived-control", suite.T())

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListControlsCommandTestSuite) TestListControlsCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing controls works when some exist",
			cmd:    "list controls" + suite.defaultKosliArguments,
			golden: "",
		},
		{
			name:   "listing controls works when there are none",
			cmd:    "list controls" + suite.acmeOrgKosliArguments,
			golden: "No controls were found.\n",
		},
		{
			name:       "listing controls with --output json works when some exist",
			cmd:        "list controls --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "non-empty"}},
		},
		{
			name:       "listing controls with --output json works when there are none",
			cmd:        "list controls --output json" + suite.acmeOrgKosliArguments,
			goldenJson: []jsonCheck{{"controls", "[]"}},
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       "list controls xxx" + suite.defaultKosliArguments,
			golden:    "Error: unknown command \"xxx\" for \"kosli list controls\"\n",
		},
		{
			name:       "--page-limit caps the page size and the response echoes the pagination params",
			cmd:        "list controls --page-limit 1 --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "length:1"}, {"page", float64(1)}, {"per_page", float64(1)}},
		},
		{
			name:       "a page beyond the data returns an empty controls list (json)",
			cmd:        "list controls --page 999 --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "[]"}},
		},
		{
			name:   "a page beyond the data reports the empty page (table)",
			cmd:    "list controls --page 999" + suite.defaultKosliArguments,
			golden: "No controls were found at page number 999.\n",
		},
		{
			wantError:   true,
			name:        "--page must be a positive integer",
			cmd:         "list controls --page 0" + suite.defaultKosliArguments,
			goldenRegex: "^Error: page number must be a positive integer",
		},
		{
			wantError:   true,
			name:        "--page-limit must be a positive integer",
			cmd:         "list controls --page-limit 0" + suite.defaultKosliArguments,
			goldenRegex: "^Error: page limit must be a positive integer",
		},
		{
			name:       "--search matches by identifier",
			cmd:        "list controls --search list-control-1 --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "length:1"}, {"controls.[0].identifier", "list-control-1"}},
		},
		{
			name:       "--tag returns only controls carrying that tag",
			cmd:        "list controls --tag slice2b:tagtest --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "length:1"}, {"controls.[0].identifier", "tagged-control"}},
		},
		{
			name:       "--sort-direction asc returns controls in ascending name order",
			cmd:        "list controls --search list-control --sort-direction asc --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "length:2"}, {"controls.[0].identifier", "list-control-1"}, {"controls.[1].identifier", "list-control-2"}},
		},
		{
			name:       "--sort-direction desc returns controls in descending name order",
			cmd:        "list controls --search list-control --sort-direction desc --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "length:2"}, {"controls.[0].identifier", "list-control-2"}, {"controls.[1].identifier", "list-control-1"}},
		},
		{
			name:       "archived controls are excluded by default",
			cmd:        "list controls --search archived-control --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "[]"}},
		},
		{
			name:       "--archived returns archived controls",
			cmd:        "list controls --archived --search archived-control --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"controls", "length:1"}, {"controls.[0].identifier", "archived-control"}},
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestListControlsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListControlsCommandTestSuite))
}
