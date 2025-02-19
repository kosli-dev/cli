package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListAttestationTypesCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	acmeOrgKosliArguments string
	attestationName1      string
	attestationName2      string
}

func (suite *ListAttestationTypesCommandTestSuite) SetupTest() {
	suite.attestationName1 = "custom-attestation-type-1"
	suite.attestationName2 = "custom-attestation-type-2"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
	CreateCustomAttestationType(suite.attestationName1, "testdata/person-schema.json", []string{".age > 21"}, suite.Suite.T())
	CreateCustomAttestationType(suite.attestationName2, "testdata/person-schema.json", []string{".age < 25"}, suite.Suite.T())
	CreateCustomAttestationType(suite.attestationName1, "testdata/person-schema.json", []string{".age > 21", ".age < 85"}, suite.Suite.T()) //Make a second version

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

}

func (suite *ListAttestationTypesCommandTestSuite) TestListAttestationTypesCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing custom attestation types works when some exist",
			cmd:    fmt.Sprintf(`list attestation-types %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "listing custom attestation types works when there are none",
			cmd:    fmt.Sprintf(`list attestation-types %s`, suite.acmeOrgKosliArguments),
			golden: "No attestation types were found.\n",
		},
		{
			name:   "listing custom attestation types with --output json works when some exist",
			cmd:    fmt.Sprintf(`list attestation-types --output json %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "listing custom attestation types with --output json works when there are none",
			cmd:    fmt.Sprintf(`list attestation-types --output json %s`, suite.acmeOrgKosliArguments),
			golden: "[]\n",
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list attestation-types xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list attestation-types\"\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListAttestationTypesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListAttestationTypesCommandTestSuite))
}
