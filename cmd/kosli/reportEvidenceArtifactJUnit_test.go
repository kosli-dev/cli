package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArtifactEvidenceJUnitCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	artifactFingerprint   string
	flowName              string
}

func (suite *ArtifactEvidenceJUnitCommandTestSuite) SetupTest() {
	suite.flowName = "junit-test"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlow(suite.flowName, suite.T())
	CreateArtifact(suite.flowName, suite.artifactFingerprint, "FooBar_1", suite.T())
}

func (suite *ArtifactEvidenceJUnitCommandTestSuite) TestArtifactEvidenceJUnitCommandCmd() {
	tests := []cmdTestCase{
		{
			name: "report JUnit test evidence works (using --fingerprint)",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --flow ` + suite.flowName + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report JUnit test evidence works when --evidence-url and --evidence-fingerprint are provided",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --flow ` + suite.flowName + `
			          --build-url example.com --results-dir testdata
					  --evidence-url https://example.com --evidence-fingerprint ` + suite.artifactFingerprint + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report JUnit test evidence with maven-surefire XML that lacks a timestamp on the <testsuite>",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint +
				` --name junit-result --flow ` + suite.flowName +
				` --build-url example.com --results-dir testdata/junit` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report JUnit test evidence works (using --artifact-type)",
			cmd: `report evidence artifact junit testdata/file1 --artifact-type file --name junit-result --flow ` + suite.flowName + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			golden: "junit test evidence is reported to artifact: " + suite.artifactFingerprint + "\n",
		},
		{
			name: "report JUnit test evidence with non-existing results dir",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --flow ` + suite.flowName + `
			          --build-url example.com --results-dir foo` + suite.defaultKosliArguments,
			wantError: true,
		},
		{
			name: "report JUnit test evidence with a results dir that does not contain any results",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result --flow ` + suite.flowName + `
			          --build-url example.com --results-dir testdata/folder1` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: no tests found in testdata/folder1 directory\n",
		},
		{
			name: "report JUnit test evidence with missing name flag",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint + ` --flow ` + suite.flowName + `
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"name\" not set\n",
		},
		{
			name: "report JUnit test evidence with a missing flow",
			cmd: `report evidence artifact junit --fingerprint ` + suite.artifactFingerprint + ` --name junit-result
			          --build-url example.com --results-dir testdata` + suite.defaultKosliArguments,
			wantError: true,
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
	}
	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArtifactEvidenceJUnitCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArtifactEvidenceJUnitCommandTestSuite))
}
