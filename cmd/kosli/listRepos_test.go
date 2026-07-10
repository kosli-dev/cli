package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListReposCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListReposCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	global.Org = "acme-org-shared"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate("list-repos", "testdata/valid_template.yml", suite.T())
	SetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "1234",
		"GITHUB_SERVER_URL":    "https://github.com",
		"GITHUB_REPOSITORY":    "kosli-dev/cli",
		"GITHUB_REPOSITORY_ID": "1234567890",
	}, suite.T())
	BeginTrail("trail-name", "list-repos", "", suite.T())

	// tag the repo so tests can assert tags are surfaced
	innerID := GetRepoInnerID(global.Org, "kosli-dev/cli", suite.T())
	TagRepo(global.Org, innerID, map[string]string{"team": "platform"}, suite.T())

	// a second repo that is never tagged, to assert blank TAGS rendering
	SetEnvVars(map[string]string{
		"GITHUB_REPOSITORY":    "list-repos-suite-org/untagged-repo",
		"GITHUB_REPOSITORY_ID": "555",
	}, suite.T())
	BeginTrail("untagged-trail", "list-repos", "", suite.T())

	// three repos sharing a search substring, in a known A–Z order, so the
	// --sort-direction tests can assert the ordering actually reverses
	sortRepos := []struct{ name, id string }{
		{"list-repos-suite-org/sort-a", "701"},
		{"list-repos-suite-org/sort-b", "702"},
		{"list-repos-suite-org/sort-c", "703"},
	}
	for _, r := range sortRepos {
		SetEnvVars(map[string]string{
			"GITHUB_REPOSITORY":    r.name,
			"GITHUB_REPOSITORY_ID": r.id,
		}, suite.T())
		BeginTrail("sort-trail-"+r.id, "list-repos", "", suite.T())
	}
}

func (suite *ListReposCommandTestSuite) TearDownTest() {
	UnSetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "",
		"GITHUB_SERVER_URL":    "",
		"GITHUB_REPOSITORY":    "",
		"GITHUB_REPOSITORY_ID": "",
	}, suite.T())
}

