package main

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type VersionTestSuite struct {
	suite.Suite
}

func (suite *VersionTestSuite) TestVersionCmd() {
	tests := []cmdTestCase{
		{
			name:   "default",
			cmd:    "version",
			golden: fmt.Sprintf("version.BuildInfo{Version:\"main\", GitCommit:\"\", GitTreeState:\"\", GoVersion:\"%s\"}\n", runtime.Version()),
		}, {
			name:   "short",
			cmd:    "version --short",
			golden: "main\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
