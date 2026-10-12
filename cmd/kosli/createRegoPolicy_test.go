package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	regoPolicyPending = `{"identifier":"pr-approved","version":1,"status":"pending",` +
		`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe"}`
	regoPolicyValid = `{"identifier":"pr-approved","version":1,"status":"valid",` +
		`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000002.25,` +
		`"opa_version":"1.21.1","entrypoint":"policy"}`
)

// fakeRegoPolicies records each publish and answers each version read with the
// next body in reads, repeating the last one once they run out.
type fakeRegoPolicies struct {
	published  []map[string]any
	publishURL string
	reads      []string
	readCount  int
	readURL    string
	unexpected int
}

func newFakeRegoPolicies(t *testing.T, reads ...string) (*httptest.Server, *fakeRegoPolicies) {
	t.Helper()
	fake := &fakeRegoPolicies{reads: reads}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/rego-policies/"):
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var body map[string]any
			require.NoError(t, json.Unmarshal(raw, &body))
			fake.published = append(fake.published, body)
			fake.publishURL = r.URL.Path
			w.WriteHeader(http.StatusAccepted)
			_, _ = fmt.Fprint(w, regoPolicyPending)

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/versions/"):
			fake.readURL = r.URL.Path
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, fake.reads[min(fake.readCount, len(fake.reads)-1)])
			fake.readCount++

		default:
			fake.unexpected++
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server, fake
}

type CreateRegoPolicyCommandTestSuite struct {
	suite.Suite
	originalWait regopolicies.WaitOptions
}

func (suite *CreateRegoPolicyCommandTestSuite) SetupTest() {
	suite.originalWait = regoPolicyWaitOptions
	regoPolicyWaitOptions = regopolicies.WaitOptions{
		Timeout: 2 * time.Second,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
}

func (suite *CreateRegoPolicyCommandTestSuite) TearDownTest() {
	regoPolicyWaitOptions = suite.originalWait
}

func (suite *CreateRegoPolicyCommandTestSuite) createCmd(host, args string) string {
	return fmt.Sprintf("create rego-policy %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *CreateRegoPolicyCommandTestSuite) TestItIsHiddenFromHelp() {
	_, combined, _, _, err := executeCommandC("create --help")

	require.NoError(suite.T(), err)
	require.NotContains(suite.T(), combined, "rego-policy")
	require.Contains(suite.T(), combined, "flow", "the rest of the help still renders")
}

func (suite *CreateRegoPolicyCommandTestSuite) TestItPublishesTheFileAndReportsTheValidVersion() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.createCmd(server.URL,
		`pr-approved testdata/policies/allow-all.rego --input-schema trail `+
			`--name "Pull request approved" --description "Approved PRs only" --comment "Initial version"`))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, fake.unexpected)
	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved", fake.publishURL)
	require.Len(suite.T(), fake.published, 1)

	published := fake.published[0]
	require.Equal(suite.T(), "trail", published["input_schema"])
	require.Equal(suite.T(), "Pull request approved", published["name"])
	require.Equal(suite.T(), "Approved PRs only", published["description"])
	require.Equal(suite.T(), "Initial version", published["comment"])
	files := published["files"].(map[string]any)
	require.Len(suite.T(), files, 1)
	require.Contains(suite.T(), files["allow-all.rego"], "package policy")

	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved/versions/1", fake.readURL)
	require.Contains(suite.T(), combined, "rego policy 'pr-approved' version 1 is valid (sha256:abc123)")
}

func (suite *CreateRegoPolicyCommandTestSuite) TestItLeavesOutTheNameAndDescriptionWhenNotGiven() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, _, _, err := executeCommandC(suite.createCmd(server.URL,
		"pr-approved testdata/policies/allow-all.rego --input-schema trail"))

	require.NoError(suite.T(), err)
	require.NotContains(suite.T(), fake.published[0], "name")
	require.NotContains(suite.T(), fake.published[0], "description")
}

func (suite *CreateRegoPolicyCommandTestSuite) TestItWaitsThroughPendingReads() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyPending, regoPolicyPending, regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.createCmd(server.URL,
		"pr-approved testdata/policies/allow-all.rego --input-schema trail"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 3, fake.readCount)
	require.Contains(suite.T(), combined, "is valid")
}

// A directory travels as `evaluate policy` sends it, so the same directory
// gives the same digest on both paths.
func (suite *CreateRegoPolicyCommandTestSuite) TestADirectoryTravelsKeyedByRelativePath() {
	dir := suite.T().TempDir()
	require.NoError(suite.T(), os.MkdirAll(filepath.Join(dir, "lib"), 0o755))
	require.NoError(suite.T(), os.WriteFile(filepath.Join(dir, "policy.rego"), []byte("package policy\n"), 0o644))
	require.NoError(suite.T(), os.WriteFile(filepath.Join(dir, "lib", "helpers.rego"), []byte("package lib\n"), 0o644))
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, _, _, err := executeCommandC(suite.createCmd(server.URL,
		"pr-approved "+dir+" --input-schema trails"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), map[string]any{
		"policy.rego":      "package policy\n",
		"lib/helpers.rego": "package lib\n",
	}, fake.published[0]["files"])
	require.Equal(suite.T(), "trails", fake.published[0]["input_schema"])
}

// Until the outcomes are told apart, nothing but a valid version may read as
// one.
func (suite *CreateRegoPolicyCommandTestSuite) TestAVersionThatIsNotValidFails() {
	rejected := strings.Replace(regoPolicyValid, `"status":"valid"`, `"status":"rejected"`, 1)
	server, _ := newFakeRegoPolicies(suite.T(), rejected)

	_, combined, _, _, err := executeCommandC(suite.createCmd(server.URL,
		"pr-approved testdata/policies/allow-all.rego --input-schema trail"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "rejected")
	require.NotContains(suite.T(), combined, "is valid")
}

func (suite *CreateRegoPolicyCommandTestSuite) TestItNeedsAnIdentifierAndAPath() {
	_, _, _, _, err := executeCommandC(suite.createCmd("http://localhost:1",
		"pr-approved --input-schema trail"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "accepts 2 arg(s), received 1")
}

func (suite *CreateRegoPolicyCommandTestSuite) TestItNeedsAnInputSchema() {
	_, _, _, _, err := executeCommandC(suite.createCmd("http://localhost:1",
		"pr-approved testdata/policies/allow-all.rego"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), `required flag(s) "input-schema" not set`)
}

func TestCreateRegoPolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CreateRegoPolicyCommandTestSuite))
}
