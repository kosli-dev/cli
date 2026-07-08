package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type TagTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	envName               string
	envType               string
	controlID             string
	repoName              string
	repoID                string
}

func (suite *TagTestSuite) SetupTest() {
	suite.flowName = "tag-flow"
	suite.envName = "tag-env"
	suite.envType = "K8S"
	suite.controlID = "tag-control"
	// unique name: in CI, other suites implicitly create a repo for the real
	// GITHUB_REPOSITORY in this org, and same-name repos make lookup ambiguous
	suite.repoName = "tag-test-org/tag-test-repo"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	CreateEnv(global.Org, suite.envName, suite.envType, suite.T())
	CreateControl(global.Org, suite.controlID, "Tag control", suite.T())

	// repos are created implicitly when a trail is begun with git repo info
	SetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "1234",
		"GITHUB_SERVER_URL":    "https://github.com",
		"GITHUB_REPOSITORY":    suite.repoName,
		"GITHUB_REPOSITORY_ID": "1234567890",
	}, suite.T())
	CreateFlowWithTemplate("tag-repo-flow", "testdata/valid_template.yml", suite.T())
	BeginTrail("tag-repo-trail", "tag-repo-flow", "", suite.T())
	suite.repoID = GetRepoInnerID(global.Org, suite.repoName, suite.T())
}

func (suite *TagTestSuite) TearDownTest() {
	UnSetEnvVars(map[string]string{
		"GITHUB_RUN_NUMBER":    "",
		"GITHUB_SERVER_URL":    "",
		"GITHUB_REPOSITORY":    "",
		"GITHUB_REPOSITORY_ID": "",
	}, suite.T())
}

