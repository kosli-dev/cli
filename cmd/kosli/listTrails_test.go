package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type ListTrailsCommandTestSuite struct {
	suite.Suite
	flowName              string
	trailName             string
	fingerprint           string
	defaultKosliArguments string
	acmeOrgKosliArguments string
}

func (suite *ListTrailsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      `docs-cmd-test-user`,
		Host:     "http://localhost:8001",
	}

	suite.flowName = "list-trails"
	suite.trailName = "trail-name"
	suite.fingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --host %s --org %s --api-token %s", suite.flowName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.Suite.T())

	global.Org = "acme-org"
	global.ApiToken = "v3OWZiYWu9G2IMQStYg9BcPQUQ88lJNNnTJTNq8jfvmkR1C5wVpHSs7F00JcB5i6OGeUzrKt3CwRq7ndcN4TTfMeo8ASVJ5NdHpZT7DkfRfiFvm8s7GbsIHh2PtiQJYs2UoN13T8DblV5C4oKb6-yWH73h67OhotPlKfVKazR-c"
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	suite.acmeOrgKosliArguments = fmt.Sprintf(" --flow %s --host %s --org %s --api-token %s", suite.flowName, global.Host, global.Org, global.ApiToken)

	CreateArtifactOnTrail(suite.flowName, suite.trailName, "artifact", suite.fingerprint, "artifact-name", suite.Suite.T())
}

func (suite *ListTrailsCommandTestSuite) TestListTrailsCmd() {
	tests := []cmdTestCase{
		{
			name:   "listing trails works when there are trails",
			cmd:    fmt.Sprintf(`list trails %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name:   "listing trails works when there are no trails",
			cmd:    fmt.Sprintf(`list trails %s`, suite.acmeOrgKosliArguments),
			golden: "No trails were found.\n",
		},
		{
			name:       "listing trails with --output json works when there are trails",
			cmd:        fmt.Sprintf(`list trails --output json %s`, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"", "non-empty"}},
		},
		{
			name:       "listing trails with --output json works when there are no trails",
			cmd:        fmt.Sprintf(`list trails --output json %s`, suite.acmeOrgKosliArguments),
			goldenJson: []jsonCheck{{"", "[]"}},
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       fmt.Sprintf(`list trails xxx %s`, suite.defaultKosliArguments),
			golden:    "Error: unknown command \"xxx\" for \"kosli list trails\"\n",
		},
		{
			wantError: true,
			name:      "negative page limit causes an error",
			cmd:       fmt.Sprintf(`list trails --page-limit -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page-limit' has value '-1' which is illegal\n",
		},
		{
			wantError: true,
			name:      "negative page number causes an error",
			cmd:       fmt.Sprintf(`list trails --page -1 %s`, suite.defaultKosliArguments),
			golden:    "Error: flag '--page' has value '-1' which is illegal\n",
		},
		{
			name:   "can list trails with pagination",
			cmd:    fmt.Sprintf(`list trails --page-limit 15 --page 2 %s`, suite.defaultKosliArguments),
			golden: "",
		},
		{
			name: "can list trails that contain an artifact with the provided fingerprint",
			cmd:  fmt.Sprintf(`list trails --fingerprint %s %s`, suite.fingerprint, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestListTrailsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListTrailsCommandTestSuite))
}
