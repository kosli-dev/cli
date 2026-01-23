package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type SearchCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
}

func (suite *SearchCommandTestSuite) SetupTest() {
	suite.flowName = "some-flow"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifact(suite.flowName, suite.fingerprint, suite.artifactName, suite.T())
}

func (suite *SearchCommandTestSuite) TestSearchCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no args provided",
			cmd:       fmt.Sprintf(`search %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "can search with a git-commit that does not exist",
			cmd:       fmt.Sprintf(`search f040404f9fb3447ed0c8ad48c6af12d5e94513ca %s`, suite.defaultKosliArguments),
			golden:    "Error: No matches for 'f040404f9fb3447ed0c8ad48c6af12d5e94513ca'.\n",
		},
		{
			wantError: true,
			name:      "can search with a git-commit that does not exist with --output json",
			cmd:       fmt.Sprintf(`search f040404f9fb3447ed0c8ad48c6af12d5e94513ca --output json%s`, suite.defaultKosliArguments),
			golden:    "Error: No matches for 'f040404f9fb3447ed0c8ad48c6af12d5e94513ca'.\n",
		},
		{
			wantError: true,
			name:      "can search with a fingerprint that does not exist",
			cmd:       fmt.Sprintf(`search 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef2123 %s`, suite.defaultKosliArguments),
			golden:    "Error: No matches for '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef2123'.\n",
		},
		{
			wantError: true,
			name:      "can search with a fingerprint that does not exist with --output json",
			cmd:       fmt.Sprintf(`search 8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef2123 --output json %s`, suite.defaultKosliArguments),
			golden:    "Error: No matches for '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef2123'.\n",
		},
		{
			name:   "can search with a git-commit that exists",
			cmd:    fmt.Sprintf(`search 0fc1ba9876f91b215679f3649b8668085d820ab5 %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "can search with a fingerprint that exists",
			cmd:    fmt.Sprintf(`search %s %s`, suite.fingerprint, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "can search with a git-commit that exists using --output json",
			cmd:    fmt.Sprintf(`search 0fc1ba9876f91b215679f3649b8668085d820ab5 --output json %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "can search with a fingerprint that exists using --output json",
			cmd:    fmt.Sprintf(`search %s --output json %s`, suite.fingerprint, suite.defaultKosliArguments),
			golden: "",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestSearchCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SearchCommandTestSuite))
}
