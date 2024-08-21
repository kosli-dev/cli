package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

/* The attest sonar command is used to attest scans from both SonarCloud and SonarQube.
 * The sonar API token for SonarCloud and SonarQube will always be different, so we need
 * to have a separate test suite for each version of the command. This means we can easily
 * skip the SonarQube tests when we're testing SonarCloud (with the SonarCloud API token),
 * and vice-versa. */

type AttestSonarCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

type AttestSonarQubeCommandTestSuite struct {
	flowName            string
	trailName           string
	artifactFingerprint string
	suite.Suite
	defaultKosliArguments string
}

func (suite *AttestSonarCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_SONAR_API_TOKEN"})
	// If we have the sonarqube url set, we're testing SonarQube and therefore should skip the SonarCloud tests
	testHelpers.SkipIfEnvVarSet(suite.T(), []string{"KOSLI_SONARQUBE_URL"})
	suite.flowName = "attest-sonar"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestSonarQubeCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_SONAR_API_TOKEN", "KOSLI_SONARQUBE_URL"})
	suite.flowName = "attest-sonar"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.T())
}

func (suite *AttestSonarCommandTestSuite) TestAttestSonarCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest sonar foo bar --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest sonar testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url example.com  --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest sonar --name foo-s --fingerprint xxxx --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-sonar' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest sonar --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail",
			cmd:    fmt.Sprintf("attest sonar --name bar --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest sonar --name additional --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url example.com --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarCloud with incorrect API token gives error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url example.com --sonar-api-token xxxx --sonar-working-dir testdata/sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: please check your API token is correct and you have the correct permissions in SonarCloud/SonarQube\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *AttestSonarQubeCommandTestSuite) TestAttestSonarQubeCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest sonar foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest sonar testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest sonar --name foo-s --fingerprint xxxx --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key test  %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-sonar' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url example.com --sonar-project-key test  %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url example.com --sonar-project-key test  %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest sonar --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key test %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail",
			cmd:    fmt.Sprintf("attest sonar --name bar --commit HEAD --origin-url example.com --sonar-project-key test %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest sonar --name additional --commit HEAD --origin-url example.com --sonar-project-key test %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url example.com --sonar-project-key test %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarQube with incorrect SonarQube URL gives error",
			cmd:       fmt.Sprintf("attest sonar --name foo --commit HEAD --origin-url example.com --sonar-project-key test --sonarqube-url example.com/ %s", suite.defaultKosliArguments),
			golden:    "Error: Incorrect SonarQube URL\n",
		},
		{
			name:   "can attest sonar with sonarqube url against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url example.com --sonar-project-key test %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarQube with incorrect API token gives error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url example.com --sonar-project-key test --sonar-api-token xxxx %s", suite.defaultKosliArguments),
			golden:    "Error: Incorrect API token or SonarQube URL\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarQube for a non-existent project gives error",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key cyber-dojo_differ  %s", suite.defaultKosliArguments),
			golden:    "Error: Component key 'cyber-dojo_differ' not found\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarQube for a non-existent branch gives error",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key test --branch-name xx %s", suite.defaultKosliArguments),
			golden:    "Error: Component 'test' on branch 'xx' not found\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarQube for a non-existent pull-request gives error",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url example.com --sonar-project-key test --pull-request-id 5 %s", suite.defaultKosliArguments),
			golden:    "Error: Component 'test' of pull request '5' not found\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestSonarCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSonarCommandTestSuite))
	suite.Run(t, new(AttestSonarQubeCommandTestSuite))
}
