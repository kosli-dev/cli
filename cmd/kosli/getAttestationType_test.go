package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetAttestationTypeCommandTestSuite struct {
	suite.Suite
	attestationTypeName   string
	archivedTypeName      string
	defaultKosliArguments string
}

func (suite *GetAttestationTypeCommandTestSuite) SetupTest() {
	suite.attestationTypeName = "custom-attestation-type-1"
	suite.archivedTypeName = "archived-type"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateCustomAttestationType(suite.attestationTypeName, "testdata/person-schema.json", []string{".age > 21"}, suite.T())
	CreateCustomAttestationType(suite.archivedTypeName, "testdata/person-schema.json", []string{".age < 21"}, suite.T())
	ArchiveCustomAttestationType(suite.archivedTypeName, suite.T())
}

func (suite *GetAttestationTypeCommandTestSuite) TestGetAttestationTypeCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "getting a non existing attestation type fails",
			cmd:       fmt.Sprintf(`get attestation-type foo %s`, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Custom attestation type 'foo' does not exist for org '%s'. \n", global.Org),
		},
		{
			wantError: true,
			name:      "getting an archived attestation type fails",
			cmd:       fmt.Sprintf(`get attestation-type %s %s`, suite.archivedTypeName, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Custom attestation type 'archived-type' is archived for org '%s'\n", global.Org),
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get attestation-type %s xxx %s`, suite.attestationTypeName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`get attestation-type %s --org orgX`, suite.attestationTypeName),
			golden:    "Error: --api-token is not set\nUsage: kosli get attestation-type TYPE-NAME [flags]\n",
		},
		{
			name:       "getting an existing attestation type works",
			cmd:        fmt.Sprintf(`get attestation-type %s %s`, suite.attestationTypeName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-attestation-type.txt",
		},
		{
			name: "getting an existing attestation type with --output json works",
			cmd:  fmt.Sprintf(`get attestation-type %s --output json %s`, suite.attestationTypeName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetAttestationTypeCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetAttestationTypeCommandTestSuite))
}
