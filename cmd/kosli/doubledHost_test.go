package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DoubledHostTestSuite struct {
	suite.Suite
}

const localHost = "http://localhost:8001"
const apiToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
const orgName = "docs-cmd-test-user"

func (suite *DoubledHostTestSuite) TestIsDoubledHost() {

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
			name:     "False when one host",
			args:     []string{"status"},
			host:     localHost,
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
		{
			name:     "False when three hosts",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s,%s", localHost, localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
		{
			name:     "False when one api-token",
			args:     []string{"status"},
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: apiToken,
			want:     false,
		},
		{
			name:     "False when three api-tokens",
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

			actual := isDoubledHost(args)

			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsDoubledHost: %s\n\texpected: '%v'\n\t--actual: '%v'\n", t.name, t.want, actual))
		})
	}
}

func (suite *DoubledHostTestSuite) TestRunDoubledHost() {

	doubledHost := fmt.Sprintf("--host=%s,%s", localHost, localHost)
	doubledApiToken := fmt.Sprintf("--api-token=%s,%s", apiToken, apiToken)
	org := fmt.Sprintf("--org=%s", orgName)

	doubledArgs := func(args []string) []string {
		return append(args, doubledHost, doubledApiToken, org)
	}

	for _, t := range []struct {
		name   string
		args   []string
		stdOut []string
		err    error
	}{
		{
			name:   "only returns primary call output when both calls succeed",
			args:   doubledArgs([]string{"kosli", "status"}),
			stdOut: []string{"OK", ""},
			err:    error(nil),
		},
		{
			name:   "in debug mode also returns secondary call output",
			args:   doubledArgs([]string{"kosli", "status", "--debug"}),
			stdOut: StatusDebugLines(),
			err:    error(nil),
		},
		{
			name:   "--help prints output once",
			args:   doubledArgs([]string{"kosli", "status", "--help"}),
			stdOut: HelpStatusLines(),
			err:    error(nil),
		},
		{
			name:   "unknown-command prints output once",
			args:   doubledArgs([]string{"kosli", "unknown-command"}),
			stdOut: UsageLines(),
			err:    error(nil),
		},
		{
			name:   "bare [kosli] call prints output once",
			args:   doubledArgs([]string{"kosli"}),
			stdOut: UsageLines(),
			err:    error(nil),
		},
		// {
		// 	name:   "bad-flag never gets to call runDoubledHost() because isDoubledHost() returns false",
		// 	args:   doubledArgs([]string{"kosli", "status", "--bad-flag"}),
		// 	stdOut: BadFlagLines(),
		// 	err:    error(nil),
		// },
	} {
		output, err := runDoubledHost(t.args)

		assert.Equal(suite.T(), t.err, err, fmt.Sprintf("TestRunDoubleHost: %s\n\texpected: '%v'\n\t--actual: '%v'\n", t.name, t.err, err))

		lines := strings.Split(output, "\n")
		d := diff(t.stdOut, lines)
		assert.Equal(suite.T(), "", d, fmt.Sprintf("TestRunDoubleHost: %s\n%s\n", t.name, d))
	}
}

func TestDoubledHostTestSuite(t *testing.T) {
	suite.Run(t, new(DoubledHostTestSuite))
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

func StatusDebugLines() []string {
	return []string{
		fmt.Sprintf("[debug] request made to %s/ready and got status 200", localHost),
		"OK",
		"",
		fmt.Sprintf("[debug] [%s]", localHost),
		fmt.Sprintf("[debug] request made to %s/ready and got status 200", localHost),
		"OK",
		"",
	}
}

func HelpStatusLines() []string {
	return []string{
		"Check the status of a Kosli server.  ",
		"The status is logged and the command always exits with 0 exit code.  ",
		"If you like to assert the Kosli server status, you can use the --assert flag or the \"kosli assert status\" command.",
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

func UsageLines() []string {
	return []string{
		"The Kosli CLI.",
		"",
		"Environment variables:",
		"You can set any flag from an environment variable by capitalizing it in snake case and adding the KOSLI_ prefix.",
		"For example, to set --api-token from an environment variable, you can export KOSLI_API_TOKEN=YOUR_API_TOKEN.",
		"",
		"Setting the API token to DRY_RUN sets the --dry-run flag.",
		"",
		"Usage:",
		"  kosli [command]",
		"",
		"Available Commands:",
		"  allow       All Kosli allow commands.",
		"  archive     All Kosli archive commands.",
		"  assert      All Kosli assertion commands. Return non-zero exit code if the assertion fails.",
		"  attest      All Kosli attest commands.",
		"  begin       All Kosli begin commands.",
		"  completion  Generate completion script",
		"  create      All Kosli create commands.",
		"  diff        All Kosli diff commands.",
		"  disable     Kosli disable commands.",
		"  enable      Kosli enable commands.",
		"  fingerprint Calculate the SHA256 fingerprint of an artifact.",
		"  get         All Kosli get commands.",
		"  help        Help about any command",
		"  list        All Kosli list commands.",
		"  log         All Kosli log commands.",
		"  rename      All Kosli rename commands.",
		"  report      All Kosli report commands.",
		"  request     All Kosli request commands.",
		"  search      Search for a git commit or an artifact fingerprint in Kosli.  ",
		"  snapshot    All Kosli snapshot commands.",
		"  status      Check the status of a Kosli server.  ",
		"  version     Print the version of a Kosli CLI.  ",
		"",
		"Flags:",
		"  -a, --api-token string      The Kosli API token.",
		"  -c, --config-file string    [optional] The Kosli config file path. (default \"kosli\")",
		"      --debug                 [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)",
		"  -h, --help                  help for kosli",
		"  -H, --host string           [defaulted] The Kosli endpoint. (default \"https://app.kosli.com\")",
		"  -r, --max-api-retries int   [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)",
		"      --org string            The Kosli organization.",
		"",
		"Use \"kosli [command] --help\" for more information about a command.",
		"",
	}
}

func BareKosliLines() []string {
	return []string{}
}
