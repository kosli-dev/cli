package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	evaluationPath      = "/api/v2/evaluations/test-org/01EVAL"
	evaluationInputPath = evaluationPath + "/input"
	publishedBundlePath = "/api/v2/rego-policies/test-org/pr-approved/versions/3/bundle"

	readCompletedPublished = `{"id":"01EVAL","status":"completed","requested_at":1757000000.5,` +
		`"recorded_at":1757000000.5,"context":{"trails":[{"flow":"my-flow","trail":"my-trail"}]},` +
		`"policy":{"identifier":"pr-approved","version":3,"digest":"sha256:abc123"},` +
		`"params":{"min_approvals":2},"result":{"allow":true,"outputs":{}},"opa_version":"1.4.0",` +
		`"decision":{"control":"ctrl-1","name":"ctrl-1-decision","flow":"my-flow","trail":"my-trail"},` +
		`"decision_attestation_id":"01DECISION"}`
	readDeniedInline = `{"id":"01EVAL","status":"completed","requested_at":1757000000.5,` +
		`"recorded_at":1757000000.5,"context":{"trails":[{"flow":"my-flow","trail":"my-trail"}]},` +
		`"policy":{"digest":"sha256:def456","files":{"policy.rego":"package policy\n",` +
		`"lib/util.rego":"package util\n"}},"params":{},` +
		`"result":{"allow":false,"outputs":{"violations":["change is not approved"]}}}`
	readFailed = `{"id":"01EVAL","status":"failed","requested_at":1757000000.5,` +
		`"recorded_at":1757000000.5,"context":{"trails":[{"flow":"my-flow","trail":"my-trail"}]},` +
		`"policy":{"digest":"sha256:def456","files":{"policy.rego":"package policy\n"}},"params":{},` +
		`"error":{"kind":"policy_error","message":"allow is undefined"}}`
	readPending = `{"id":"01EVAL","status":"pending","requested_at":1757000000.5,` +
		`"recorded_at":1757000000.5,"context":{"trails":[{"flow":"my-flow","trail":"my-trail"}]},` +
		`"policy":{"digest":"sha256:def456","files":{"policy.rego":"package policy\n"}},"params":{}}`

	evaluationInput = `{"evaluation":{"time":"2025-09-04T15:33:20.500000Z"},` +
		`"trails":[{"flow_name":"my-flow","trail_name":"my-trail"}],` +
		`"trail":{"flow_name":"my-flow","trail_name":"my-trail"}}`
)

type GetEvaluationCommandTestSuite struct {
	suite.Suite
}

func (suite *GetEvaluationCommandTestSuite) SetupTest() {
	// Makes formattedTimestamp print a fixed UTC time.
	suite.T().Setenv("KOSLI_TESTS", "true")
}

func (suite *GetEvaluationCommandTestSuite) getCmd(host, args string) string {
	return fmt.Sprintf("get evaluation %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *GetEvaluationCommandTestSuite) TestItIsBeta() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	cmd, _, err := root.Find([]string{"get", "evaluation"})
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), "evaluation", cmd.Name())
	require.Contains(suite.T(), cmd.Annotations, betaCLIAnnotation)
}

func (suite *GetEvaluationCommandTestSuite) TestItNeedsTheOrgAndApiToken() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)
	noConfig := withoutOrgAndApiToken(suite.T())

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"get evaluation 01EVAL --api-token test-token --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --org is not set\n")
	require.Empty(suite.T(), fake.reads)
}

func (suite *GetEvaluationCommandTestSuite) TestItTakesExactlyOneId() {
	for _, args := range []string{"", "01EVAL 01OTHER"} {
		server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)

		_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, args))

		require.Error(suite.T(), err, args)
		require.Contains(suite.T(), combined, "accepts 1 arg(s)", args)
		require.Empty(suite.T(), fake.reads, args)
	}
}

