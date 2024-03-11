package main

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type DoubleHostTestSuite struct {
	suite.Suite
}

const localHost = "http://localhost:8001"
const apiToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9"

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
			org := fmt.Sprintf("--org=%s", "docs-cmd-test-user")

			defer func(args []string) { os.Args = args }(os.Args)
			os.Args = []string{"status", host, apiToken, org}

			actual := isDoubleHost()

			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsDoubleHost: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}
