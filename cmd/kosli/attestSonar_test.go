package main

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
 * If/when this happens, they will need to be updated. There is an instruction file at testdata/sonar/update-sonarqube-test-data.txt
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
	prScannerWorkDir    string
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
	// If we have SONARQUBE set (e.g. to true), we're testing SonarQube Server and therefore should skip the SonarQube Cloud tests
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

	suite.prScannerWorkDir = createPRScannerWorkDir(suite.T())
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
			name:      "01 fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest sonar foo bar --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "02 fails when both --fingerprint and --artifact-type",
			cmd:       fmt.Sprintf("attest sonar testdata/file1 --fingerprint xxxx --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com  --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --fingerprint, --artifact-type is allowed\n",
		},
		{
			wantError: true,
			name:      "03 fails when --fingerprint is not valid",
			cmd:       fmt.Sprintf("attest sonar --name foo-s --fingerprint xxxx --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest sonar [IMAGE-NAME | FILE-PATH | DIR-PATH] [flags]\n",
		},
		{
			wantError: true,
			name:      "04 attesting against an artifact that does not exist fails",
			cmd:       fmt.Sprintf("attest sonar --fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint 1234e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 does not exist in trail \"test-123\" of flow \"attest-sonar\" belonging to organization \"docs-cmd-test-user\"\n",
		},
		{
			name:   "05 can attest sonar against an artifact using artifact name and --artifact-type",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "06 can attest sonar against an artifact using artifact name and --artifact-type when --name does not exist in the trail template",
			cmd:    fmt.Sprintf("attest sonar testdata/file1 --artifact-type file --name bar --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "07 can attest sonar against an artifact using --fingerprint",
			cmd:    fmt.Sprintf("attest sonar --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "08 can attest sonar against a trail",
			cmd:    fmt.Sprintf("attest sonar --name bar --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'bar' is reported to trail: test-123\n",
		},
		{
			name:   "09 can attest sonar against a trail when name is not found in the trail template",
			cmd:    fmt.Sprintf("attest sonar --name additional --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'additional' is reported to trail: test-123\n",
		},
		{
			name:   "10 can attest sonar against an artifact it is created using dot syntax in --name",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "11 trying to fetch data from SonarCloud with incorrect API token gives error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-api-token xxxx --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork %s", suite.defaultKosliArguments),
			golden:    "Error: please check your API token is correct and you have the correct permissions in SonarQube\n",
		},
		{
			wantError: true,
			name:      "12 if no path to the scannerwork directory is provided and the command is not being run in the same base directory, and project key/revision/pull request not provided, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: open .scannerwork/report-task.txt: no such file or directory. Check your working directory is set correctly. Alternatively provide the project key and either revision or pull-request ID for the scan to attest\n",
		},
		{
			name:   "13 can retrieve scan results using project key and revision and attest them",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --sonar-revision 38f3dc8b63abb632ac94a12b3f818b49f8047fa1 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "14 if report-task.txt file found, we don't use the sonar-project-key, sonar-revision or sonar-server-url flags",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork --sonar-project-key anyKey --sonar-revision anyRevision --sonar-server-url http://example.com %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "15 if outdated task given (i.e. we try to get results for an older scan that SonarCloud has deleted), we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork-old %s", suite.defaultKosliArguments),
			golden:    "Error: No activity found for task 'AZERk4uWpzGpahwkB9ac' on https://sonarcloud.io. \nSonarQube may be experiencing problems, please check https://status.sonarqube.com/ and try again later. \nOtherwise if you are attesting an older scan, the snapshot may have been deleted by SonarQube\n",
		},
		{
			wantError: true,
			name:      "16 if incorrect revision given (or the scan for the given revision has been deleted by SonarCloud)",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --sonar-revision b4d1053f2aac18c9fb4b9a289a8289199c932e12 %s", suite.defaultKosliArguments),
			golden:    "Error: analysis for revision b4d1053f2aac18c9fb4b9a289a8289199c932e12 of project cyber-dojo_differ not found. Check the revision is correct. \nThe scan may still be being processed by SonarQube, try again later.\n Otherwise if you are attesting an older scan, the snapshot may also have been deleted by SonarQube\n",
		},
		{
			wantError: true,
			name:      "17 if incorrect project key given, we get an error message from Sonar",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo-differ --sonar-revision 38f3dc8b63abb632ac94a12b3f818b49f8047fa1 %s", suite.defaultKosliArguments),
			golden:    "Error: SonarQube error: Component key 'cyber-dojo-differ' not found\n",
		},
		{
			name:   "18 can retrieve scan results using report-task.txt file and pull-request flag",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork --pull-request 359 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "19 providing an incorrect pull-request ID gives an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarcloud/.scannerwork --pull-request 1 %s", suite.defaultKosliArguments),
			golden:    "Error: pull request 1 not found for project cyber-dojo_differ on https://sonarcloud.io\n",
		},
		{
			name:   "20 can retrieve scan results using project key and pull-request flag",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --pull-request 359 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "21 can attest sonar for a pull request scan using report-task.txt without --pull-request flag",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir %s %s", suite.prScannerWorkDir, suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			name:   "22 can attest sonar using --sonar-ce-task-url without report-task.txt",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-ce-task-url https://sonarcloud.io/api/ce/task?id=AZrs5eywBfkZKeU0sde9 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "23 if no report-task.txt is available, and project key is not provided, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com %s", suite.defaultKosliArguments),
			golden:    "Error: open .scannerwork/report-task.txt: no such file or directory. Check your working directory is set correctly. Alternatively provide the project key and either revision or pull-request ID for the scan to attest\n",
		},
		{
			wantError: true,
			name:      "24 if no report-task.txt is available, and neither revision or pull-request is given, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ %s", suite.defaultKosliArguments),
			golden:    "Error: open .scannerwork/report-task.txt: no such file or directory. Check your working directory is set correctly. Alternatively provide the project key and either revision or pull-request ID for the scan to attest\n",
		},
		{
			wantError: true,
			name:      "25 can't provide both revision and pull-request",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --sonar-revision xxx --pull-request 5 %s", suite.defaultKosliArguments),
			golden:    "Error: only one of --sonar-revision, --pull-request is allowed\n",
		},
		{
			name:   "26 can attest sonar for a pull request scan using --sonar-ce-task-url",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-ce-task-url https://sonarcloud.io/api/ce/task?id=AZ2Qge89T7Y829rQbv87 %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
		{
			wantError: true,
			name:      "27 providing an incorrect pull-request ID with project key gives an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-project-key cyber-dojo_differ --pull-request 1 %s", suite.defaultKosliArguments),
			golden:    "Error: pull request 1 not found for project cyber-dojo_differ on https://sonarcloud.io\n",
		},
		{
			wantError: true,
			name:      "28 if expired CE task URL is provided, we get an error",
			cmd:       fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-ce-task-url https://sonarcloud.io/api/ce/task?id=AZERk4uWpzGpahwkB9ac %s", suite.defaultKosliArguments),
			golden:    "Error: No activity found for task 'AZERk4uWpzGpahwkB9ac' on https://sonarcloud.io. \nSonarQube may be experiencing problems, please check https://status.sonarqube.com/ and try again later. \nOtherwise if you are attesting an older scan, the snapshot may have been deleted by SonarQube\n",
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
			golden:    "Error: open .scannerwork/report-task.txt: no such file or directory. Check your working directory is set correctly. Alternatively provide the project key and either revision or pull-request ID for the scan to attest\n",
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
			golden:    "Error: SonarQube error: Component key 'test99' not found\n",
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
		{
			name:   "can retrieve scan results using report-task.txt file and pull-request flag",
			cmd:    fmt.Sprintf("attest sonar --name cli.foo --commit HEAD --origin-url http://www.example.com --sonar-working-dir testdata/sonar/sonarqube/.scannerwork --pull-request 359 --sonar-server-url http://example.com %s", suite.defaultKosliArguments),
			golden: "sonar attestation 'foo' is reported to trail: test-123\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// createPRScannerWorkDir downloads the report-task.txt from the latest
// SonarCloud PR scan of cyber-dojo_differ. The file is uploaded as a GitHub
// Actions artifact by the sonar-pr-trigger workflow in that repo.
func createPRScannerWorkDir(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()
	httpClient := &http.Client{}

	// GitHub token is required to download workflow artifacts.
	// Check GH_TOKEN (gh CLI), GITHUB_TOKEN (CI), in that order.
	ghToken := os.Getenv("GH_TOKEN")
	if ghToken == "" {
		ghToken = os.Getenv("GITHUB_TOKEN")
	}
	if ghToken == "" {
		t.Fatalf("GH_TOKEN or GITHUB_TOKEN must be set to download artifacts from GitHub")
	}

	// Find the artifact ID via GitHub API
	req, err := http.NewRequest("GET",
		"https://api.github.com/repos/cyber-dojo/differ/actions/artifacts?name=sonar-pr-report-task&per_page=1", nil)
	if err != nil {
		t.Fatalf("failed to create artifact list request: %v", err)
	}
	req.Header.Add("Authorization", "Bearer "+ghToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		t.Fatalf("failed to list artifacts from GitHub: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GitHub artifacts API returned status %d", resp.StatusCode)
	}

	var result struct {
		Artifacts []struct {
			ID      int    `json:"id"`
			Expired bool   `json:"expired"`
			Name    string `json:"name"`
		} `json:"artifacts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode artifacts response: %v", err)
	}
	if len(result.Artifacts) == 0 {
		t.Fatalf("no sonar-pr-report-task artifact found in cyber-dojo/differ")
	}
	if result.Artifacts[0].Expired {
		t.Fatalf("sonar-pr-report-task artifact has expired")
	}

	// Download the artifact zip
	downloadURL := fmt.Sprintf("https://api.github.com/repos/cyber-dojo/differ/actions/artifacts/%d/zip", result.Artifacts[0].ID)
	dlReq, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		t.Fatalf("failed to create artifact download request: %v", err)
	}
	dlReq.Header.Add("Authorization", "Bearer "+ghToken)

	dlResp, err := httpClient.Do(dlReq)
	if err != nil {
		t.Fatalf("failed to download artifact: %v", err)
	}
	defer func() { _ = dlResp.Body.Close() }()

	if dlResp.StatusCode != http.StatusOK {
		t.Fatalf("artifact download returned status %d", dlResp.StatusCode)
	}

	// Save zip to temp file, then extract report-task.txt
	zipPath := filepath.Join(tmpDir, "artifact.zip")
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("failed to create temp zip file: %v", err)
	}
	if _, err := io.Copy(zipFile, dlResp.Body); err != nil {
		_ = zipFile.Close()
		t.Fatalf("failed to write artifact zip: %v", err)
	}
	_ = zipFile.Close()

	zipReader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("failed to open artifact zip: %v", err)
	}
	defer func() { _ = zipReader.Close() }()

	for _, f := range zipReader.File {
		if f.Name == "report-task.txt" {
			rc, err := f.Open()
			if err != nil {
				t.Fatalf("failed to open report-task.txt in zip: %v", err)
			}
			outPath := filepath.Join(tmpDir, "report-task.txt")
			outFile, err := os.Create(outPath)
			if err != nil {
				_ = rc.Close()
				t.Fatalf("failed to create report-task.txt: %v", err)
			}
			_, err = io.Copy(outFile, rc)
			_ = rc.Close()
			_ = outFile.Close()
			if err != nil {
				t.Fatalf("failed to extract report-task.txt: %v", err)
			}
			return tmpDir
		}
	}

	t.Fatalf("report-task.txt not found in artifact zip")
	return ""
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestSonarCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestSonarCommandTestSuite))
	suite.Run(t, new(AttestSonarQubeCommandTestSuite))
}
