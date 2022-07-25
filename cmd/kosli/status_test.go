package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type StatusTestSuite struct {
	suite.Suite
}

func (suite *StatusTestSuite) TestStatusCmd() {
	tests := []cmdTestCase{
		{
			name:   "default",
			cmd:    "status",
			golden: "OK\n",
		}, {
			name:      "assert fail",
			cmd:       "status --assert --host 123",
			wantError: true,
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestStatusTestSuite(t *testing.T) {
	suite.Run(t, new(StatusTestSuite))
}