func (suite *ListReposCommandTestSuite) TestListReposCmd() {
	tests := []cmdTestCase{
		// THIS TEST IS FLAKY IN CI SINCE CI VARIABLES ARE SET THERE AND REPOS MAY EXIST FROM OTHER TESTS
		// {
		// 	name:   "01-listing repos works when there are repos",
		// 	cmd:    fmt.Sprintf(`list repos %s`, suite.defaultKosliArguments),
		// 	golden: "No repos were found.\n",
		// },
		{
			name:        "02-listing repos works when there are repos",
			cmd:         fmt.Sprintf(`list repos %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?sm)NAME\s+URL\s+PROVIDER\s+TAGS.*^kosli-dev/cli\s+https://github\.com/kosli-dev/cli\s+github\s+team=platform`,
		},
		{
			name:       "03-listing repos with --output json works when there are repos",
			cmd:        fmt.Sprintf(`list repos --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"repos", "non-empty"}},
		},
		// THIS TEST IS FLAKY IN CI SINCE CI VARIABLES ARE SET THERE AND REPOS MAY EXIST FROM OTHER TESTS
		// {
		// 	name:       "04-listing repos with --output json works when there are no repos",
		// 	cmd:        fmt.Sprintf(`list repos --output json %s`, suite.defaultKosliArguments),
		// 	goldenJson: []jsonCheck{{"repos", "[]"}},
		// },
		{
			wantError: true,
			name:      "05-providing an argument causes an error",
			cmd:       fmt.Sprintf(`list repos xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list repos\"\n",
		},
		{
			wantError: true,
			name:      "06-negative page limit causes an error",
			cmd:       fmt.Sprintf(`list repos --page-limit -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "07-negative page number causes an error",
			cmd:       fmt.Sprintf(`list repos --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			name:   "08-can list repos with pagination",
			cmd:    fmt.Sprintf(`list repos --page-limit 25 --page 2 %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:        "09-listing repos with --name filter works",
			cmd:         fmt.Sprintf(`list repos --name kosli-dev/cli %s`, suite.acmeOrgKosliArguments),
			goldenRegex: ".*\nkosli-dev/cli.*https://github.com/kosli-dev/cli.*github.*",
		},
		{
			name:       "10-listing repos with --name filter and --output json works",
			cmd:        fmt.Sprintf(`list repos --name kosli-dev/cli --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"repos", "non-empty"}},
		},
		{
			name:        "11-listing repos with --provider filter works",
			cmd:         fmt.Sprintf(`list repos --provider github %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?m)^kosli-dev/cli\s+https://github\.com/kosli-dev/cli\s+github\b`,
		},
		{
			name:   "12-listing repos with non-matching --provider returns no repos message",
			cmd:    fmt.Sprintf(`list repos --provider gitlab %s`, suite.acmeOrgKosliArguments),
			golden: "No repos were found.\n",
		},
		{
			name:   "13-listing repos with non-matching --repo-id returns no repos message",
			cmd:    fmt.Sprintf(`list repos --repo-id non-existing-id %s`, suite.acmeOrgKosliArguments),
			golden: "No repos were found.\n",
		},
		{
			name:        "14a-listing repos with --search substring works",
			cmd:         fmt.Sprintf(`list repos --search cli %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?m)^kosli-dev/cli\s+https://github\.com/kosli-dev/cli\s+github\b`,
		},
		{
			name:       "14b-listing repos with --search and --output json works",
			cmd:        fmt.Sprintf(`list repos --search cli --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"repos", "non-empty"}},
		},
		{
			wantError: true,
			name:      "14c-using --name and --search together causes an error",
			cmd:       fmt.Sprintf(`list repos --name kosli-dev/cli --search cli %s`, suite.acmeOrgKosliArguments),
			golden:    "Error: if any flags in the group [name search] are set none of the others can be; [name search] were all set\n",
		},
		{
			name:        "14d-listing repos filtered by --tag key:value works",
			cmd:         fmt.Sprintf(`list repos --tag team:platform %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?m)^kosli-dev/cli\s+https://github\.com/kosli-dev/cli\s+github\s+team=platform`,
		},
		{
			name:        "14e-listing repos filtered by --tag key only works",
			cmd:         fmt.Sprintf(`list repos --tag team %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?m)^kosli-dev/cli\s+https://github\.com/kosli-dev/cli\s+github\s+team=platform`,
		},
		{
			name:   "14f-listing repos with a non-matching --tag returns no repos message",
			cmd:    fmt.Sprintf(`list repos --tag team:doesnotexist %s`, suite.acmeOrgKosliArguments),
			golden: "No repos were found.\n",
		},
		{
			name: "14g-listing repos with --sort-direction asc keeps A–Z order",
			cmd:  fmt.Sprintf(`list repos --search list-repos-suite-org/sort- --sort-direction asc --page 1 --page-limit 50 --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{
				{"repos", "length:3"},
				{"repos.[0].name", "list-repos-suite-org/sort-a"},
				{"repos.[2].name", "list-repos-suite-org/sort-c"},
			},
		},
		{
			name: "14h-listing repos with --sort-direction desc reverses the order",
			cmd:  fmt.Sprintf(`list repos --search list-repos-suite-org/sort- --sort-direction desc --page 1 --page-limit 50 --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{
				{"repos", "length:3"},
				{"repos.[0].name", "list-repos-suite-org/sort-c"},
				{"repos.[2].name", "list-repos-suite-org/sort-a"},
			},
		},
		{
			name:        "14-a repo without tags renders a blank TAGS cell",
			cmd:         fmt.Sprintf(`list repos %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?m)^list-repos-suite-org/untagged-repo\s+https://github\.com/list-repos-suite-org/untagged-repo\s+github\s*$`,
		},
		{
			name:        "15-table output shows the pagination footer",
			cmd:         fmt.Sprintf(`list repos %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?m)^Showing page 1 of \d+, total \d+ repos$`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListReposCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListReposCommandTestSuite))
}

func TestPrintReposListAsTableRendersBlankTagsCell(t *testing.T) {
	raw := `{"repos":[
		{"name":"o/tagged","url":"https://github.com/o/tagged","provider":"github","tags":{"team":"platform"}},
		{"name":"o/untagged","url":"https://github.com/o/untagged","provider":"github","tags":{}}
	],"page":1,"total_pages":3,"total_count":45}`
	var buf bytes.Buffer
	require.NoError(t, printReposListAsTable(raw, &buf, 1))
	out := buf.String()
	require.Regexp(t, `(?m)^NAME\s+URL\s+PROVIDER\s+TAGS\s*$`, out)
	require.Regexp(t, `(?m)^o/tagged\s+https://github\.com/o/tagged\s+github\s+team=platform\s*$`, out)
	require.Regexp(t, `(?m)^o/untagged\s+https://github\.com/o/untagged\s+github\s*$`, out)
	require.Regexp(t, `(?m)^Showing page 1 of 3, total 45 repos$`, out)
}
