package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetRepoCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *GetRepoCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate("get-repo", "testdata/valid_template.yml", suite.T())
	SetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "1234",
		"GITHUB_SERVER_URL":    "https://github.com",
		"GITHUB_REPOSITORY":    "kosli-dev/cli",
		"GITHUB_REPOSITORY_ID": "1234567890",
	}, suite.T())
	BeginTrail("trail-name", "get-repo", "", suite.T())
}

func (suite *GetRepoCommandTestSuite) TearDownTest() {
	UnSetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "",
		"GITHUB_SERVER_URL":    "",
		"GITHUB_REPOSITORY":    "",
		"GITHUB_REPOSITORY_ID": "",
	}, suite.T())
}

func (suite *GetRepoCommandTestSuite) TestGetRepoCmd() {
	tests := []cmdTestCase{
		{
			name:   "01-getting a non-existing repo returns not-found message",
			cmd:    fmt.Sprintf(`get repo non-existing/repo %s`, suite.defaultKosliArguments),
			golden: "Repo was not found.\n",
		},
		{
			name: "02-getting an existing repo works",
			cmd:  fmt.Sprintf(`get repo kosli-dev/cli %s`, suite.acmeOrgKosliArguments),
		},
		{
			name:       "03-getting an existing repo with --output json works",
			cmd:        fmt.Sprintf(`get repo kosli-dev/cli --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"_embedded.repos", "non-empty"}},
		},
		{
			name: "04-getting an existing repo with matching --provider works",
			cmd:  fmt.Sprintf(`get repo kosli-dev/cli --provider github %s`, suite.acmeOrgKosliArguments),
		},
		{
			name:       "05-getting an existing repo with matching --provider and --output json works",
			cmd:        fmt.Sprintf(`get repo kosli-dev/cli --provider github --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"_embedded.repos", "non-empty"}},
		},
		{
			name:   "06-getting a repo with a non-matching --provider returns not-found message",
			cmd:    fmt.Sprintf(`get repo kosli-dev/cli --provider gitlab %s`, suite.acmeOrgKosliArguments),
			golden: "Repo was not found.\n",
		},
		{
			name:   "07-getting a repo with a non-matching --repo-id returns not-found message",
			cmd:    fmt.Sprintf(`get repo kosli-dev/cli --repo-id non-existing-id %s`, suite.acmeOrgKosliArguments),
			golden: "Repo was not found.\n",
		},
		{
			wantError: true,
			name:      "08-providing no argument fails",
			cmd:       fmt.Sprintf(`get repo %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "09-providing more than one argument fails",
			cmd:       fmt.Sprintf(`get repo foo bar %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetRepoCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetRepoCommandTestSuite))
}
