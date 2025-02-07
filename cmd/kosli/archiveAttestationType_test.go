package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ArchiveAttestationTypeCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	attestationTypeName   string
}

func (suite *ArchiveAttestationTypeCommandTestSuite) SetupTest() {
	suite.attestationTypeName = "archive-attestation-type"

	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateCustomAttestationType(suite.attestationTypeName, "testdata/person-schema.json", []string{".age > 21"}, suite.T())
}

func (suite *ArchiveAttestationTypeCommandTestSuite) TestArchiveAttestationTypeCmd() {
	tests := []cmdTestCase{
		{
			name:   "can archive custom attestation type",
			cmd:    fmt.Sprintf(`archive attestation-type %s %s`, suite.attestationTypeName, suite.defaultKosliArguments),
			golden: "Custom attestation type archive-attestation-type was archived\n",
		},
		{
			wantError: true,
			name:      "archiving non-existing custom attestation type fails",
			cmd:       fmt.Sprintf(`archive attestation-type non-existing %s`, suite.defaultKosliArguments),
			golden:    "Error: Custom attestation type 'non-existing' does not exist for org 'docs-cmd-test-user'. \n",
		},
		{
			wantError: true,
			name:      "archive attestation-type fails when 2 args are provided",
			cmd:       fmt.Sprintf(`archive attestation-type %s arg2 %s`, suite.attestationTypeName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "archive attestation-type fails when no args are provided",
			cmd:       fmt.Sprintf(`archive attestation-type %s`, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestArchiveAttestationTypeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveAttestationTypeCommandTestSuite))
}
