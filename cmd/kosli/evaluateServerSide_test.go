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
	// createdBody answers the create, so a test can shape what comes back
	// from it; empty means the ordinary pending answer.
	createdBody string
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
			answer := fake.createdBody
			if answer == "" {
				answer = createdPending
			}
			_, _ = fmt.Fprint(w, answer)

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

// Something that does not serve this route answers 404 without a message of
// its own, which is how it is told apart from a 404 about a trail. It can say
// so in more than one way, and none of them should reach a user raw.
func (suite *EvaluateServerSideTestSuite) TestAnOlderServerIsNamedAsSuch() {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "a server that routes nothing here", body: `{"detail":"Not Found"}`},
		{name: "a proxy answering in html", body: `<html><body>404 Not Found</body></html>`},
		{name: "an answer with no body at all", body: ``},
		{name: "a bare string where an object was expected", body: `"Not Found"`},
	} {
		suite.Run(test.name, func() {
			server := newRefusingServer(suite.T(), http.StatusNotFound, test.body)

			_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), "does not support server-side evaluation")
			require.Contains(suite.T(), err.Error(), "--server-side")
			require.NotContains(suite.T(), err.Error(), "map[", "no internal rendering leaks out")
			require.NotContains(suite.T(), err.Error(), "invalid character",
				"no decoder complaint leaks out")
			require.NotContains(suite.T(), err.Error(), "unexpected end of JSON input")
		})
	}
}

// A refusal can arrive from something in front of Kosli, which has no sentence
// of the API's to quote. Quoting what it did send would dress a decoder
// complaint as the server's reason, exactly as the 404 branch once did.
func (suite *EvaluateServerSideTestSuite) TestARefusalWithNothingToQuoteSaysOnlyWhatIsKnown() {
	for _, test := range []struct {
		name string
		body string
	}{
		{name: "a proxy answering in html", body: `<html><body>403 Forbidden</body></html>`},
		{name: "an answer with no body at all", body: ``},
		{name: "an object with no message", body: `{"detail":"Forbidden"}`},
		{name: "a bare string", body: `"Forbidden"`},
		{name: "a message key holding nothing", body: `{"message":""}`},
		// The client trims this phrase out of a message, which can leave
		// nothing behind even though the key was there.
		{name: "a message trimmed away to nothing", body: `{"message":"You have requested a thing"}`},
	} {
		suite.Run(test.name, func() {
			server := newRefusingServer(suite.T(), http.StatusForbidden, test.body)

			_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), "test-org")
			require.Contains(suite.T(), err.Error(), "is-server-side-evaluation-enabled")
			require.NotContains(suite.T(), err.Error(), "invalid character",
				"no decoder complaint dressed as the server's reason")
			require.NotContains(suite.T(), err.Error(), "unexpected end of JSON input")
			require.NotContains(suite.T(), err.Error(), "map[")
			require.NotContains(suite.T(), err.Error(), "': .",
				"no stray punctuation where a sentence would have gone")
		})
	}
}

// An API error prints as its message alone, so one that arrived without a
// message would otherwise print as nothing at all.
func (suite *EvaluateServerSideTestSuite) TestARefusalWithNoMessageStillSaysWhatHappened() {
	server := newRefusingServer(suite.T(), http.StatusBadRequest, `{"message":""}`)

	_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "400")
	require.Contains(suite.T(), err.Error(), "said nothing about why")
}

// A refusal can come from a token without rights on the org rather than from
// the feature flag, so the server's own reason has to travel with ours.
func (suite *EvaluateServerSideTestSuite) TestARefusalKeepsTheServersOwnReason() {
	server := newRefusingServer(suite.T(), http.StatusForbidden,
		`{"message":"API token does not have access to this organization"}`)

	_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "API token does not have access to this organization")
	require.Contains(suite.T(), err.Error(), "test-org")
}

// The status decides, not what else the answer happens to carry.
func (suite *EvaluateServerSideTestSuite) TestAFailureCarryingAResultIsStillNotAVerdict() {
	server, _ := newFakeEvaluations(suite.T(),
		`{"id":"01EVAL","status":"failed","requested_at":1.0,"recorded_at":1.0,`+
			`"result":{"allow":true,"violations":[]},`+
			`"error":{"kind":"compile","message":"policy does not compile"}}`)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "compile")
	require.NotContains(suite.T(), combined, "ALLOWED", "a failure never prints a verdict")
	require.NotContains(suite.T(), combined, "DENIED")
}

// Without an id there is nothing to read the verdict back from, and asking
// anyway would fetch a different resource and then blame the answer.
func (suite *EvaluateServerSideTestSuite) TestAnEvaluationWithNoIdIsRefusedNotPolled() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)
	fake.createdBody = `{"status":"pending","requested_at":1.0,"recorded_at":1.0}`

	_, _, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "named no id")
	require.Equal(suite.T(), 0, fake.reads, "nothing to read back, so nothing is read")
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
	original := serverSideWaitOptions
	serverSideWaitOptions = evaluations.WaitOptions{
		Timeout: 20 * time.Millisecond,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
	defer func() { serverSideWaitOptions = original }()

	server, _ := newFakeEvaluations(suite.T(), createdPending)

	_, combined, _, _, err := executeCommandC(suite.serverSideCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "still pending")
	require.Contains(suite.T(), err.Error(), "01EVAL", "name the evaluation still running")
	require.NotContains(suite.T(), err.Error(), "could not get a server-side evaluation",
		"the server answered, so this is not a transport failure")
	require.NotContains(suite.T(), combined, "DENIED")
	require.NotContains(suite.T(), combined, "ALLOWED")
}

