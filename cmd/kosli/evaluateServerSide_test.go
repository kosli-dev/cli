package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// fakeEvaluations stands in for the server-side evaluation API. It is a stub
// rather than the local test server on purpose: that stack runs a server, a
// database and object storage, with no queue, no worker and no evaluator, so
// an evaluation started there can never reach a verdict.
type fakeEvaluations struct {
	created    []map[string]interface{}
	reads      int
	trailReads int
	unexpected int
	verdict    string
}

const (
	createdPending = `{"id":"01EVAL","status":"pending","requested_at":1.0,"recorded_at":1.0}`

	verdictAllowed = `{"id":"01EVAL","status":"completed","requested_at":1.0,"recorded_at":1.0,` +
		`"result":{"allow":true,"violations":[]}}`
	verdictAllowedNoList = `{"id":"01EVAL","status":"completed","requested_at":1.0,"recorded_at":1.0,` +
		`"result":{"allow":true}}`
	verdictDenied = `{"id":"01EVAL","status":"completed","requested_at":1.0,"recorded_at":1.0,` +
		`"result":{"allow":false,"violations":["change is not approved"]}}`

	// What the local path prints for a verdict with nothing to report. The
	// server path has to match it byte for byte, or the two disagree on the
	// page while agreeing on the verdict.
	localAllowedJSON = "{\n  \"allow\": true,\n  \"violations\": null\n}"
)

