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
 * and vice-versa.
 * Note that if you want to run the SonarQube tests, there are a few steps to take:
 * 1. Set the environment variable SONARQUBE to something (value doesn't matter)
 * so we know which test suite to use.
 * 2. Set up an instance of SonarQube (e.g. on localhost), with a project that has been
 * scanned at least once.
 * 3. Replace testdata/sonar/sonarqube/.scannerwork/report-task.txt with the report-task.txt
 * from your sonarqube project (this should be located in a .scannerwork folder in
 * the base directory of your project)
 * 4. In the final two tests, where the CE-task-url flag is provided, replace the current
 * CE-task-url with the one for your project's scan. This can be found in your
 * report-task.txt file. */

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
	// If we have SONARQUBE set (e.g. to true), we're testing SonarQube and therefore should skip the SonarCloud tests
	testHelpers.SkipIfEnvVarSet(suite.T(), []string{"SONARQUBE"})
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
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_SONAR_API_TOKEN", "SONARQUBE"})
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
			cmd:       fmt.Sprintf("attest sonar foo bar --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest sonar testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com  --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest sonar --name foo-s --fingerprint xxxx --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-sonar' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest sonar --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail",
			cmd:    fmt.Sprintf("attest sonar --name bar --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest sonar --name additional --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarCloud with incorrect API token gives error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-api-token xxxx --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: please check your API token is correct and you have the correct permissions in SonarCloud/SonarQube\n",
		},
		{
			wantError: true,
			name:      "if no path to the scannerwork directory is provided and the command is not being run in the same base directory (and no CE task URL is provided), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: report-task.txt not found. Check your working directory is set correctly: open .scannerwork/report-task.txt: no such file or directory\n",
		},
		{
			wantError: true,
			name:      "if incorrect path to the scannerwork directory is provided (and no CE task URL is provided), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: report-task.txt not found. Check your working directory is set correctly: open sonar/.scannerwork/report-task.txt: no such file or directory\n",
		},
		{
			name:   "can retrieve scan results using provided CE task URL and attest them",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --CE-task-url 'https://sonarcloud.io/api/ce/task?id=AZE2jzvUF2N-a1ygL1sM' %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "if incorrect CE task URL given, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --CE-task-url 'https://sonarcloud.io/api/ce/task?id=AZE2jzvUF2N-a1ygL1sm' %s", suite.defaultKosliArguments),
			golden:    "Error: analysis ID not found. Please check the ceTaskURL is correct\n",
		},
		{
			wantError: true,
			name:      "if outdated task given (i.e. we try to get results for an older scan that SonarCloud has deleted), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --CE-task-url 'https://sonarcloud.io/api/ce/task?id=AZERk4uWpzGpahwkB9ac' %s", suite.defaultKosliArguments),
			golden:    "Error: analysis with ID AZERk4xKSYJCvL0vWjio not found. Snapshot has most likely been deleted by Sonar\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *AttestSonarQubeCommandTestSuite) TestAttestSonarQubeCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest sonar foo bar --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest sonar testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com  --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest sonar --name foo-s --fingerprint xxxx --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9' does not exist in flow 'attest-sonar' belonging to organization 'docs-cmd-test-user'\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest sonar --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail",
			cmd:    fmt.Sprintf("attest sonar --name bar --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest sonar --name additional --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "can attest sonar against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "trying to fetch data from SonarQube with incorrect API token gives error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-api-token xxxx --sonar-working-dir testdata/sonar/sonarqube/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: please check your API token is correct and you have the correct permissions in SonarCloud/SonarQube\n",
		},
		{
			wantError: true,
			name:      "if no path to the scannerwork directory is provided and the command is not being run in the same base directory (and no CE task URL is provided), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: report-task.txt not found. Check your working directory is set correctly: open .scannerwork/report-task.txt: no such file or directory\n",
		},
		{
			wantError: true,
			name:      "if incorrect path to the scannerwork directory is provided (and no CE task URL is provided), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir sonar/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: report-task.txt not found. Check your working directory is set correctly: open sonar/.scannerwork/report-task.txt: no such file or directory\n",
		},
		{
			name:   "can retrieve scan results using provided CE task URL and attest them",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --CE-task-url 'http://localhost:9000/api/ce/task?id=9427d05e-a671-4942-95c4-ff0595b6f0fe' %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "if incorrect CE task URL given, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --CE-task-url 'http://localhost:9000/api/ce/task?id=9427d05e-a671-4942-95c4-ff0595b6f0ff' %s", suite.defaultKosliArguments),
			golden:    "Error: analysis ID not found. Please check the ceTaskURL is correct\n",
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