// The server never fetches a customer URL, so a remote policy is fetched here
// and its source is what travels.
func (suite *EvaluateServerSideTestSuite) TestARemotePolicyIsFetchedAndItsSourceUploaded() {
	policyServer := newPolicyServer(suite.T(), "package policy\n\nallow := true\n")
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trail my-trail --flow my-flow --policy %s/policies/pr.rego --server-side "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0",
		policyServer.URL, server.URL))

	require.NoError(suite.T(), err)

	files := fake.created[0]["policy"].(map[string]interface{})["files"].(map[string]interface{})
	require.Len(suite.T(), files, 1)
	require.Contains(suite.T(), files, "pr.rego", "named after the file, not the host")
	require.Equal(suite.T(), "package policy\n\nallow := true\n", files["pr.rego"])

	for name := range files {
		require.NotContains(suite.T(), name, policyServer.URL, "the URL never travels")
	}
}

func (suite *EvaluateServerSideTestSuite) TestThePolicyIsNamedByItsFileAlone() {
	for _, test := range []struct {
		name   string
		policy string
		want   string
	}{
		{
			name:   "a path that climbs out and back",
			policy: "testdata/policies/../policies/allow-all.rego",
			want:   "allow-all.rego",
		},
		{
			name:   "a path with a leading dot",
			policy: "./testdata/policies/allow-all.rego",
			want:   "allow-all.rego",
		},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, _, _, _, err := executeCommandC(fmt.Sprintf(
				"evaluate trail my-trail --flow my-flow --policy %s --server-side "+
					"--host %s --org test-org --api-token test-token --max-api-retries 0",
				test.policy, server.URL))

			require.NoError(suite.T(), err)
			files := fake.created[0]["policy"].(map[string]interface{})["files"].(map[string]interface{})
			require.Contains(suite.T(), files, test.want)
			for name := range files {
				require.NotContains(suite.T(), name, "..", "a bundle path never climbs out")
				require.NotContains(suite.T(), name, "/", "a bundle path is a file, not a route")
			}
		})
	}
}

// A URL ending in a slash names no file, so the bundle needs a name of its own
// rather than one derived from nothing.
func (suite *EvaluateServerSideTestSuite) TestAPolicyUrlNamingNoFileStillGetsAName() {
	policyServer := newPolicyServer(suite.T(), "package policy\n\nallow := true\n")
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate trail my-trail --flow my-flow --policy %s/ --server-side "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0",
		policyServer.URL, server.URL))

	require.NoError(suite.T(), err)
	files := fake.created[0]["policy"].(map[string]interface{})["files"].(map[string]interface{})
	require.Contains(suite.T(), files, "policy.rego")
}

func (suite *EvaluateServerSideTestSuite) TestAPolicyOverTheApiCapIsRefusedHere() {
	suite.Run("a local policy just over the cap", func() {
		server, fake := newFakeEvaluations(suite.T(), verdictAllowed)
		policy := writePolicyFile(suite.T(), "big.rego", 1048576)

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"evaluate trail my-trail --flow my-flow --policy %s --server-side "+
				"--host %s --org test-org --api-token test-token --max-api-retries 0",
			policy, server.URL))

		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "over the 1048576 byte limit")
		require.Empty(suite.T(), fake.created, "refused here, so the server is never asked")
	})

	suite.Run("a local policy exactly at the cap", func() {
		server, fake := newFakeEvaluations(suite.T(), verdictAllowed)
		// The cap counts the name as well as the source, as the API does.
		policy := writePolicyFile(suite.T(), "big.rego", 1048576-len("big.rego"))

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"evaluate trail my-trail --flow my-flow --policy %s --server-side "+
				"--host %s --org test-org --api-token test-token --max-api-retries 0",
			policy, server.URL))

		require.NoError(suite.T(), err)
		require.Len(suite.T(), fake.created, 1)
	})

	// The remote read allows five times what the API accepts, so a policy can
	// be fetched in full and still be too big to send.
	suite.Run("a remote policy the fetch allows but the API would not", func() {
		policyServer := newPolicyServer(suite.T(), strings.Repeat("x", 2<<20))
		server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"evaluate trail my-trail --flow my-flow --policy %s/big.rego --server-side "+
				"--host %s --org test-org --api-token test-token --max-api-retries 0",
			policyServer.URL, server.URL))

		require.Error(suite.T(), err)
		require.Contains(suite.T(), err.Error(), "over the 1048576 byte limit")
		require.Empty(suite.T(), fake.created)
	})
}

func newPolicyServer(t *testing.T, source string) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, source)
	}))
	t.Cleanup(server.Close)
	return server
}

func writePolicyFile(t *testing.T, name string, size int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(strings.Repeat("x", size)), 0600))
	return path
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
