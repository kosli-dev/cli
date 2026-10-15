package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kosli-dev/cli/internal/kosli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type GetArtifactCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	artifactName          string
	artifactPath          string
	fingerprint           string
	trailName             string
}

func (suite *GetArtifactCommandTestSuite) SetupTest() {
	suite.flowName = "get-artifact"
	suite.artifactName = "arti"
	suite.artifactPath = "testdata/folder1/hello.txt"
	suite.trailName = "cli-build-1"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	fingerprintOptions := &fingerprintOptions{
		artifactType: "file",
	}
	var err error
	suite.fingerprint, err = GetSha256Digest(suite.artifactPath, fingerprintOptions, logger)
	require.NoError(suite.T(), err)
	CreateArtifactOnTrail(suite.flowName, suite.trailName, "cli", suite.fingerprint, suite.artifactName, suite.T())
}

func (suite *GetArtifactCommandTestSuite) TestGetArtifactCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "getting a non existing artifact fails",
			cmd:       fmt.Sprintf(`get artifact %s@8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: Artifact with fingerprint '8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c' does not exist in flow 'get-artifact' belonging to organization 'docs-cmd-test-user-shared'\n",
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`get artifact %s@8e568bd886069f1290def0caabc1e97ce0e7b80c105e611258b57d76fcef234c xxx %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`get artifact %s@%s --org orgX`, suite.flowName, suite.fingerprint),
			golden:    "Error: --api-token is not set\nUsage: kosli get artifact EXPRESSION [flags]\n",
		},
		{
			wantError: true,
			name:      "getting an existing artifact using an invalid expression fails",
			cmd:       fmt.Sprintf(`get artifact %s#%s %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: invalid expression: %s#%s\n", suite.flowName, suite.fingerprint),
		},
		{
			name:       "getting an existing artifact using fingerprint works",
			cmd:        fmt.Sprintf(`get artifact %s@%s %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
			goldenFile: "output/get/get-artifact.txt",
		},
		{
			name: "getting an existing artifact using fingerprint with --output json works",
			cmd:  fmt.Sprintf(`get artifact %s@%s --output json %s`, suite.flowName, suite.fingerprint, suite.defaultKosliArguments),
		},
		{
			name:       "get an existing artifact using commit works",
			cmd:        fmt.Sprintf(`get artifact %s:0fc1ba9876f91b215679f3649b8668085d820ab5 %s`, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-artifact.txt",
		},
		{
			name:       "get an existing artifact using commit with a provided trail works",
			cmd:        fmt.Sprintf(`get artifact %s:0fc1ba9876f91b215679f3649b8668085d820ab5 --trail=%s %s`, suite.flowName, suite.trailName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-artifact.txt",
		},
		{
			name:       "get an existing artifact using fingerprint with a provided trail works",
			cmd:        fmt.Sprintf(`get artifact %s@%s --trail=%s %s`, suite.flowName, suite.fingerprint, suite.trailName, suite.defaultKosliArguments),
			goldenFile: "output/get/get-artifact.txt",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestGetArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetArtifactCommandTestSuite))
}

func TestPrintArtifactsAsTableToleratesMissingCollections(t *testing.T) {
	// an artifact without created_at, running, exited or history must render
	// N/A and omit those sections rather than panic
	artifacts := []kosli.Artifact{{Filename: "arti", FlowName: "flow-1", Fingerprint: "abc", GitCommit: "0fc1ba9",
		CommitURL: "https://c", BuildURL: "https://b", HTMLURL: "https://h", State: "COMPLIANT"}}
	var buf bytes.Buffer
	require.NotPanics(t, func() {
		require.NoError(t, printArtifactsAsTable(artifacts, &buf, 0))
	})
	out := buf.String()
	require.Regexp(t, `Created on:\s+N/A\n`, out)
	require.NotContains(t, out, "History:")
	require.NotContains(t, out, "Running in environments:")
	require.NotContains(t, out, "Trail:")
}

func TestPrintArtifactsAsTablePrintsFractionalTimestamps(t *testing.T) {
	artifacts := []kosli.Artifact{{Filename: "arti", CreatedAt: "1452902400.25",
		History: []kosli.ArtifactEvent{{Event: "cli reported", Timestamp: "1452902400"}}}}
	var buf bytes.Buffer
	require.NoError(t, printArtifactsAsTable(artifacts, &buf, 0))
	out := buf.String()
	require.NotRegexp(t, `Created on:\s+N/A`, out)
	require.Contains(t, out, "History:\n")
	require.Contains(t, out, "    cli reported")
}

func TestPrintArtifactsAsTablePrintsPresentButEmptyTrailRows(t *testing.T) {
	// a key the server sends prints its row even when empty; only an absent
	// or null key leaves the row out
	empty := ""
	artifacts := []kosli.Artifact{{Filename: "arti", TrailName: &empty, TemplateReferenceName: &empty}}
	var buf bytes.Buffer
	require.NoError(t, printArtifactsAsTable(artifacts, &buf, 0))
	out := buf.String()
	require.Regexp(t, `(?m)^Trail:\s*$`, out)
	require.Regexp(t, `(?m)^Name in template:\s*$`, out)
}

func TestGetArtifactPrintsJSONThatTheTableCannotDecode(t *testing.T) {
	// -o json prints what the server sent, so a field the typed table cannot
	// decode must not fail it; only the table reports the decode error
	body := `{"filename":"arti","created_at":""}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(body))
		assert.NoError(t, err)
	}))
	defer server.Close()
	flags := fmt.Sprintf("--host %s --org test-org --api-token test-token", server.URL)

	_, _, stdout, _, err := executeCommandC("get artifact flow-1@abc --output json " + flags)
	require.NoError(t, err)
	require.JSONEq(t, body, stdout)

	_, _, _, _, err = executeCommandC("get artifact flow-1@abc " + flags)
	require.ErrorContains(t, err, "failed to decode the Kosli API response")
}
