package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArtifactEvidenceJUnitCommandTestSuite struct {
	suite.Suite
}

func (suite *ArtifactEvidenceJUnitCommandTestSuite) TestArtifactEvidenceJUnitCommandCmd() {

	defaultKosliArguments := " -H http://localhost:8001 --owner docs-cmd-test-user -a eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY"

	tests := []cmdTestCase{
		{
			name:   "report JUnit test evidence",
			cmd:    "pipeline artifact report evidence junit --sha256 847411c6124e719a4e8da2550ac5c116b7ff930493ce8a061486b48db8a5aaa0 --name junit-result --pipeline FooBarPipeline --build-url example.com --results-dir testdata --dry-run" + defaultKosliArguments,
			golden: "",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceJUnitCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceJUnitCommandTestSuite))
}
