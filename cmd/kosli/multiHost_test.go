package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type MultiHostTestSuite struct {
	suite.Suite
}

const localHost = "http://localhost:8001"
const apiToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
const orgName = "docs-cmd-test-user"

func (suite *MultiHostTestSuite) TestIsMultiHost() {

	for _, t := range []struct {
		name     string
		args     []string
		host     string
		apiToken string
		want     bool
	}{
		{
			name:     "True when two hosts and two api-tokens",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     true,
		},
		{
			name:     "True when three hosts and three api-tokens",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s,%s", localHost, localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s,%s", apiToken, apiToken, apiToken),
			want:     true,
		},
		{
			name:     "False when one host and two api-tokens",
			args:     []string{"status"},
			host:     localHost,
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
		{
			name:     "False when three hosts and two api-tokens",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s,%s", localHost, localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
		{
			name:     "False when one api-token and two hostnames",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: apiToken,
			want:     false,
		},
		{
			name:     "False when three api-tokens and two hostnames",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s,%s", apiToken, apiToken, apiToken),
			want:     false,
		},
		{
			name:     "True when unknown command",
			args:     []string{"not-a-command"},
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     true,
		},
		{
			name:     "False when unknown flag",
			args:     []string{"status", "--not-a-flag"},
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
	} {
		suite.Run(t.name, func() {
			host := fmt.Sprintf("--host=%s", t.host)
			apiToken := fmt.Sprintf("--api-token=%s", t.apiToken)
			org := fmt.Sprintf("--org=%s", orgName)
			args := append(t.args, host, apiToken, org)

			defer func(original []string) { os.Args = original }(os.Args)
			os.Args = args
			actual := isMultiHost()

			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsMultiHost: %s\n\texpected: '%v'\n\t--actual: '%v'\n", t.name, t.want, actual))
		})
	}
}

func (suite *MultiHostTestSuite) TestRunDoubledHost() {

	MultiHost := fmt.Sprintf("--host=%s,%s", localHost, localHost)
	doubledApiToken := fmt.Sprintf("--api-token=%s,%s", apiToken, apiToken)
	org := fmt.Sprintf("--org=%s", orgName)

	doubledArgs := func(args []string) []string {
		return append(args, MultiHost, doubledApiToken, org)
	}

	for _, t := range []struct {
		name   string
		args   []string
		stdOut []string
		err    error
	}{
		{
			name:   "only returns primary call output when both (2) calls succeed",
			args:   doubledArgs([]string{"kosli", "status"}),
			stdOut: []string{"OK", ""},
			err:    error(nil),
		},
	} {
		defer func(original []string) { os.Args = original }(os.Args)
		os.Args = t.args
		output, err := runMultiHost(t.args)

		assert.Equal(suite.T(), t.err, err, fmt.Sprintf("TestRunDoubleHost: %s\n\texpected: '%v'\n\t--actual: '%v'\n", t.name, t.err, err))

		lines := strings.Split(output, "\n")
		d := diff(t.stdOut, lines)
		assert.Equal(suite.T(), "", d, fmt.Sprintf("TestRunDoubleHost: %s\n%s\n", t.name, d))
	}
}

// TestRunMultiHostRefusesAnEmptyHostInTheList pins that an empty element in the
// --host list is refused. This file splits the list rather than pflag, so what
// catches the empty element is the per-host call: each element is appended as
// that call's own --host, and an empty one is refused as it is set. The error
// names the host it came from, which for this element is empty.
func (suite *MultiHostTestSuite) TestRunMultiHostRefusesAnEmptyHostInTheList() {
	args := []string{
		"kosli", "status",
		fmt.Sprintf("--host=%s,,%s", localHost, localHost),
		fmt.Sprintf("--api-token=%s,%s,%s", apiToken, apiToken, apiToken),
		fmt.Sprintf("--org=%s", orgName),
	}

	defer func(original []string) { os.Args = original }(os.Args)
	os.Args = args
	_, err := runMultiHost(args)

	assert.EqualError(suite.T(), err,
		"\n[]\n[kosli status] flag '--host' was given an empty value")
}

// TestRunDoubledHostAcceptsSpaceSeparatedBoolFlag checks that a boolean flag
// written in the space form is honoured on the multi-host path too. --debug is
// false here, so the output must be the bare fingerprint: no per-host [debug]
// prefix and no secondary-call output. fingerprint is used because it needs no
// server, so the two hosts are never contacted.
func (suite *MultiHostTestSuite) TestRunDoubledHostAcceptsSpaceSeparatedBoolFlag() {
	args := []string{
		"kosli", "fingerprint", "testdata/person-schema.json",
		"--artifact-type", "file",
		"--debug", "false",
		fmt.Sprintf("--host=%s,%s", localHost, localHost),
		fmt.Sprintf("--api-token=%s,%s", apiToken, apiToken),
		fmt.Sprintf("--org=%s", orgName),
	}
	want := []string{"1bef738d0bb1e690500f99a5b57d958caf3a5eb3e00d9012e1f4369fc6812e01", ""}

	defer func(original []string) { os.Args = original }(os.Args)
	os.Args = args
	output, err := runMultiHost(args)

	assert.Equal(suite.T(), error(nil), err)
	assert.Equal(suite.T(), "", diff(want, strings.Split(output, "\n")))
}

func (suite *MultiHostTestSuite) TestRunTripledHost() {

	multiHost := fmt.Sprintf("--host=%s,%s,%s", localHost, localHost, localHost)
	multiApiToken := fmt.Sprintf("--api-token=%s,%s,%s", apiToken, apiToken, apiToken)
	org := fmt.Sprintf("--org=%s", orgName)

	tripledArgs := func(args []string) []string {
		return append(args, multiHost, multiApiToken, org)
	}

	for _, t := range []struct {
		name   string
		args   []string
		stdOut []string
		err    error
	}{
		{
			name:   "only returns primary call output when all three calls succeed",
			args:   tripledArgs([]string{"kosli", "status"}),
			stdOut: []string{"OK", ""},
			err:    error(nil),
		},
	} {
		defer func(original []string) { os.Args = original }(os.Args)
		os.Args = t.args
		output, err := runMultiHost(t.args)

		assert.Equal(suite.T(), t.err, err, fmt.Sprintf("TestRunTripledHost: %s\n\texpected: '%v'\n\t--actual: '%v'\n", t.name, t.err, err))

		lines := strings.Split(output, "\n")
		d := diff(t.stdOut, lines)
		assert.Equal(suite.T(), "", d, fmt.Sprintf("TestRunTripledHost: %s\n%s\n", t.name, d))
	}
}

func TestMultiHostTestSuite(t *testing.T) {
	suite.Run(t, new(MultiHostTestSuite))
}

// func StatusDebugLines() []string {
// 	return []string{
// 		fmt.Sprintf("[debug] request made to %s/ready and got status 200", localHost),
// 		"OK",
// 		"",
// 		fmt.Sprintf("[debug] [%s]", localHost),
// 		fmt.Sprintf("[debug] request made to %s/ready and got status 200", localHost),
// 		"OK",
// 		"",
// 	}
// }

func HelpStatusLines() []string {
	return []string{
		"Check the status of a Kosli server.  ",
		"The status is logged and the command always exits with 0 exit code.  ",
		"If you like to assert the Kosli server status, you can use the ^--assert^ flag or the \"kosli assert status\" command.",
		"",
		"Usage:",
		"  kosli status [flags]",
		"",
		"Flags:",
		"      --assert   [optional] Exit with non-zero code if Kosli server is not responding.",
		"  -h, --help     help for status",
		"",
		"Global Flags:",
		"  -a, --api-token string      The Kosli API token.",
		"  -c, --config-file string    [optional] The Kosli config file path. (default \"kosli\")",
		"      --debug                 [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)",
		"  -H, --host string           [defaulted] The Kosli endpoint. (default \"https://app.kosli.com\")",
		"  -r, --max-api-retries int   [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)",
		"      --org string            The Kosli organization.",
		"",
	}
}

func diff(expect []string, actual []string) string {
	if len(expect) != len(actual) {
		return fmt.Sprintf("len(expect)==%v, len(actual)==%v\n", len(expect), len(actual))
	}
	for i := 0; i < len(expect); i++ {
		e := expect[i]
		a := actual[i]
		d := diffLine(i, e, a)
		if d != "" {
			return d
		}
	}
	return ""
}

func diffLine(n int, expect string, actual string) string {
	m := max(len(expect), len(actual))
	for i := 0; i < m; i++ {
		e := charAt(expect, i)
		a := charAt(actual, i)
		if e != a {
			msg := []string{
				fmt.Sprintf("line: %v", n),
				fmt.Sprintf("expect: '%v'", expect),
				fmt.Sprintf("actual: '%v'", actual),
				fmt.Sprintf("len(expect): %v", len(expect)),
				fmt.Sprintf("len(actual): %v", len(actual)),
				fmt.Sprintf("expect[%v]: %v", i, e),
				fmt.Sprintf("actual[%v]: %v", i, a),
			}
			return strings.Join(msg, "\n")
		}
	}
	return ""
}

func charAt(s string, n int) string {
	if n >= len(s) {
		return "nil"
	}
	c := s[n]
	if c == '\t' {
		return "TAB"
	}
	if c == '\n' {
		return "NL"
	}
	return fmt.Sprintf("%v", c)
}

// {
// 	name:   "in debug mode also returns secondary call output",
// 	args:   doubledArgs([]string{"kosli", "status", "--debug"}),
// 	stdOut: StatusDebugLines(),
// 	err:    error(nil),
// },
// {
// 	name:   "--help prints output once",
// 	args:   doubledArgs([]string{"kosli", "status", "--help"}),
// 	stdOut: HelpStatusLines(),
// 	err:    error(nil),
// },
// {
// 	name:   "bad-flag never gets to call runMultiHost() because isMultiHost() returns false",
// 	args:   doubledArgs([]string{"kosli", "status", "--bad-flag"}),
// 	stdOut: BadFlagLines(),
// 	err:    error(nil),
// },
