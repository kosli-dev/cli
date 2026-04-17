package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/version"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type RootCommandTestSuite struct {
	suite.Suite
}

func (suite *RootCommandTestSuite) TestConfigProcessing() {
	tests := []cmdTestCase{
		{
			name:        "using a plain text api token",
			cmd:         "version --config-file testdata/config/plain-text-token.yaml --debug",
			goldenRegex: "\\[debug\\] processing config file \\[testdata\\/config\\/plain-text-token.yaml\\]\n\\[warning\\].*\n\\[warning\\] using api token from \\[testdata\\/config\\/plain-text-token.yaml\\] as plain text. It is recommended to encrypt your api token by setting it with: kosli config --api-token <token>.*\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestRootCommandTestSuite(t *testing.T) {
	suite.Run(t, new(RootCommandTestSuite))
}

type UpdateNoticeTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *UpdateNoticeTestSuite) SetupTest() {
	suite.defaultKosliArguments = fmt.Sprintf("--host %s --org %s --api-token %s",
		global.Host, global.Org, global.ApiToken)
}

func (suite *UpdateNoticeTestSuite) TestVersionNoticeSkippedForJSON() {
	const fakeNotice = "\nA new version of the Kosli CLI is available: v9.99.0 (you have v0.0.1)\nUpgrade: https://docs.kosli.com/getting_started/install/\n"

	orig := version.OverrideCheckForUpdate
	version.OverrideCheckForUpdate = func(string) (string, error) { return fakeNotice, nil }
	defer func() { version.OverrideCheckForUpdate = orig }()

	// with --output json: no notice in stderr
	_, _, _, stderr, err := executeCommandC(
		fmt.Sprintf("list flows --output json %s", suite.defaultKosliArguments))
	suite.NoError(err)
	suite.Empty(stderr)

	// with --output table: notice IS in stderr
	_, _, _, stderr, err = executeCommandC(
		fmt.Sprintf("list flows --output table %s", suite.defaultKosliArguments))
	suite.NoError(err)
	suite.Contains(stderr, "A new version")
}

func TestUpdateNoticeTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateNoticeTestSuite))
}
