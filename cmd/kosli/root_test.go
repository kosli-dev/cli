package main

import (
	"bytes"
	"fmt"
	"io"
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
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf("--host %s --org %s --api-token %s",
		global.Host, global.Org, global.ApiToken)
}

func (suite *UpdateNoticeTestSuite) TestVersionFlagPrintsNotice() {
	const fakeNotice = "\nA new version of the Kosli CLI is available: v9.99.0 (you have v0.0.1)\nUpgrade: https://docs.kosli.com/getting_started/install/\n"
	defer version.SetCheckForUpdateOverride(func(string) (string, error) { return fakeNotice, nil })()

	var errBuf bytes.Buffer
	origErrOut := logger.ErrOut
	logger.ErrOut = &errBuf
	defer func() { logger.ErrOut = origErrOut }()

	cmd, err := newRootCmd(io.Discard, &errBuf, []string{"--version"})
	suite.Require().NoError(err)

	cmd.SetArgs([]string{"--version"})
	suite.NoError(innerMain(cmd, []string{"kosli", "--version"}))
	suite.Contains(errBuf.String(), "A new version")
}

func (suite *UpdateNoticeTestSuite) TestVersionNoticeSkippedForJSON() {
	const fakeNotice = "\nA new version of the Kosli CLI is available: v9.99.0 (you have v0.0.1)\nUpgrade: https://docs.kosli.com/getting_started/install/\n"

	defer version.SetCheckForUpdateOverride(func(string) (string, error) { return fakeNotice, nil })()

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
