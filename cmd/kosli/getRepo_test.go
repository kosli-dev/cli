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
	repoInnerID           string
}

func (suite *GetRepoCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	global.Org = "iu-org-shared"
	global.ApiToken = "qM9u2_grv6pJLbACwsMMMT5LIQy82tQj2k1zjZnlXti1smnFaGwCKW4jzk0La7ae9RrSYvEwCXSsXknD6YZqd-onLaaIUUKtEn6-B6yh53vWIe9EC5u85FCbKZjFbaicp_d0Me0Zcqq_KcCgrAZRX9xggl_pBb2oaCsNdllqNjk"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate("get-repo", "testdata/valid_template.yml", suite.T())
	SetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "1234",
		"GITHUB_SERVER_URL":    "https://github.com",
		"GITHUB_REPOSITORY":    "kosli-dev/cli",
		"GITHUB_REPOSITORY_ID": "1234567890",
	}, suite.T())
	BeginTrail("trail-name", "get-repo", "", suite.T())

	// two repos sharing a name but with different external ids, to exercise
	// the "narrow down the search" guard. The name is deliberately distinctive
	// so no other suite creates a same-named repo in this shared org and makes
	// the match count (asserted below) unstable.
	SetEnvVars(map[string]string{
		"GITHUB_REPOSITORY":    "get-repo-suite-org/get-repo-ambiguous-repo",
		"GITHUB_REPOSITORY_ID": "111",
	}, suite.T())
	BeginTrail("ambiguous-trail-1", "get-repo", "", suite.T())
	SetEnvVars(map[string]string{
		"GITHUB_REPOSITORY_ID": "222",
	}, suite.T())
	BeginTrail("ambiguous-trail-2", "get-repo", "", suite.T())

	// tag the repo so tests can assert tags are surfaced
	suite.repoInnerID = GetRepoInnerID(global.Org, "kosli-dev/cli", suite.T())
	TagRepo(global.Org, suite.repoInnerID, map[string]string{"team": "platform"}, suite.T())
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
			wantError:   true,
			name:        "01-getting a non-existing repo gives a not-found error",
			cmd:         fmt.Sprintf(`get repo non-existing/repo %s`, suite.defaultKosliArguments),
			goldenRegex: `^Error: Repo 'non-existing/repo' not found`,
		},
		{
			name:        "02-getting an existing repo surfaces its id and tags",
			cmd:         fmt.Sprintf(`get repo kosli-dev/cli %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `(?s)Name:\s+kosli-dev/cli\n.*ID:\s+\S+\n.*Tags:\s+team=platform`,
		},
		{
			name: "03-getting an existing repo with --output json works",
			cmd:  fmt.Sprintf(`get repo kosli-dev/cli --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{
				{"name", "kosli-dev/cli"},
				{"id", "not-nil"},
				{"tags.team", "platform"},
			},
		},
		{
			name: "04-getting an existing repo with matching --provider works",
			cmd:  fmt.Sprintf(`get repo kosli-dev/cli --provider github %s`, suite.acmeOrgKosliArguments),
		},
		{
			name:       "05-getting an existing repo with matching --provider and --output json works",
			cmd:        fmt.Sprintf(`get repo kosli-dev/cli --provider github --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"provider", "github"}, {"id", "not-nil"}},
		},
		{
			wantError:   true,
			name:        "06-getting a repo with a non-matching --provider gives a not-found error",
			cmd:         fmt.Sprintf(`get repo kosli-dev/cli --provider gitlab %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `^Error: Repo 'kosli-dev/cli' not found`,
		},
		{
			name:        "07-getting a repo by its internal id works",
			cmd:         fmt.Sprintf(`get repo --repo-id %s %s`, suite.repoInnerID, suite.acmeOrgKosliArguments),
			goldenRegex: `(?s)Name:\s+kosli-dev/cli\n.*ID:\s+\S+`,
		},
		{
			wantError:   true,
			name:        "08-providing neither a name argument nor --repo-id fails",
			cmd:         fmt.Sprintf(`get repo %s`, suite.defaultKosliArguments),
			goldenRegex: `^Error: exactly one of the REPO-NAME argument or --repo-id must be provided`,
		},
		{
			wantError: true,
			name:      "09-providing more than one argument fails",
			cmd:       fmt.Sprintf(`get repo foo bar %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError:   true,
			name:        "10-getting a repo with multiple matches suggests specifying the provider",
			cmd:         fmt.Sprintf(`get repo get-repo-suite-org/get-repo-ambiguous-repo %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `^Error: Multiple repos named 'get-repo-suite-org/get-repo-ambiguous-repo' exist \(github\)\. Specify the 'provider'`,
		},
		{
			name: "11-narrowing an ambiguous repo down with --provider works",
			cmd:  fmt.Sprintf(`get repo get-repo-suite-org/get-repo-ambiguous-repo --provider github %s`, suite.acmeOrgKosliArguments),
		},
		{
			wantError:   true,
			name:        "12-providing both a name argument and --repo-id fails",
			cmd:         fmt.Sprintf(`get repo kosli-dev/cli --repo-id %s %s`, suite.repoInnerID, suite.acmeOrgKosliArguments),
			goldenRegex: `^Error: exactly one of the REPO-NAME argument or --repo-id must be provided`,
		},
		{
			wantError:   true,
			name:        "13-getting a non-existing repo id gives a not-found error",
			cmd:         fmt.Sprintf(`get repo --repo-id no-such-inner-id %s`, suite.acmeOrgKosliArguments),
			goldenRegex: `^Error: Repo .* not found`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetRepoCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetRepoCommandTestSuite))
}