func (suite *GetEvaluationCommandTestSuite) TestItPrintsACompletedEvaluationOfAPublishedPolicy() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), []string{evaluationPath}, fake.reads)
	for _, row := range [][2]string{
		{"ID", "01EVAL"},
		{"Status", "completed"},
		{"Policy", "pr-approved, version 3 (sha256:abc123)"},
		{"Trails", "my-flow/my-trail"},
		{"Requested at", "Sat, 16 Jan 2016 00:00:00 UTC"},
		{"Recorded at", "Sat, 16 Jan 2016 00:00:00 UTC"},
		{"Outcome", "ALLOWED"},
		{"OPA version", "1.4.0"},
		{"Decision", "ctrl-1-decision for control ctrl-1 in my-flow/my-trail"},
		{"Decision attestation", "01DECISION"},
	} {
		requireRow(suite.T(), combined, row[0], row[1])
	}
}

func (suite *GetEvaluationCommandTestSuite) TestItPrintsADenialWithItsViolationsAndAnInlinePolicyByDigest() {
	server, _ := newFakeEvaluationRead(suite.T(), readDeniedInline)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL"))

	require.NoError(suite.T(), err, "reading a denial is not a denial")
	requireRow(suite.T(), combined, "Policy", "sha256:def456")
	requireRow(suite.T(), combined, "Outcome", "DENIED")
	requireRow(suite.T(), combined, "Violations", "change is not approved")
	require.NotContains(suite.T(), combined, "Decision")
}

func (suite *GetEvaluationCommandTestSuite) TestItPrintsWhyAnEvaluationFailedAndExitsZero() {
	server, _ := newFakeEvaluationRead(suite.T(), readFailed)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL"))

	require.NoError(suite.T(), err)
	requireRow(suite.T(), combined, "Status", "failed")
	requireRow(suite.T(), combined, "Error", "policy_error: allow is undefined")
	require.NotContains(suite.T(), combined, "Outcome")
}

func (suite *GetEvaluationCommandTestSuite) TestItPrintsAPendingEvaluationWithNoOutcome() {
	server, _ := newFakeEvaluationRead(suite.T(), readPending)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL"))

	require.NoError(suite.T(), err)
	requireRow(suite.T(), combined, "Status", "pending")
	require.NotContains(suite.T(), combined, "Outcome")
}

func (suite *GetEvaluationCommandTestSuite) TestItPrintsTheEvaluationAsJSONAsTheServerSentIt() {
	server, _ := newFakeEvaluationRead(suite.T(), readCompletedPublished)

	_, _, stdout, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), readCompletedPublished, stdout)
}

func (suite *GetEvaluationCommandTestSuite) TestItReturnsANotFoundInTheServersWords() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)
	fake.evaluation.status = http.StatusNotFound
	fake.evaluation.body = `{"message":"Evaluation '01EVAL' does not exist in org 'test-org'"}`

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Evaluation '01EVAL' does not exist in org 'test-org'")
}

func (suite *GetEvaluationCommandTestSuite) TestItKeepsTheIdInOneSegment() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)

	_, _, _, _, _ = executeCommandC(suite.getCmd(server.URL, "a/b"))

	require.Equal(suite.T(), []string{"/api/v2/evaluations/test-org/a%2Fb"}, fake.reads)
}

func (suite *GetEvaluationCommandTestSuite) TestItRefusesAnIdThatIsNotASegment() {
	for _, id := range []string{".", ".."} {
		server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)

		_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "'"+id+"'"))

		require.Error(suite.T(), err, id)
		require.Contains(suite.T(), combined, fmt.Sprintf("%q is not a valid evaluation id", id), id)
		require.Empty(suite.T(), fake.reads, id)
	}
}

func (suite *GetEvaluationCommandTestSuite) TestShowInputAddsTheInputUnderItsOwnKey() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)

	_, _, stdout, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --show-input --output json"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), []string{evaluationPath, evaluationInputPath}, fake.reads)
	var printed map[string]any
	require.NoError(suite.T(), json.Unmarshal([]byte(stdout), &printed))
	require.JSONEq(suite.T(), evaluationInput, mustJSON(suite.T(), printed["input"]))
	delete(printed, "input")
	require.JSONEq(suite.T(), readCompletedPublished, mustJSON(suite.T(), printed))
}

