package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/evaluations"
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

// Two flags have no server-side answer. Refusing them is kinder than
// accepting them and quietly doing something else: a filter that is ignored
// would evaluate more than the caller asked about, and an input shown from
// here would not be the input the server judged.
func (suite *EvaluateServerSideTestSuite) TestItRefusesWhatTheServerCannotDo() {
	for _, test := range []struct {
		name    string
		extra   string
		message string
	}{
		{
			name:    "filtering attestations",
			extra:   "--attestations some-attestation",
			message: "--attestations is not supported with --server-side",
		},
		{
			name:    "showing the policy input",
			extra:   "--show-input",
			message: "--show-input is not supported with --server-side",
		},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, test.extra))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), test.message)
			require.Empty(suite.T(), fake.created, "refused before anything is sent")
			require.Equal(suite.T(), 0, fake.reads)
			require.NotContains(suite.T(), combined, "ALLOWED")
		})
	}
}

// Both trail commands share the refusal, so neither can drift from the other.
func (suite *EvaluateServerSideTestSuite) TestTheRefusalCoversTheMultiTrailCommandToo() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trails first second --flow my-flow --show-input "+
			"--policy testdata/policies/allow-all.rego --server-side "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "--show-input is not supported with --server-side")
	require.Empty(suite.T(), fake.created)
}

func (suite *EvaluateServerSideTestSuite) TestPolicyParametersTravelUnchanged() {
	for _, test := range []struct {
		name  string
		extra string
		want  map[string]interface{}
	}{
		{
			name:  "given inline",
			extra: `--params '{"threshold":2}'`,
			want:  map[string]interface{}{"threshold": float64(2)},
		},
		{
			name:  "read from a file",
			extra: "--params @testdata/evaluate/params-low-threshold.json",
			want:  map[string]interface{}{"threshold": float64(3)},
		},
		{
			name:  "not given at all",
			extra: "",
			want:  map[string]interface{}{},
		},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, test.extra))

			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.want, fake.created[0]["params"])
		})
	}
}

func (suite *EvaluateServerSideTestSuite) TestUnreadableParametersAreRefusedBeforeAnythingIsSent() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, "--params not-json"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "failed to parse --params")
	require.Empty(suite.T(), fake.created)
}

// A policy that could not run has decided nothing. Reporting any of these as a
// denial would block a release over a typo in the policy.
func (suite *EvaluateServerSideTestSuite) TestABrokenPolicyIsNeverADenial() {
	for _, kind := range []string{
		"no_policy", "entrypoint", "compile", "result_shape",
		"input_shape", "evaluate", "enqueue_failed",
		// Not one of the agreed kinds. The set belongs to the evaluator, so an
		// unknown one has to reach the user rather than be flattened away.
		"future_kind",
	} {
		suite.Run(kind, func() {
			server, _ := newFakeEvaluations(suite.T(), fmt.Sprintf(
				`{"id":"01EVAL","status":"failed","requested_at":1.0,"recorded_at":1.0,`+
					`"error":{"kind":%q,"message":"policy.rego:4: something is wrong"}}`, kind))

			_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), kind)
			require.Contains(suite.T(), err.Error(), "policy.rego:4: something is wrong")
			require.NotContains(suite.T(), combined, "DENIED")
			require.NotContains(suite.T(), combined, "ALLOWED")
		})
	}
}

func (suite *EvaluateServerSideTestSuite) TestAFailureWithNoReasonIsStillNotADenial() {
	server, _ := newFakeEvaluations(suite.T(),
		`{"id":"01EVAL","status":"failed","requested_at":1.0,"recorded_at":1.0}`)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "01EVAL")
	require.NotContains(suite.T(), combined, "DENIED")
}

func (suite *EvaluateServerSideTestSuite) TestAnUnentitledOrgIsToldWhatToDo() {
	server := newRefusingServer(suite.T(), http.StatusForbidden,
		`{"message":"Server-side evaluation is not enabled for this organization"}`)

	_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "test-org", "name the org that was refused")
	require.Contains(suite.T(), err.Error(), "is-server-side-evaluation-enabled")
	require.Contains(suite.T(), err.Error(), "--server-side", "say how to carry on regardless")
}

// A Kosli server old enough to lack the endpoint answers 404 with no message
// field, which is how it is told apart from a 404 about a trail.
func (suite *EvaluateServerSideTestSuite) TestAnOlderServerIsNamedAsSuch() {
	server := newRefusingServer(suite.T(), http.StatusNotFound, `{"detail":"Not Found"}`)

	_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "does not support server-side evaluation")
	require.Contains(suite.T(), err.Error(), "--server-side")
	require.NotContains(suite.T(), err.Error(), "map[", "no internal rendering leaks out")
}

func (suite *EvaluateServerSideTestSuite) TestAServerRefusalIsPassedOnInItsOwnWords() {
	for _, test := range []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{
			name:   "a trail that does not exist",
			status: http.StatusNotFound,
			body:   `{"message":"These trails do not exist in org 'test-org': my-flow/my-trail"}`,
			want:   "These trails do not exist in org 'test-org': my-flow/my-trail",
		},
		{
			name:   "a policy over the byte cap",
			status: http.StatusBadRequest,
			body:   `{"message":"policy bundle is 1048600 bytes, over the 1048576 byte limit"}`,
			want:   "policy bundle is 1048600 bytes, over the 1048576 byte limit",
		},
	} {
		suite.Run(test.name, func() {
			server := newRefusingServer(suite.T(), test.status, test.body)

			_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), test.want)
		})
	}
}

// The shared HTTP client retries a 5xx, gives up and throws the body away, so
// the server's own sentence is gone by the time it reaches here.
func (suite *EvaluateServerSideTestSuite) TestAServerFaultGetsASentenceOfOurOwn() {
	server := newRefusingServer(suite.T(), http.StatusServiceUnavailable,
		`{"message":"Evaluation '01EVAL' could not be queued"}`)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "could not get a server-side evaluation")
	require.NotContains(suite.T(), err.Error(), "could not be queued", "the body is gone by now")
	require.NotContains(suite.T(), combined, "DENIED")
}

func (suite *EvaluateServerSideTestSuite) TestAWaitThatExpiresNamesTheEvaluationAndNoVerdict() {
	original := evaluations.DefaultWaitOptions
	evaluations.DefaultWaitOptions = evaluations.WaitOptions{
		Timeout: 20 * time.Millisecond,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
	defer func() { evaluations.DefaultWaitOptions = original }()

	server, _ := newFakeEvaluations(suite.T(), createdPending)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "still pending")
	require.Contains(suite.T(), err.Error(), "01EVAL", "name the evaluation still running")
	require.NotContains(suite.T(), combined, "DENIED")
	require.NotContains(suite.T(), combined, "ALLOWED")
}

// newRefusingServer answers every request with one status and body, which is
// how a create that is refused outright is exercised.
func newRefusingServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprint(w, body)
	}))
	t.Cleanup(server.Close)
	return server
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