func newFakeEvaluations(t *testing.T, verdict string) (*httptest.Server, *fakeEvaluations) {
	t.Helper()
	fake := &fakeEvaluations{verdict: verdict}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v2/evaluations/"):
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var body map[string]interface{}
			require.NoError(t, json.Unmarshal(raw, &body))
			fake.created = append(fake.created, body)
			w.WriteHeader(http.StatusCreated)
			_, _ = fmt.Fprint(w, createdPending)

		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v2/evaluations/"):
			fake.reads++
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, fake.verdict)

		// Only the client-side path asks for this, which is how a test proves
		// the two paths do not overlap.
		case strings.HasPrefix(r.URL.Path, "/api/v2/trails/"):
			fake.trailReads++
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"name":"my-trail","compliance_status":{"attestations_statuses":[]}}`)

		default:
			fake.unexpected++
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server, fake
}

type EvaluateServerSideTestSuite struct {
	suite.Suite
}

func (suite *EvaluateServerSideTestSuite) serverSideCmd(host, extra string) string {
	return fmt.Sprintf(
		"evaluate trail my-trail --flow my-flow --policy testdata/policies/allow-all.rego "+
			"--server-side --host %s --org test-org --api-token test-token --max-api-retries 0 %s",
		host, extra)
}

// The flag is deliberately undocumented: it exists to run the two evaluation
// paths against each other while neither is a contract anyone can rely on.
func (suite *EvaluateServerSideTestSuite) TestTheFlagIsHiddenFromHelp() {
	for _, command := range []string{"trail", "trails"} {
		suite.Run(command, func() {
			_, combined, _, _, err := executeCommandC("evaluate " + command + " --help")

			require.NoError(suite.T(), err)
			require.NotContains(suite.T(), combined, "--server-side")
			require.Contains(suite.T(), combined, "--policy", "the rest of the help still renders")
		})
	}
}

// `evaluate input` names no trails, so there is no evaluation to create from
// it and the flag must not be there to reach for.
func (suite *EvaluateServerSideTestSuite) TestEvaluateInputHasNoSuchFlag() {
	_, combined, _, _, err := executeCommandC(
		"evaluate input --input-file testdata/evaluate/trail-input.json " +
			"--policy testdata/policies/allow-all.rego --server-side")

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "unknown flag: --server-side")
	require.NotContains(suite.T(), combined, "ALLOWED")
}

func (suite *EvaluateServerSideTestSuite) TestItSendsThePolicyAndTheTrailAndPrintsTheVerdict() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)

	require.Len(suite.T(), fake.created, 1, "one evaluation, not one per trail")
	require.Equal(suite.T(), 0, fake.trailReads, "the server reads the trail, not us")
	require.Equal(suite.T(), 0, fake.unexpected)

	created := fake.created[0]
	context := created["context"].(map[string]interface{})
	require.Equal(suite.T(), []interface{}{
		map[string]interface{}{"flow": "my-flow", "trail": "my-trail"},
	}, context["trails"])

	files := created["policy"].(map[string]interface{})["files"].(map[string]interface{})
	require.Len(suite.T(), files, 1)
	require.Contains(suite.T(), files, "allow-all.rego", "the policy travels under its own name")
	require.Contains(suite.T(), files["allow-all.rego"], "package policy")
}

// The verdict is the same either way, so the page must be too.
func (suite *EvaluateServerSideTestSuite) TestItPrintsTheSameJsonAsTheLocalPath() {
	for _, test := range []struct {
		name    string
		verdict string
	}{
		{"the server sent an empty violations list", verdictAllowed},
		{"the server sent no violations list at all", verdictAllowedNoList},
	} {
		suite.Run(test.name, func() {
			server, _ := newFakeEvaluations(suite.T(), test.verdict)

			_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, "--output json"))

			require.NoError(suite.T(), err)
			require.Equal(suite.T(), localAllowedJSON, combined)
		})
	}
}

func (suite *EvaluateServerSideTestSuite) TestADenialIsPrintedAndAsserted() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err, "a denial exits non-zero by default")
	require.Contains(suite.T(), err.Error(), "policy denied")
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
	require.Contains(suite.T(), combined, "change is not approved")
}

func (suite *EvaluateServerSideTestSuite) TestADenialCanBeAskedNotToAssert() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, "--no-assert"))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
}

func (suite *EvaluateServerSideTestSuite) TestADenialStillAssertsInJson() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, "--output json"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `"allow": false`)
	require.Contains(suite.T(), combined, "change is not approved")
}

// Without the flag nothing about the command changes, which is the promise the
// whole ticket rests on.
func (suite *EvaluateServerSideTestSuite) TestWithoutTheFlagNoEvaluationIsCreated() {
	server, fake := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trail my-trail --flow my-flow --policy testdata/policies/allow-all.rego "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)
	require.Equal(suite.T(), 1, fake.trailReads, "the trail is read locally, as before")
	require.Empty(suite.T(), fake.created, "no evaluation is created")
	require.Equal(suite.T(), 0, fake.reads)
}

// The evaluate commands carry no --dry-run flag of their own, because until
// now they only ever read. The sentinel API token is the way in, and it has to
// keep working: a dry run must reach the network no more than any other.
func (suite *EvaluateServerSideTestSuite) TestADryRunSendsNothing() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trail my-trail --flow my-flow --policy testdata/policies/allow-all.rego "+
			"--server-side --host %s --org test-org --api-token DRY_RUN --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	require.Empty(suite.T(), fake.created)
	require.Equal(suite.T(), 0, fake.reads, "nothing was created, so there is nothing to wait for")
	require.Equal(suite.T(), 0, fake.unexpected)
}

// The server discovers a policy's entrypoint from the bundle, so it accepts
// policies the local evaluator refuses outright. Under the flag the CLI must
// not pre-judge them.
func (suite *EvaluateServerSideTestSuite) TestItUploadsAPolicyTheLocalPathWouldRefuse() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trail my-trail --flow my-flow --policy testdata/policies/no-package-policy.rego "+
			"--server-side --host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err, "the local package rule must not be applied under the flag")
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)

	files := fake.created[0]["policy"].(map[string]interface{})["files"].(map[string]interface{})
	require.Contains(suite.T(), files, "no-package-policy.rego")
}

// Every trail resolves at one instant, which is only true if they travel in
// one evaluation. One request per trail would smear the answer across however
// long the reads took.
func (suite *EvaluateServerSideTestSuite) TestEveryTrailGoesInOneEvaluation() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trails first second third --flow my-flow "+
			"--policy testdata/policies/allow-all.rego --server-side "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)
	require.Len(suite.T(), fake.created, 1, "one evaluation, however many trails")
	require.Equal(suite.T(), 0, fake.trailReads)

	context := fake.created[0]["context"].(map[string]interface{})
	require.Equal(suite.T(), []interface{}{
		map[string]interface{}{"flow": "my-flow", "trail": "first"},
		map[string]interface{}{"flow": "my-flow", "trail": "second"},
		map[string]interface{}{"flow": "my-flow", "trail": "third"},
	}, context["trails"], "named in the order given")
}

func (suite *EvaluateServerSideTestSuite) TestManyTrailsDenyAndAssertTogether() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trails first second --flow my-flow "+
			"--policy testdata/policies/allow-all.rego --server-side "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "policy denied")
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
}

func (suite *EvaluateServerSideTestSuite) TestManyTrailsCanBeAskedNotToAssert() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trails first second --flow my-flow "+
			"--policy testdata/policies/allow-all.rego --server-side --no-assert "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
}

// A repeat is stored once by the server rather than refused, so refusing it
// here would be stricter than the thing we are calling.
func (suite *EvaluateServerSideTestSuite) TestARepeatedTrailIsSentAsGiven() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trails same same --flow my-flow "+
			"--policy testdata/policies/allow-all.rego --server-side "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	context := fake.created[0]["context"].(map[string]interface{})
	require.Len(suite.T(), context["trails"], 2)
}

func (suite *EvaluateServerSideTestSuite) TestTheCeilingIsAHundredTrails() {
	suite.Run("a hundred are accepted", func() {
		server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"evaluate trails %s --flow my-flow "+
				"--policy testdata/policies/allow-all.rego --server-side "+
				"--host %s --org test-org --api-token test-token --max-api-retries 0",
			trailNames(100), server.URL))

		require.NoError(suite.T(), err)
		require.Len(suite.T(), fake.created[0]["context"].(map[string]interface{})["trails"], 100)
	})

	suite.Run("a hundred and one are refused before anything is sent", func() {
		server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"evaluate trails %s --flow my-flow "+
				"--policy testdata/policies/allow-all.rego --server-side "+
				"--host %s --org test-org --api-token test-token --max-api-retries 0",
			trailNames(101), server.URL))

		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "at most 100 trails")
		require.Contains(suite.T(), err.Error(), "101")
		require.Empty(suite.T(), fake.created, "refused here, so the server is never asked")
	})
}

func trailNames(count int) string {
	names := make([]string, count)
	for i := range names {
		names[i] = fmt.Sprintf("trail-%d", i)
	}
	return strings.Join(names, " ")
}

func TestEvaluateServerSideTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateServerSideTestSuite))
}
