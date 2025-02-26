package main

import (
	"fmt"
	"testing"

	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

/* The attest sonar command is used to attest scans from both SonarQube Server and SonarQube Cloud.
 * The sonar API token for SonarQube Server and Cloud will always be different, so we need
 * to have a separate test suite for each version of the command. This means we can easily
 * skip the SonarQube Server tests when we're testing SonarQube Cloud (with the SonarQube Cloud API token),
 * and vice-versa.
 *
 * Note that SonarQube Cloud regularly deletes older scans (see https://docs.sonarsource.com/sonarcloud/digging-deeper/housekeeping/ )
 * so the current report-task.txt files and the revisions used in the tests may not be valid in the future.
 * If/when this happens, they will need to be updated.
 *
 * Note also that if you want to run the SonarQube Server tests, there are a few steps to take:
 * 1. Set the environment variable SONARQUBE to something (value doesn't matter)
 * so we know which test suite to use.
 * 2. Set up an instance of SonarQube Server (or SonarQube Community on localhost), with a project that has been
 * scanned at least once.
 * 3. Replace testdata/sonar/sonarqube/.scannerwork/report-task.txt with the report-task.txt
 * from your SonarQube project (this should be located in a .scannerwork folder in
 * the base directory of your project) */

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
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_SONAR_API_TOKEN"})
	// If we have SONARQUBE set (e.g. to true), we're testing SonarQube Server and therefore should skip the SonarQube Cloud tests
	testHelpers.SkipIfEnvVarSet(suite.Suite.T(), []string{"SONARQUBE"})
	suite.flowName = "attest-sonar"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.Suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.Suite.T())
}

func (suite *AttestSonarQubeCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.Suite.T(), []string{"KOSLI_SONAR_API_TOKEN", "SONARQUBE"})
	suite.flowName = "attest-sonar"
	suite.trailName = "test-123"
	suite.artifactFingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.Suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.Suite.T())
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.artifactFingerprint, "file1", suite.Suite.T())
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
			golden:    "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-sonar\" belonging to organization \"docs-cmd-test-user\"\n",
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
			golden:    "Error: please check your API token is correct and you have the correct permissions in SonarQube\n",
		},
		{
			wantError: true,
			name:      "if no path to the scannerwork directory is provided and the command is not being run in the same base directory, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: open .scannerwork/report-task.txt: no such file or directory. Check your working directory is set correctly. Alternatively provide the project key and revision for the scan to attest\n",
		},
		{
			name:   "can retrieve scan results using project key and revision and attest them",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --sonar-revision 38f3dc8b63abb632ac94a12b3f818b49f8047fa1 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "if report-task.txt file found, we don't use the sonar-project-key, sonar-revision or sonar-server-url flags",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork --sonar-project-key anyKey --sonar-revision anyRevision --sonar-server-url http://example.com %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "if outdated task given (i.e. we try to get results for an older scan that SonarCloud has deleted), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork-old %s", suite.defaultKosliArguments),
			golden:    "Error: analysis with ID AZERk4xKSYJCvL0vWjio not found. Snapshot may have been deleted by SonarQube\n",
		},
		{
			wantError: true,
			name:      "if incorrect revision given (or the scan for the given revision has been deleted by SonarCloud)",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --sonar-revision b4d1053f2aac18c9fb4b9a289a8289199c932e12 %s", suite.defaultKosliArguments),
			golden:    "Error: analysis for revision b4d1053f2aac18c9fb4b9a289a8289199c932e12 of project cyber-dojo_differ not found. Check the revision is correct. Snapshot may also have been deleted by SonarQube\n",
		},
		{
			wantError: true,
			name:      "if incorrect project key given, we get an error message from Sonar",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo-differ --sonar-revision 38f3dc8b63abb632ac94a12b3f818b49f8047fa1 %s", suite.defaultKosliArguments),
			golden:    "Error: sonar error: Component key 'cyber-dojo-differ' not found\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
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
			golden:    "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-sonar\" belonging to organization \"docs-cmd-test-user\"\n",
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
			golden:    "Error: please check your API token is correct and you have the correct permissions in SonarQube\n",
		},
		{
			wantError: true,
			name:      "if no path to the scannerwork directory is provided and the command is not being run in the same base directory, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: open .scannerwork/report-task.txt: no such file or directory. Check your working directory is set correctly. Alternatively provide the project key and revision for the scan to attest\n",
		},
		{
			name:   "can retrieve scan results using project key and revision and attest them",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-server-url http://localhost:9000 --sonar-project-key test5 --sonar-revision 8e6f9489e5f2ddf8e719b503e374975e8b607fd1 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "if incorrect revision given, give an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-server-url http://localhost:9000 --sonar-project-key test5 --sonar-revision 8e6f9489e5f2ddf8e719b503e374975e8b607fd2 %s", suite.defaultKosliArguments),
			golden:    "Error: analysis for revision 8e6f9489e5f2ddf8e719b503e374975e8b607fd2 of project test5 not found. Check the revision is correct. Snapshot may also have been deleted by SonarQube\n",
		},
		{
			wantError: true,
			name:      "if incorrect project key given, we get an error message from Sonar",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-server-url http://localhost:9000 --sonar-project-key test99 --sonar-revision 38f3dc8b63abb632ac94a12b3f818b49f8047fa1 %s", suite.defaultKosliArguments),
			golden:    "Error: sonar error: Component key 'test99' not found\n",
		},
		{
			wantError: true,
			name:      "if incorrect sonarqube server url given, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-server-url http://example.com --sonar-project-key test99 --sonar-revision 38f3dc8b63abb632ac94a12b3f818b49f8047fa1 %s", suite.defaultKosliArguments),
			golden:    "Error: please check your API token and SonarQube server URL are correct and you have the correct permissions in SonarQube\n",
		},
		{
			name:   "if report-task.txt file found, we don't use the sonar-project-key, sonar-revision or sonar-server-url flags",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork --sonar-project-key anyKey --sonar-revision anyRevision --sonar-server-url http://example.com %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.Suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestSonarCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSonarCommandTestSuite))
	suite.Run(t, new(AttestSonarQubeCommandTestSuite))
}
