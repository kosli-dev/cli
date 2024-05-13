package main

import (
	"testing"

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
