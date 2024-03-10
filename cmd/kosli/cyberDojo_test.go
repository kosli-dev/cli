package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type CyberDojoCommandTestSuite struct {
	suite.Suite
}

func (suite *CyberDojoCommandTestSuite) TestIsCyberDojoDoubleHost() {
	for _, t := range []struct {
		name    string
		envVars map[string]string
		want    bool
	}{
		{
			name:    "True when: org is cyber-dojo, host is prod-comma-staging, two api-tokens",
			envVars: defaultEnvVars(),
			want:    true,
		},
		{
			name:    "False when org != cyber-dojo",
			envVars: tweakedEnvVars("KOSLI_ORG", "not-cyber-dojo"),
			want:    false,
		},
		{
			name:    "False when more than one org",
			envVars: tweakedEnvVars("KOSLI_ORG", "cyber-dojo,x"),
			want:    false,
		},
		{
			name:    "False when one host",
			envVars: tweakedEnvVars("KOSLI_HOST", prodHostURL),
			want:    false,
		},
		{
			name:    "False when two host but not prod then staging",
			envVars: tweakedEnvVars("KOSLI_HOST", fmt.Sprintf("%s,%s", stagingHostURL, prodHostURL)),
			want:    false,
		},
		{
			name:    "False when three hosts",
			envVars: tweakedEnvVars("KOSLI_HOST", fmt.Sprintf("%s,%s,%s", stagingHostURL, prodHostURL, "http://a.b.com")),
			want:    false,
		},
		{
			name:    "False when one api-token",
			envVars: tweakedEnvVars("KOSLI_API_TOKEN", "abc"),
			want:    false,
		},
		{
			name:    "False when three api-tokens",
			envVars: tweakedEnvVars("KOSLI_API_TOKEN", "a,b,c"),
			want:    false,
		},
	} {
		suite.Run(t.name, func() {
			suite.setEnvVars(t.envVars)
			actual := IsCyberDojoDoubleHost()
			// clean up
			suite.unsetEnvVars(t.envVars)
			assert.Equal(suite.T(), t.want, actual, fmt.Sprintf("TestIsCyberDojoDoubleHost: %s , got: %v -- want: %v", t.name, actual, t.want))
		})
	}
}

func defaultEnvVars() map[string]string {
	return map[string]string{"KOSLI_ORG": "cyber-dojo", "KOSLI_HOST": fmt.Sprintf("%s,%s", prodHostURL, stagingHostURL), "KOSLI_API_TOKEN": "abc,def"}
}

func tweakedEnvVars(key, value string) map[string]string {
	defaulted := defaultEnvVars()
	defaulted[key] = value
	return defaulted
}

func TestCyberDojoCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CyberDojoCommandTestSuite))
}

// setEnvVars sets env variables
func (suite *CyberDojoCommandTestSuite) setEnvVars(envVars map[string]string) {
	for key, value := range envVars {
		err := os.Setenv(key, value)
		require.NoErrorf(suite.T(), err, "error setting env variable %s", key)
	}
}

// unsetEnvVars unsets env variables
func (suite *CyberDojoCommandTestSuite) unsetEnvVars(envVars map[string]string) {
	for key := range envVars {
		err := os.Unsetenv(key)
		require.NoErrorf(suite.T(), err, "error unsetting env variable %s", key)
	}
}
