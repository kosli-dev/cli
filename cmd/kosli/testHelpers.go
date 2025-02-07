package main

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/mattn/go-shellwords"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// cmdTestCase describes a cmd test case.
type cmdTestCase struct {
	name             string
	cmd              string
	golden           string
	goldenFile       string
	goldenRegex      string
	wantError        bool
	additionalConfig interface{}
}

// executeCommandStdinC executes a command as a user would and return the output
// this creates a new kosli command that is run, but it cannot be used in other tests
// because newRootCmd overwrites the global options
func executeCommandC(cmd string) (*cobra.Command, string, error) {
	args, err := shellwords.Parse(cmd)
	if err != nil {
		return nil, "", err
	}

	buf := new(bytes.Buffer)

	root, err := newRootCmd(buf, buf, args)
	if err != nil {
		return nil, "", err
	}

	root.SilenceErrors = false
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)

	c, err := root.ExecuteC()
	output := buf.String()

	return c, output, err
}

// runTestCmd runs a table of cmd test cases
func runTestCmd(t *testing.T, tests []cmdTestCase) {
	t.Helper()
	for _, key := range [...]string{"KOSLI_API_TOKEN", "KOSLI_ORG"} {
		if os.Getenv(key) != "" {
			t.Errorf("Environment variable %s should not be set when running tests ", key)
		}
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.golden != "" && tt.goldenFile != "" {
				t.Error("golden and goldenPath cannot be set together")
			}
			t.Logf("running cmd: %s", tt.cmd)
			_, out, err := executeCommandC(tt.cmd)
			if (err != nil) != tt.wantError {
				t.Errorf("error expectation not matched\n\n WANT error is: %t\n\n but GOT: '%v'", tt.wantError, err)
			}
			if tt.golden != "" {
				if !bytes.Equal([]byte(tt.golden), []byte(out)) {
					t.Errorf("does not match golden\n\nWANT:\n'%s'\n\nGOT:\n'%s'\n", tt.golden, out)
				}
			} else if tt.goldenFile != "" {
				if err := compareAgainstFile([]byte(out), goldenPath(tt.goldenFile)); err != nil {
					t.Error(err)
				}
			} else if tt.goldenRegex != "" {
				require.Regexp(t, tt.goldenRegex, out)
			}
		})
	}
}

func goldenPath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}
	return filepath.Join("testdata", filename)
}

func compareTwoFiles(actualFilename, expectedFilename string) error {
	actual, err := os.ReadFile(actualFilename)
	if err != nil {
		return errors.Wrapf(err, "unable to read golden file %s", actualFilename)
	}

	expected, err := os.ReadFile(expectedFilename)
	if err != nil {
		return errors.Wrapf(err, "unable to read golden file %s", expectedFilename)
	}
	return compareFileBytes(actual, expected)
}

// compareAgainstFile compares the content of an actual file against a file containing regex patterns.
func compareAgainstFile(actual []byte, filename string) error {
	// Read the expected file with regex patterns
	expectedFile, err := os.Open(filename)
	if err != nil {
		return errors.Wrapf(err, "unable to read golden file %s", filename)
	}
	defer expectedFile.Close()

	// Scanner to read the expected file line by line
	expectedScanner := bufio.NewScanner(expectedFile)
	actualScanner := bufio.NewScanner(bytes.NewReader(actual))

	lineNum := 1
	for expectedScanner.Scan() {
		if !actualScanner.Scan() {
			return errors.Errorf("line %d: expected more lines in actual content", lineNum)
		}

		expectedLine := expectedScanner.Text()
		actualLine := actualScanner.Text()
		expectedLine = strings.TrimSpace(expectedLine)
		actualLine = strings.TrimSpace(actualLine)

		// Compile the regex pattern from the expected file
		re, err := regexp.Compile(expectedLine + "$")
		if err != nil {
			return errors.Wrapf(err, "invalid regex on line %d in golden file %s", lineNum, filename)
		}

		// Check if the actual line matches the regex pattern
		if !re.MatchString(actualLine) {
			return errors.Errorf("line %d does not match: expected pattern '%s', got '%s'", lineNum, expectedLine, actualLine)
		}

		lineNum++
	}

	// Check if there are extra lines in the actual content
	if actualScanner.Scan() {
		return errors.Errorf("unexpected additional content after line %d in actual content", lineNum)
	}

	if err := expectedScanner.Err(); err != nil {
		return errors.Wrapf(err, "error reading golden file %s", filename)
	}

	if err := actualScanner.Err(); err != nil {
		return errors.Wrap(err, "error reading actual content")
	}

	return nil
}

