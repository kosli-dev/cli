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

// TestFindUnknownCommand locks in the fix for issue #1043: an unrecognized
// command token must be reported (so the process exits non-zero) while genuine
// group commands invoked with no leftover positional keep resolving cleanly.
func (suite *RootCommandTestSuite) TestFindUnknownCommand() {
	cases := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{name: "unknown top-level command", args: []string{"garbage"}, wantErr: true},
		{name: "unknown command with trailing args", args: []string{"garbage", "list", "flows"}, wantErr: true},
		{name: "unknown subcommand of a group", args: []string{"list", "garbage"}, wantErr: true},
		{name: "group command with no leftover is fine", args: []string{"list"}, wantErr: false},
		{name: "runnable leaf command is fine", args: []string{"version"}, wantErr: false},
		{name: "no args is fine", args: []string{}, wantErr: false},
		{name: "unknown flag defers to cobra", args: []string{"--badflag", "list", "flows"}, wantErr: false},
		// A space-form bool flag on a group command must be normalized the same
		// way innerMain normalizes it, otherwise its value ("false") is mistaken
		// for a leftover positional and wrongly reported as an unknown command.
		{name: "space-form bool flag on a group command is fine", args: []string{"list", "--debug", "false"}, wantErr: false},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			err := findUnknownCommand(tc.args)
			if tc.wantErr {
				suite.Require().Error(err)
				suite.Contains(err.Error(), "unknown command:")
				suite.Contains(err.Error(), "available subcommands are:")
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

// TestInnerMainPreservesGlobalAfterProbe guards the regression the unknown-command
// probe could introduce: findUnknownCommand builds a throwaway command tree via
// newRootCmd, which reassigns the package-level `global`. If it is not restored,
// the real command's persistent-flag pointers are orphaned and every global.*
// read (ApiToken, Host, ...) sees the throwaway struct's defaults instead of the
// user's flag values. Drive a successful real command with global flags through
// innerMain (which the golden harness bypasses) and assert the values survive.
func (suite *RootCommandTestSuite) TestInnerMainPreservesGlobalAfterProbe() {
	args := []string{
		"fingerprint", "testdata/person-schema.json",
		"--artifact-type", "file",
		"--host", "https://example.kosli.com",
		"--api-token", "DRY_RUN",
	}
	cmd, err := newRootCmd(io.Discard, io.Discard, args)
	suite.Require().NoError(err)

	err = innerMain(cmd, append([]string{"kosli"}, args...))

	suite.Require().NoError(err)
	suite.Equal("https://example.kosli.com", global.Host,
		"global.Host must reflect --host, not the probe's throwaway struct default")
	suite.Equal("DRY_RUN", global.ApiToken,
		"global.ApiToken must reflect --api-token, not the probe's throwaway struct default")
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

// TestEmptyApiTokenEnvVarStillDecryptsConfigToken pins that KOSLI_API_TOKEN set
// to an empty string does not suppress decryption of a config-file token.
// bindFlags decides the token came from the environment with os.LookupEnv, where
// a variable set to "" reports as present, so it skips decryption and applies
// the config value verbatim - which sends ciphertext for anyone who stored their
// token with `kosli config --api-token`.
//
// The fixture's token is not valid ciphertext, so decryption fails and falls
// through to the plain-text warning whether or not an encryption key exists in
// the credentials store. That makes the warning a stable marker for "the
// decryption branch was entered".
func (suite *RootCommandTestSuite) TestEmptyApiTokenEnvVarStillDecryptsConfigToken() {
	suite.T().Setenv("KOSLI_API_TOKEN", "")

	_, _, _, stderr, err := executeCommandC(
		"version --config-file testdata/config/plain-text-token.yaml")

	suite.Require().NoError(err)
	suite.Contains(stderr, "as plain text",
		"decryption must be attempted even when KOSLI_API_TOKEN is set but empty")
}

// TestEmptyConfigFileEnvVarFallsBackToDefault pins that KOSLI_CONFIG_FILE set
// to an empty string does not discard the config file. initialize reads it with
// os.LookupEnv, where a variable set to "" reports as present, so the config
// path becomes "" and no config file is loaded at all - taking org, api-token
// and every other configured default with it.
//
// Treating the empty value as unset is deliberate here. Making an empty KOSLI_*
// variable an outright error is the wider change tracked in
// kosli-dev/server#6070.
func (suite *RootCommandTestSuite) TestEmptyConfigFileEnvVarFallsBackToDefault() {
	suite.T().Setenv("KOSLI_CONFIG_FILE", "")

	_, _, _, _, err := executeCommandC("version")

	suite.Require().NoError(err)
	suite.Equal(getConfigFileFlagDefault(), global.ConfigFile,
		"an empty KOSLI_CONFIG_FILE must leave the default config path in place")
}

// TestEmptyValueAfterBoolFlagIsRejected pins that the rejection reaches a real
// command run, not only the token scanner that implements it. attest override
// is used because --new-compliance-status is declared false, so an empty value
// is the case where the recorded verdict is the opposite of the one asked for.
func (suite *RootCommandTestSuite) TestEmptyValueAfterBoolFlagIsRejected() {
	_, _, _, _, err := executeCommandC(
		`attest override --new-compliance-status "" ` +
			`--fingerprint 0000000000000000000000000000000000000000000000000000000000000001 ` +
			`--name foo --reason audit --original-attestation-type generic ` +
			`--flow f --trail t --org demo --api-token DRY_RUN --dry-run`)

	suite.Require().Error(err)
	suite.ErrorContains(err, "--new-compliance-status")
}

// TestInnerMainRejectsEmptyValueAfterBoolFlag drives the empty boolean value
// through innerMain, which is the path a real CLI user takes and which the
// golden test harness bypasses. Without the check there, the rejection would
// hold only in tests.
func (suite *RootCommandTestSuite) TestInnerMainRejectsEmptyValueAfterBoolFlag() {
	args := []string{
		"attest", "override",
		"--new-compliance-status", "",
		"--fingerprint", "0000000000000000000000000000000000000000000000000000000000000001",
		"--name", "foo", "--reason", "audit",
		"--original-attestation-type", "generic",
		"--flow", "f", "--trail", "t",
		"--org", "demo", "--api-token", "DRY_RUN",
	}
	cmd, err := newRootCmd(io.Discard, io.Discard, args)
	suite.Require().NoError(err)

	err = innerMain(cmd, append([]string{"kosli"}, args...))

	suite.Require().Error(err)
	suite.ErrorContains(err, "--new-compliance-status")
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

// TestInnerMainEnrichesError drives a failing command through innerMain (which
// the golden-test harness bypasses) to lock in the wiring: that executedCmd
// from ExecuteC() is the leaf command and that its --flow/--trail flags are
// read into the enriched error. cmd.SetArgs is required because newRootCmd does
// not set the command's args, so ExecuteC would otherwise fall back to os.Args.
func (suite *RootCommandTestSuite) TestInnerMainEnrichesError() {
	args := []string{"attest", "snyk", "--flow", "cyber-dojo", "--trail", "live-snyk-scan"}
	cmd, err := newRootCmd(io.Discard, io.Discard, args)
	suite.Require().NoError(err)
	cmd.SetArgs(args)

	err = innerMain(cmd, append([]string{"kosli"}, args...))
	suite.Require().Error(err)
	suite.ErrorContains(err, "[kosli attest snyk flow=cyber-dojo trail=live-snyk-scan]")
}

// TestExecuteCommandCNormalizesSpaceSeparatedBoolFlag checks that the golden
// test harness normalizes args the same way innerMain does. Without this the
// harness sets args on the command itself and bypasses normalization, so golden
// tests would assert behaviour that real CLI users no longer get.
func (suite *RootCommandTestSuite) TestExecuteCommandCNormalizesSpaceSeparatedBoolFlag() {
	runTestCmd(suite.T(), []cmdTestCase{
		{
			name:   "a bool flag written in the space form is accepted",
			cmd:    "fingerprint testdata/person-schema.json --artifact-type file --debug false",
			golden: "1bef738d0bb1e690500f99a5b57d958caf3a5eb3e00d9012e1f4369fc6812e01\n",
		},
	})
}

// TestInnerMainAcceptsSpaceSeparatedBoolFlag drives a command whose bool flag
// is written in the space form ("--debug false") through innerMain, which is
// where the args are normalized. Without that normalization pflag leaves
// "false" behind as a second positional argument and the command fails with
// "accepts 1 arg(s), received 2". Note that this test deliberately does not
// call cmd.SetArgs: setting the normalized args is innerMain's job.
func (suite *RootCommandTestSuite) TestInnerMainAcceptsSpaceSeparatedBoolFlag() {
	args := []string{
		"fingerprint", "testdata/person-schema.json",
		"--artifact-type", "file",
		"--debug", "false",
	}
	out := new(bytes.Buffer)
	cmd, err := newRootCmd(out, io.Discard, args)
	suite.Require().NoError(err)

	err = innerMain(cmd, append([]string{"kosli"}, args...))

	suite.Require().NoError(err)
	suite.Equal("1bef738d0bb1e690500f99a5b57d958caf3a5eb3e00d9012e1f4369fc6812e01\n", out.String())
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
