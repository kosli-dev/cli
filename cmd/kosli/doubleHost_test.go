package main

import (
	"fmt"
	"strings"
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
			args := []string{"status", host, apiToken, org}

			actual := isDoubleHost(args)

			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsDoubleHost: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func (suite *DoubleHostTestSuite) TestRunDoubleHost() {

	doubledHost := fmt.Sprintf("--host=%s,%s", localHost, localHost)
	doubledApiToken := fmt.Sprintf("--api-token=%s,%s", apiToken, apiToken)
	org := fmt.Sprintf("--org=%s", org)
	doubledArgs := []string{"status", doubledHost, doubledApiToken, org}

	line1 := fmt.Sprintf("[debug] request made to %s/ready and got status 200", localHost)
	line2 := "OK"
	line3 := fmt.Sprintf("[debug] request made to %s/ready and got status 200", localHost)
	line4 := "OK"
	expectedOutputInDebugMode := strings.Join([]string{line1, line2, line3, line4}, "\n")

	for _, t := range []struct {
		name   string
		args   []string
		golden string
	}{
		{
			name:   "only prints primary call output when both calls succeed",
			args:   doubledArgs,
			golden: "OK\n",
		},
		{
			name:   "in debug mode also prints secondary call output",
			args:   append(doubledArgs, " --debug"),
			golden: expectedOutputInDebugMode,
		},
	} {
		suite.Run(t.name, func() {
			// Can't test using runTestCmd() as that calls executeCommandC() which directly calls newRootCmd()
			actual := runDoubleHost(t.args)

			assert.Equal(suite.T(), t.golden, actual, fmt.Sprintf("TestRunDoubleHost: %s , got: %v -- want: %v", t.name, actual, t.golden))
		})
	}
}

func TestDoubleHostTestSuite(t *testing.T) {
	suite.Run(t, new(DoubleHostTestSuite))
}