func (suite *TagTestSuite) TestTagCmd() {
	tests := []cmdTestCase{
		{
			name:   "can tag an environment with one tag",
			cmd:    fmt.Sprintf("tag environment %s --set foo=bar %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] added for environment 'tag-env'\n",
		},
		{
			name:   "can tag an environment using the -s short flag",
			cmd:    fmt.Sprintf("tag environment %s -s foo=bar %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] added for environment 'tag-env'\n",
		},
		{
			name:   "can remove a tag from an environment using the -u short flag",
			cmd:    fmt.Sprintf("tag environment %s -u foo %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] removed for environment 'tag-env'\n",
		},
		{
			name:        "tag command help lists the valid resource types",
			cmd:         "tag --help",
			goldenRegex: `Valid resource types are: flow, flows, env, environment, environments, control, controls, repo, repos\.`,
		},
		{
			name:   "can tag an environment with multiple tags",
			cmd:    fmt.Sprintf("tag environment %s --set foo=bar --set key=value %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo, key] added for environment 'tag-env'\n",
		},
		{
			name:   "can remove tags from an environment",
			cmd:    fmt.Sprintf("tag environment %s --unset foo %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] removed for environment 'tag-env'\n",
		},
		{
			name:   "can add and remove tags from an environment at the same time",
			cmd:    fmt.Sprintf("tag environment %s --set key=value --unset foo %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [key] added, and Tag(s) [foo] removed for environment 'tag-env'\n",
		},
		{
			name:   "removing a non-existing tag is okay",
			cmd:    fmt.Sprintf("tag environment %s --unset non-existing %s", suite.envName, suite.defaultKosliArguments),
			golden: "Tag(s) [non-existing] removed for environment 'tag-env'\n",
		},
		{
			name:   "can tag a flow",
			cmd:    fmt.Sprintf("tag flow %s --set foo=bar %s", suite.flowName, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] added for flow 'tag-flow'\n",
		},
		{
			name:   "can tag a control",
			cmd:    fmt.Sprintf("tag control %s --set foo=bar %s", suite.controlID, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] added for control 'tag-control'\n",
		},
		{
			name:   "can remove a tag from a control",
			cmd:    fmt.Sprintf("tag control %s --unset foo %s", suite.controlID, suite.defaultKosliArguments),
			golden: "Tag(s) [foo] removed for control 'tag-control'\n",
		},
		{
			name:   "can tag a control using the plural resource type",
			cmd:    fmt.Sprintf("tag controls %s --set key=value %s", suite.controlID, suite.defaultKosliArguments),
			golden: "Tag(s) [key] added for controls 'tag-control'\n",
		},
		{
			wantError:   true,
			name:        "tagging a non-existing control gives a clear error",
			cmd:         fmt.Sprintf("tag control no-such-control --set foo=bar %s", suite.defaultKosliArguments),
			goldenRegex: "^Error: \"Control 'no-such-control' does not exist in organization",
		},
		{
			name:   "can tag a repo by its name",
			cmd:    fmt.Sprintf("tag repo %s --set team=platform %s", suite.repoName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("Tag(s) [team] added for repo '%s'\n", suite.repoName),
		},
		{
			name:   "can tag a repo with --provider",
			cmd:    fmt.Sprintf("tag repo %s --provider github --set env=prod %s", suite.repoName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("Tag(s) [env] added for repo '%s'\n", suite.repoName),
		},
		{
			name:   "can remove a tag from a repo",
			cmd:    fmt.Sprintf("tag repo %s --unset team %s", suite.repoName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("Tag(s) [team] removed for repo '%s'\n", suite.repoName),
		},
		{
			name:   "can tag a repo using the plural resource type",
			cmd:    fmt.Sprintf("tag repos %s --set key=value %s", suite.repoName, suite.defaultKosliArguments),
			golden: fmt.Sprintf("Tag(s) [key] added for repos '%s'\n", suite.repoName),
		},
		{
			wantError:   true,
			name:        "tagging a non-existing repo gives a clear error",
			cmd:         fmt.Sprintf("tag repo no-such-org/no-such-repo --set foo=bar %s", suite.defaultKosliArguments),
			goldenRegex: "^Error: Repo 'no-such-org/no-such-repo' not found",
		},
		{
			wantError: true,
			name:      "--provider with a non-repo resource type gives a clear error",
			cmd:       fmt.Sprintf("tag flow %s --provider github --set foo=bar %s", suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: --provider is only valid when tagging repos\n",
		},
		{
			name:   "can tag a repo by its internal id via --repo-id",
			cmd:    fmt.Sprintf("tag repo --repo-id %s --set foo=bar %s", suite.repoID, suite.defaultKosliArguments),
			golden: fmt.Sprintf("Tag(s) [foo] added for repo '%s'\n", suite.repoID),
		},
		{
			wantError: true,
			name:      "providing both a repo name and --repo-id gives a clear error",
			cmd:       fmt.Sprintf("tag repo %s --repo-id %s --set foo=bar %s", suite.repoName, suite.repoID, suite.defaultKosliArguments),
			golden:    "Error: exactly one of the RESOURCE-ID argument or --repo-id must be provided\n",
		},
		{
			wantError: true,
			name:      "--repo-id with a non-repo resource type gives a clear error",
			cmd:       fmt.Sprintf("tag flow --repo-id %s --set foo=bar %s", suite.repoID, suite.defaultKosliArguments),
			golden:    "Error: --repo-id is only valid when tagging repos\n",
		},
		{
			wantError: true,
			name:      "--provider combined with --repo-id gives a clear error",
			cmd:       fmt.Sprintf("tag repo --repo-id %s --provider github --set foo=bar %s", suite.repoID, suite.defaultKosliArguments),
			golden:    "Error: --provider cannot be combined with --repo-id\n",
		},
		{
			wantError: true,
			name:      "omitting the resource id without --repo-id gives a clear error",
			cmd:       fmt.Sprintf("tag flow --set foo=bar %s", suite.defaultKosliArguments),
			golden:    "Error: the RESOURCE-ID argument is required unless tagging a repo with --repo-id\n",
		},
		{
			wantError:   true,
			name:        "an invalid resource type gives a clear error listing valid types",
			cmd:         fmt.Sprintf("tag junk some-id --set foo=bar %s", suite.defaultKosliArguments),
			goldenRegex: `^Error: junk is not a valid resource type\. Valid resource types are: \[flow flows env environment environments control controls repo repos\]`,
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestTagTestSuite(t *testing.T) {
	suite.Run(t, new(TagTestSuite))
}
