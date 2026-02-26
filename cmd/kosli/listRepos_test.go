package main

import (
	"fmt"
	"testing"

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
			name:        "02-listing repos works when there are no repos",
			cmd:         fmt.Sprintf(`list repos %s`, suite.acmeOrgKosliArguments),
			goldenRegex: ".*\nkosli-dev/cli  https://github.com/kosli-dev/cli  Trail Started at.*",
		},
		{
			name:       "03-listing repos with --output json works when there are repos",
			cmd:        fmt.Sprintf(`list repos --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"_embedded.repos", "non-empty"}},
		},
		// THIS TEST IS FLAKY IN CI SINCE CI VARIABLES ARE SET THERE AND REPOS MAY EXIST FROM OTHER TESTS
		// {
		// 	name:       "04-listing repos with --output json works when there are no repos",
		// 	cmd:        fmt.Sprintf(`list repos --output json %s`, suite.defaultKosliArguments),
		// 	goldenJson: []jsonCheck{{"_embedded.repos", "[]"}},
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
			cmd:    fmt.Sprintf(`list repos --page-limit 15 --page 2 %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:        "09-listing repos with --name filter works",
			cmd:         fmt.Sprintf(`list repos --name kosli-dev/cli %s`, suite.acmeOrgKosliArguments),
			goldenRegex: ".*\nkosli-dev/cli  https://github.com/kosli-dev/cli  Trail Started at.*",
		},
		{
			name:       "10-listing repos with --name filter and --output json works",
			cmd:        fmt.Sprintf(`list repos --name kosli-dev/cli --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"_embedded.repos", "non-empty"}},
		},
		{
			name:        "11-listing repos with --provider filter works",
			cmd:         fmt.Sprintf(`list repos --provider github %s`, suite.acmeOrgKosliArguments),
			goldenRegex: ".*\nkosli-dev/cli  https://github.com/kosli-dev/cli  Trail Started at.*",
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
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListReposCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListReposCommandTestSuite))
}