// func compareAgainstFile(actual []byte, filename string) error {
// 	expected, err := os.ReadFile(filename)
// 	if err != nil {
// 		return errors.Wrapf(err, "unable to read golden file %s", filename)
// 	}
// 	if err := compareFileBytes(actual, expected); err != nil {
// 		return errors.Errorf("does not match golden file %s\n\nWANT:\n'%s'\n\nGOT:\n'%s'", filename, expected, actual)
// 	}

// 	return nil
// }

func compareFileBytes(actual, expected []byte) error {
	actual = normalize(actual)
	expected = normalize(expected)

	if !bytes.Equal(actual, expected) {
		return errors.Errorf("actual does not match expected")
	}
	return nil
}

func normalize(in []byte) []byte {
	normalized := bytes.Replace(in, []byte("\r\n"), []byte("\n"), -1)
	return []byte(strings.TrimSpace(string(normalized)))
}

func ArchiveCustomAttestationType(typeName string, t *testing.T) {
	t.Helper()

	err := newArchiveAttestationTypeCmd(os.Stdout).RunE(nil, []string{typeName})
	require.NoError(t, err, "attestation type should be created without error")
}

func CreateCustomAttestationType(typeName, schemaFilePath string, jqEvaluators []string, t *testing.T) {
	t.Helper()
	o := &createAttestationTypeOptions{
		payload: CreateAttestationTypePayload{
			TypeName: typeName,
		},
		schemaFilePath: schemaFilePath,
		jqRules:        jqEvaluators,
	}
	err := o.run([]string{typeName})
	require.NoError(t, err, "attestation type should be created without error")
}

// CreateFlow creates a flow on the server
func CreateFlow(flowName string, t *testing.T) {
	t.Helper()
	o := &createFlowOptions{
		payload: FlowPayload{
			Name:        flowName,
			Description: "test flow",
			Visibility:  "private",
		},
	}

	err := o.run([]string{flowName})
	require.NoError(t, err, "flow should be created without error")
}

// CreateFlowWithTemplate creates a flow with a yaml template on the server
func CreateFlowWithTemplate(flowName, templatePath string, t *testing.T) {
	t.Helper()
	o := &createFlowOptions{
		payload: FlowPayload{
			Name:        flowName,
			Description: "test flow",
			Visibility:  "private",
		},
		TemplateFile: templatePath,
	}

	err := o.run([]string{flowName})
	require.NoError(t, err, "flow should be created without error")
}

// BeginTrail creates a trail with a yaml template on the server
func BeginTrail(trailName, flowName, templatePath string, t *testing.T) {
	t.Helper()
	o := &beginTrailOptions{
		payload: TrailPayload{
			Name:        trailName,
			Description: "test trail",
		},
		templateFile: templatePath,
		flow:         flowName,
	}

	err := o.run([]string{trailName})
	require.NoError(t, err, "trail should be begun without error")
}

// CreateArtifact creates an artifact on the server
func CreateArtifact(flowName, artifactFingerprint, artifactName string, t *testing.T) {
	t.Helper()
	o := &reportArtifactOptions{
		srcRepoRoot: "../..",
		flowName:    flowName,
		// name:         "",
		gitReference: "0fc1ba9876f91b215679f3649b8668085d820ab5",
		payload: ArtifactPayload{
			Fingerprint: artifactFingerprint,
			GitCommit:   "0fc1ba9876f91b215679f3649b8668085d820ab5",
			BuildUrl:    "www.yr.no",
			CommitUrl:   "https://www.nrk.no",
		},
	}

	o.fingerprintOptions = new(fingerprintOptions)

	err := o.run([]string{artifactName})
	require.NoError(t, err, "artifact should be created without error")
}

