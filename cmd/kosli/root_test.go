package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/kosli-dev/cli/internal/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
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

func (suite *RootCommandTestSuite) TestQuietFlagSuppressesWarnings() {
	_, _, _, stderr, err := executeCommandC(
		"version --config-file testdata/config/plain-text-token.yaml --quiet")
	suite.NoError(err)
	suite.NotContains(stderr, "[warning]",
		"--quiet should suppress warning output, got: %q", stderr)
}

func (suite *RootCommandTestSuite) TestDebugWinsOverQuiet() {
	_, _, _, stderr, err := executeCommandC(
		"version --config-file testdata/config/plain-text-token.yaml --quiet --debug")
	suite.NoError(err)
	suite.Contains(stderr, "[warning]",
		"--debug should override --quiet, expected warnings in stderr, got: %q", stderr)
	suite.Contains(stderr, "[debug] --quiet is ignored because --debug is set",
		"expected debug notice that --quiet was overridden, got: %q", stderr)
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

	var errBuf bytes.Buffer
	origErrOut := logger.ErrOut
	logger.ErrOut = &errBuf
	defer func() { logger.ErrOut = origErrOut }()

	cmd, err := newRootCmd(io.Discard, &errBuf, []string{"--version"})
	suite.Require().NoError(err)

	var called bool
	defer version.SetCheckForUpdateOverride(func(string) (string, error) {
		called = true
		return fakeNotice, nil
	})()

	cmd.SetArgs([]string{"--version"})
	suite.NoError(innerMain(cmd, []string{"kosli", "--version"}))
	suite.True(called, "expected CheckForUpdate override to be called for --version")
	suite.Contains(errBuf.String(), "A new version")
}

func (suite *UpdateNoticeTestSuite) TestVersionNoticeNotShownOnRegularCommands() {
	const fakeNotice = "\nA new version of the Kosli CLI is available: v9.99.0 (you have v0.0.1)\nUpgrade: https://docs.kosli.com/getting_started/install/\n"

	defer version.SetCheckForUpdateOverride(func(string) (string, error) { return fakeNotice, nil })()

	// The update check only runs for the `version` subcommand and the
	// `--version` flag — regular commands must not print the notice,
	// regardless of output format.
	for _, format := range []string{"json", "table"} {
		_, _, _, stderr, err := executeCommandC(
			fmt.Sprintf("list flows --output %s %s", format, suite.defaultKosliArguments))
		suite.NoError(err)
		suite.NotContains(stderr, "A new version", "no update notice expected for --output %s", format)
	}
}

func TestUpdateNoticeTestSuite(t *testing.T) {
	suite.Run(t, new(UpdateNoticeTestSuite))
}

func TestEnrichError(t *testing.T) {
	// leaf builds kosli -> attest -> snyk and returns the snyk leaf command,
	// optionally defining and setting the flow/trail flags on it.
	leaf := func(withFlags bool, flow, trail string) *cobra.Command {
		root := &cobra.Command{Use: "kosli"}
		attest := &cobra.Command{Use: "attest"}
		snyk := &cobra.Command{Use: "snyk"}
		root.AddCommand(attest)
		attest.AddCommand(snyk)
		if withFlags {
			snyk.Flags().String("flow", "", "")
			snyk.Flags().String("trail", "", "")
			if flow != "" {
				require.NoError(t, snyk.Flags().Set("flow", flow))
			}
			if trail != "" {
				require.NoError(t, snyk.Flags().Set("trail", trail))
			}
		}
		return snyk
	}

	t.Run("nil error passes through unchanged", func(t *testing.T) {
		require.NoError(t, enrichError(leaf(false, "", ""), nil))
	})

	t.Run("nil cmd passes error through unchanged", func(t *testing.T) {
		e := errors.New("boom")
		require.Equal(t, e, enrichError(nil, e))
	})

	t.Run("command path only when no flow/trail flags exist", func(t *testing.T) {
		got := enrichError(leaf(false, "", ""), errors.New("server returned 404"))
		require.EqualError(t, got, `[kosli attest snyk] server returned 404`)
	})

	t.Run("includes flow and trail when set", func(t *testing.T) {
		got := enrichError(leaf(true, "cyber-dojo", "live-snyk-scan"), errors.New("server returned 404"))
		require.EqualError(t, got,
			`[kosli attest snyk flow=cyber-dojo trail=live-snyk-scan] server returned 404`)
	})

	t.Run("empty flag values are omitted", func(t *testing.T) {
		got := enrichError(leaf(true, "", ""), errors.New("boom"))
		require.EqualError(t, got, `[kosli attest snyk] boom`)
	})

	t.Run("preserves the wrapped error for errors.Is", func(t *testing.T) {
		// enrichError must wrap with %w so callers (and errors.Is/errors.As)
		// can still unwrap the original error. This guards against an
		// accidental switch to %v / %s.
		sentinel := errors.New("server returned 404")
		got := enrichError(leaf(true, "cyber-dojo", "live-snyk-scan"), sentinel)
		require.ErrorIs(t, got, sentinel)
	})
}