func (suite *GetEvaluationCommandTestSuite) TestShowInputNeedsJSON() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --show-input"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "--show-input needs --output json")
	require.Empty(suite.T(), fake.reads)
}

func (suite *GetEvaluationCommandTestSuite) TestAnInputThatCannotBeBuiltIsReportedInTheServersWords() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)
	fake.input.status = http.StatusConflict
	fake.input.body = `{"message":"No policy input can be built: trail has more than one attestation named 'tests'."}`

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --show-input -o json"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "No policy input can be built")
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureOfAnInlinePolicyWritesInputParamsAndFiles() {
	server, fake := newFakeEvaluationRead(suite.T(), readDeniedInline)
	dir := filepath.Join(suite.T().TempDir(), "replay")

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture --dir "+dir))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), []string{evaluationPath, evaluationInputPath}, fake.reads)
	require.JSONEq(suite.T(), evaluationInput, readFile(suite.T(), dir, "input.json"))
	require.JSONEq(suite.T(), `{"params":{}}`, readFile(suite.T(), dir, "data.json"))
	require.Equal(suite.T(), "package policy\n", readFile(suite.T(), dir, "policy/policy.rego"))
	require.Equal(suite.T(), "package util\n", readFile(suite.T(), dir, "policy/lib/util.rego"))
	require.Contains(suite.T(), combined, "Wrote the input, params and policy of evaluation 01EVAL to "+dir)
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureOfAPublishedPolicyUnpacksItsBundle() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)
	fake.bundle.body = gzippedTar(suite.T(), map[string]string{
		"policy.rego":   "package policy\n",
		"data.yaml":     "threshold: 1\n",
		"lib/util.rego": "package util\n",
	})
	dir := filepath.Join(suite.T().TempDir(), "replay")

	_, _, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture --dir "+dir))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), []string{evaluationPath, evaluationInputPath, publishedBundlePath}, fake.reads)
	require.JSONEq(suite.T(), `{"params":{"min_approvals":2}}`, readFile(suite.T(), dir, "data.json"))
	require.Equal(suite.T(), "package policy\n", readFile(suite.T(), dir, "policy/policy.rego"))
	require.Equal(suite.T(), "threshold: 1\n", readFile(suite.T(), dir, "policy/data.yaml"))
	require.Equal(suite.T(), "package util\n", readFile(suite.T(), dir, "policy/lib/util.rego"))
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureOfAPendingEvaluationIsWritten() {
	server, _ := newFakeEvaluationRead(suite.T(), readPending)
	dir := filepath.Join(suite.T().TempDir(), "replay")

	_, _, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture --dir "+dir))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), evaluationInput, readFile(suite.T(), dir, "input.json"))
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureNeedsADirectory() {
	server, fake := newFakeEvaluationRead(suite.T(), readDeniedInline)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "--fixture needs --dir")
	require.Empty(suite.T(), fake.reads)
}