// CreateArtifactOnTrail creates an artifact on a trail on the server
func CreateArtifactOnTrail(flowName, trailName, stepName, artifactFingerprint, artifactName string, t *testing.T) {
	t.Helper()
	o := &attestArtifactOptions{
		srcRepoRoot:  "../..",
		flowName:     flowName,
		gitReference: "0fc1ba9876f91b215679f3649b8668085d820ab5",
		payload: AttestArtifactPayload{
			Fingerprint: artifactFingerprint,
			GitCommit:   "0fc1ba9876f91b215679f3649b8668085d820ab5",
			BuildUrl:    "www.yr.no",
			CommitUrl:   "https://www.nrk.no",
			TrailName:   trailName,
			Name:        stepName,
		},
	}

	o.fingerprintOptions = new(fingerprintOptions)

	err := o.run([]string{artifactName})
	require.NoError(t, err, "artifact should be created without error")
}

func CreateArtifactWithCommit(flowName, artifactFingerprint, artifactName string, gitCommit string, t *testing.T) {
	t.Helper()
	o := &reportArtifactOptions{
		srcRepoRoot: "../..",
		flowName:    flowName,
		// name:         "",
		gitReference: gitCommit,
		payload: ArtifactPayload{
			Fingerprint: artifactFingerprint,
			GitCommit:   gitCommit,
			BuildUrl:    "www.yr.no",
			CommitUrl:   "https://www.nrk.no",
		},
	}

	o.fingerprintOptions = new(fingerprintOptions)

	err := o.run([]string{artifactName})
	require.NoError(t, err, "artifact should be created without error")
}

// CreateApproval creates an approval for an artifact in a flow
func CreateApproval(flowName, fingerprint string, t *testing.T) {
	t.Helper()
	o := &reportApprovalOptions{
		payload: ApprovalPayload{
			ArtifactFingerprint: fingerprint,
			Description:         "some description",
		},
		flowName:        flowName,
		oldestSrcCommit: "75690c740e7b222a3948f4f7618262a5254044e2",
		newestSrcCommit: "cfbdba789edd14e5970405896c637dbf073ef831",
		srcRepoRoot:     "../..",
	}

	err := o.run([]string{"filename"}, false)
	require.NoError(t, err, "approval should be created without error")
}

// EnableBeta enables beta features for the org
func EnableBeta(t *testing.T) {
	t.Helper()
	o := &betaOptions{}
	o.payload.Enabled = true
	err := o.run([]string{})
	require.NoError(t, err, "beta should be enabled without error")
}

// ExpectDeployment reports a deployment expectation of a given artifact to the server
func ExpectDeployment(flowName, fingerprint, envName string, t *testing.T) {
	t.Helper()
	o := &expectDeploymentOptions{
		flowName: flowName,
		payload: ExpectDeploymentPayload{
			Fingerprint: fingerprint,
			Environment: envName,
			BuildUrl:    "https://example.com",
		},
	}
	err := o.run([]string{})
	require.NoError(t, err, "deployment should be expected without error")
}

// CreateEnv creates an env on the server
func CreateEnv(org, envName, envType string, t *testing.T) {
	t.Helper()
	o := &createEnvOptions{
		payload: CreateEnvironmentPayload{
			Name:        envName,
			Type:        envType,
			Description: "test env",
		},
	}

	err := o.run([]string{envName})
	require.NoError(t, err, "env should be created without error")
}

// ReportServerArtifactToEnv reports files/dirs in paths as server env artifacts
func ReportServerArtifactToEnv(paths []string, envName string, t *testing.T) {
	t.Helper()
	o := &snapshotServerOptions{
		paths: paths,
	}
	err := o.run([]string{envName})
	require.NoError(t, err, "server env should be reported without error")
}

func SetEnvVars(envVars map[string]string, t *testing.T) {
	for key, value := range envVars {
		err := os.Setenv(key, value)
		require.NoErrorf(t, err, "error setting env variable %s", key)
	}
}

func UnSetEnvVars(envVars map[string]string, t *testing.T) {
	for key := range envVars {
		err := os.Unsetenv(key)
		require.NoErrorf(t, err, "error unsetting env variable %s", key)
	}
}

// CreatePolicy creates a policy on the server
func CreatePolicy(org, policyName string, t *testing.T) {
	t.Helper()
	o := &createPolicyOptions{
		payload: PolicyPayload{
			Name:        policyName,
			Type:        "env",
			Description: "test policy",
		},
	}

	err := o.run([]string{policyName, "testdata/policy-files/test-policy.yml"})
	require.NoError(t, err, "policy should be created without error")
}
