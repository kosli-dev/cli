package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DoubleHostTestSuite struct {
	suite.Suite
}

const localHost = "http://localhost:8001"
const apiToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"
const org = "docs-cmd-test-user"

func (suite *DoubleHostTestSuite) TestIsDoubleHost() {

	for _, t := range []struct {
		name     string
		host     string
		apiToken string
		want     bool
	}{
		{
			name:     "True when two hosts and two api-tokens",
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     true,
		},
		{
			name:     "False when one host",
			host:     localHost,
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
		{
			name:     "False when three hosts",
			host:     fmt.Sprintf("%s,%s,%s", localHost, localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s", apiToken, apiToken),
			want:     false,
		},
		{
			name:     "False when one api-token",
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: apiToken,
			want:     false,
		},
		{
			name:     "False when three api-tokens",
			host:     fmt.Sprintf("%s,%s", localHost, localHost),
			apiToken: fmt.Sprintf("%s,%s,%s", apiToken, apiToken, apiToken),
			want:     false,
		},
	} {
		suite.Run(t.name, func() {
			host := fmt.Sprintf("--host=%s", t.host)
			apiToken := fmt.Sprintf("--api-token=%s", t.apiToken)
			org := fmt.Sprintf("--org=%s", org)

			defer func(args []string) { os.Args = args }(os.Args)
			os.Args = []string{"status", host, apiToken, org}

			actual := isDoubleHost()

			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsDoubleHost: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *DoubleHostTestSuite) TestRunDoubleHost() {

	doubledArgsCmd := fmt.Sprintf("status --host=%s,%s --api-token=%s,%s --org=%s", localHost, localHost, apiToken, apiToken, org)

	line1 := fmt.Sprintf("[debug] request made to %s/ready and got status 200\n", localHost)
	line2 := "OK\n"
	line3 := fmt.Sprintf("[debug] request made to %s/ready and got status 200\n", localHost)
	line4 := "OK\n"

	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "only prints primary call output when both calls succeed",
			cmd:       doubledArgsCmd,
			golden:    "OK\n",
		},
		{
			wantError: false,
			name:      "in debug mode also prints secondary call output",
			cmd:       doubledArgsCmd + " --debug",
			golden:    line1 + line2 + line3 + line4,
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestDoubleHostTestSuite(t *testing.T) {
	suite.Run(t, new(DoubleHostTestSuite))
}