func (suite *GetEvaluationCommandTestSuite) TestADirectoryNeedsAFixture() {
	server, fake := newFakeEvaluationRead(suite.T(), readDeniedInline)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --dir "+suite.T().TempDir()))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "--dir needs --fixture")
	require.Empty(suite.T(), fake.reads)
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureAndShowInputAreRefusedTogether() {
	server, fake := newFakeEvaluationRead(suite.T(), readDeniedInline)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL,
		"01EVAL --fixture --dir "+suite.T().TempDir()+" --show-input -o json"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "only one of --fixture or --show-input can be given")
	require.Empty(suite.T(), fake.reads)
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureIsNotWrittenIntoADirectoryThatHasFiles() {
	server, fake := newFakeEvaluationRead(suite.T(), readDeniedInline)
	dir := suite.T().TempDir()
	require.NoError(suite.T(), os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("mine"), 0o644))

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture --dir "+dir))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, dir+" is not empty")
	require.Empty(suite.T(), fake.reads)
	require.Equal(suite.T(), "mine", readFile(suite.T(), dir, "keep.txt"))
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureRefusesAPolicyFileOutsideTheDirectory() {
	escaping := `{"id":"01EVAL","status":"completed","requested_at":1.0,"recorded_at":1.0,` +
		`"context":{"trails":[]},"policy":{"digest":"sha256:def456",` +
		`"files":{"../escaped.rego":"package policy\n"}},"params":{}}`
	server, _ := newFakeEvaluationRead(suite.T(), escaping)
	parent := suite.T().TempDir()
	dir := filepath.Join(parent, "replay")

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture --dir "+dir))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `"../escaped.rego" is not a path inside the policy directory`)
	require.NoFileExists(suite.T(), filepath.Join(dir, "escaped.rego"))
}

func (suite *GetEvaluationCommandTestSuite) TestAFixtureRefusesABundleEntryOutsideTheDirectory() {
	server, fake := newFakeEvaluationRead(suite.T(), readCompletedPublished)
	fake.bundle.body = gzippedTar(suite.T(), map[string]string{
		"../escaped.rego": "package policy\n",
	})
	dir := filepath.Join(suite.T().TempDir(), "replay")

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "01EVAL --fixture --dir "+dir))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `"../escaped.rego" is not a path inside the policy directory`)
	require.NoFileExists(suite.T(), filepath.Join(dir, "escaped.rego"))
}

// fakeEvaluationRead answers the evaluation, input and bundle routes with
// their body and status, 200 unless set, and records the escaped path of every
// request. Anything else is a 404.
type fakeEvaluationRead struct {
	evaluation fakeRoute
	input      fakeRoute
	bundle     fakeRoute
	reads      []string
}

type fakeRoute struct {
	body   string
	status int
}

func newFakeEvaluationRead(t *testing.T, evaluation string) (*httptest.Server, *fakeEvaluationRead) {
	t.Helper()
	fake := &fakeEvaluationRead{
		evaluation: fakeRoute{body: evaluation},
		input:      fakeRoute{body: evaluationInput},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.reads = append(fake.reads, r.URL.EscapedPath())
		route := fakeRoute{status: http.StatusNotFound}
		switch r.URL.EscapedPath() {
		case evaluationPath:
			route = fake.evaluation
		case evaluationInputPath:
			route = fake.input
		case publishedBundlePath:
			route = fake.bundle
		}
		if route.status == 0 {
			route.status = http.StatusOK
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(route.status)
		_, _ = io.WriteString(w, route.body)
	}))
	t.Cleanup(server.Close)
	return server, fake
}

func gzippedTar(t *testing.T, files map[string]string) string {
	t.Helper()
	var archive bytes.Buffer
	zipped := gzip.NewWriter(&archive)
	tw := tar.NewWriter(zipped)
	for name, body := range files {
		require.NoError(t, tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(body))}))
		_, err := tw.Write([]byte(body))
		require.NoError(t, err)
	}
	require.NoError(t, tw.Close())
	require.NoError(t, zipped.Close())
	return archive.String()
}

// requireRow asserts that a table row holds the label and the value, however
// wide the label column is.
func requireRow(t *testing.T, table, label, value string) {
	t.Helper()
	require.Regexp(t, "(?m)^"+regexp.QuoteMeta(label)+`:\s+`+regexp.QuoteMeta(value)+"$", table)
}

func readFile(t *testing.T, dir, name string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(dir, name))
	require.NoError(t, err)
	return string(content)
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	encoded, err := json.Marshal(value)
	require.NoError(t, err)
	return string(encoded)
}

func TestGetEvaluationCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetEvaluationCommandTestSuite))
}
