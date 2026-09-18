package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// This repository's test environment has no evaluator, so every test drives
// the stub the server-side suite stands up.
type EvaluatePolicyCommandTestSuite struct {
	suite.Suite
}

func (suite *EvaluatePolicyCommandTestSuite) cmd(host, extra string) string {
	return fmt.Sprintf(
		"evaluate policy --context trail=my-flow/my-trail "+
			"--policy testdata/policies/allow-all.rego "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0 %s",
		host, extra)
}

// Hidden while the command is proved out against a server that can run it.
func (suite *EvaluatePolicyCommandTestSuite) TestTheCommandIsHiddenForNow() {
	_, listed, _, _, err := executeCommandC("evaluate --help")

	require.NoError(suite.T(), err)
	require.NotContains(suite.T(), listed, evaluatePolicyShortDesc)
	for _, sibling := range []string{"trail", "trails", "input"} {
		require.Contains(suite.T(), listed, sibling, "the rest of the listing still renders")
	}

	// Its own help still renders, for anyone told to try it.
	_, help, _, _, err := executeCommandC("evaluate policy --help")
	require.NoError(suite.T(), err)
	require.Contains(suite.T(), help, "--context")
}

// The flags that only make sense on this machine are absent rather than
// hidden, because a hidden flag is still reachable.
func (suite *EvaluatePolicyCommandTestSuite) TestItOffersOnlyItsOwnFlags() {
	_, combined, _, _, err := executeCommandC("evaluate policy --help")

	require.NoError(suite.T(), err)
	for _, flag := range []string{"--context", "--policy", "--params", "--output", "--assert"} {
		require.Contains(suite.T(), combined, flag)
	}
	for _, flag := range []string{"--server-side", "--attestations", "--show-input", "--no-assert", "--sync"} {
		require.NotContains(suite.T(), combined, flag)
	}
}

func (suite *EvaluatePolicyCommandTestSuite) TestItNamesTheRequiredFlagItWasNotGiven() {
	for _, test := range []struct {
		missing string
		cmd     string
	}{
		{"context", "evaluate policy --policy testdata/policies/allow-all.rego"},
		{"policy", "evaluate policy --context trail=my-flow/my-trail"},
	} {
		suite.Run(test.missing, func() {
			_, _, _, _, err := executeCommandC(test.cmd + " --org test-org --api-token test-token")

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), test.missing)
		})
	}
}

func (suite *EvaluatePolicyCommandTestSuite) TestItSendsThePolicyAndTheTrailAndPrintsTheVerdict() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)

	require.Len(suite.T(), fake.created, 1)
	require.Equal(suite.T(), 0, fake.trailReads, "the trail is read where the policy runs")
	require.Equal(suite.T(), 0, fake.unexpected)

	created := fake.created[0]
	context := created["context"].(map[string]any)
	require.Equal(suite.T(), []any{
		map[string]any{"flow": "my-flow", "trail": "my-trail"},
	}, context["trails"])

	files := created["policy"].(map[string]any)["files"].(map[string]any)
	require.Len(suite.T(), files, 1)
	require.Contains(suite.T(), files, "allow-all.rego", "the policy travels under its own name")
	require.Contains(suite.T(), files["allow-all.rego"], "package policy")
}

// The body forbids what it does not name, so an unasked-for decision block is
// absent rather than empty.
func (suite *EvaluatePolicyCommandTestSuite) TestItAsksForNoDecision() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.NotContains(suite.T(), fake.created[0], "decision")
}

func (suite *EvaluatePolicyCommandTestSuite) TestItPassesParamsOnUnchanged() {
	for _, test := range []struct {
		name string
		flag string
		want map[string]any
	}{
		{"inline json", `--params '{"min_approvers":2}'`, map[string]any{"min_approvers": float64(2)}},
		{"a file", "--params @testdata/evaluate/params-low-threshold.json", map[string]any{"threshold": float64(3)}},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, _, _, _, err := executeCommandC(suite.cmd(server.URL, test.flag))

			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.want, fake.created[0]["params"])
		})
	}
}

// The API's params field rejects a null where it accepts an empty object.
func (suite *EvaluatePolicyCommandTestSuite) TestNoParamsTravelAsAnEmptyObject() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), map[string]any{}, fake.created[0]["params"])
}

// A caller moving here from `evaluate trail` must not have to re-parse.
func (suite *EvaluatePolicyCommandTestSuite) TestItPrintsTheSameJsonAsEvaluateTrail() {
	server, _ := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, "--output json"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), localAllowedJSON, combined)
}

// Asserting is opt-in: recording a decision is not a reason to fail the step
// that asked for it.
func (suite *EvaluatePolicyCommandTestSuite) TestADenialPrintsInFullAndExitsZero() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
	require.Contains(suite.T(), combined, "change is not approved")
}

func (suite *EvaluatePolicyCommandTestSuite) TestAssertFailsOnADenial() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, "--assert"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "policy denied")
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
	require.Contains(suite.T(), combined, "change is not approved")
}

func (suite *EvaluatePolicyCommandTestSuite) TestAssertPassesOnAnAllow() {
	server, _ := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, "--assert"))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)
}

// The verdict is printed before the assertion, so the page is the same
// whichever exit code follows.
func (suite *EvaluatePolicyCommandTestSuite) TestAssertStillPrintsTheVerdictInJson() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, "--assert --output json"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `"allow": false`)
	require.Contains(suite.T(), combined, "change is not approved")
}

// An evaluation that has not answered is never a verdict, with or without
// --assert.
func (suite *EvaluatePolicyCommandTestSuite) TestAnExpiredWaitNamesTheEvaluationAndNoVerdict() {
	for _, extra := range []string{"", "--assert"} {
		suite.Run("with "+extra, func() {
			original := serverSideWaitOptions
			serverSideWaitOptions = evaluations.WaitOptions{
				Timeout: 20 * time.Millisecond,
				Initial: time.Millisecond,
				Max:     2 * time.Millisecond,
			}
			defer func() { serverSideWaitOptions = original }()

			server, _ := newFakeEvaluations(suite.T(), createdPending)

			_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, extra))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), "still pending")
			require.Contains(suite.T(), err.Error(), "01EVAL")
			require.NotContains(suite.T(), combined, "DENIED")
			require.NotContains(suite.T(), combined, "ALLOWED")
		})
	}
}

// Every trail resolves at one instant, which holds only if they travel in one
// evaluation.
func (suite *EvaluatePolicyCommandTestSuite) TestEveryContextGoesInOneEvaluation() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--context trail=other-flow/second --context trail=my-flow/third"))

	require.NoError(suite.T(), err)
	require.Len(suite.T(), fake.created, 1, "one evaluation, however many trails")

	context := fake.created[0]["context"].(map[string]any)
	require.Equal(suite.T(), []any{
		map[string]any{"flow": "my-flow", "trail": "my-trail"},
		map[string]any{"flow": "other-flow", "trail": "second"},
		map[string]any{"flow": "my-flow", "trail": "third"},
	}, context["trails"], "named in the order given")
}

// The API stores a repeat once rather than refusing it, so refusing it here
// would be stricter than the thing being called.
func (suite *EvaluatePolicyCommandTestSuite) TestARepeatedContextIsSentAsGiven() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL, "--context trail=my-flow/my-trail"))

	require.NoError(suite.T(), err)
	context := fake.created[0]["context"].(map[string]any)
	require.Len(suite.T(), context["trails"], 2)
}

func (suite *EvaluatePolicyCommandTestSuite) TestAMalformedContextIsRefusedBeforeAnyRequest() {
	for _, test := range []struct {
		name  string
		value string
	}{
		{"no key", "my-flow/my-trail"},
		{"an unknown key", "artifact=my-flow/my-trail"},
		{"no flow and trail", "trail=my-trail"},
		{"an empty value", "trail="},
		{"an empty flow", "trail=/my-trail"},
		{"an empty trail", "trail=my-flow/"},
		{"more than a flow and a trail", "trail=my-flow/my-trail/extra"},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, combined, _, _, err := executeCommandC(fmt.Sprintf(
				"evaluate policy --context %s --policy testdata/policies/allow-all.rego "+
					"--host %s --org test-org --api-token test-token --max-api-retries 0",
				test.value, server.URL))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), "trail=<flow>/<trail>",
				"the refusal names the form it expects")
			require.Empty(suite.T(), fake.created, "nothing is sent")
			require.NotContains(suite.T(), combined, "RESULT")
		})
	}
}

func (suite *EvaluatePolicyCommandTestSuite) TestTooManyContextsAreRefusedBeforeAnyRequest() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	contexts := ""
	for i := 0; i <= maxServerSideTrails; i++ {
		contexts += fmt.Sprintf("--context trail=my-flow/trail-%d ", i)
	}

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate policy %s--policy testdata/policies/allow-all.rego "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0",
		contexts, server.URL))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), fmt.Sprintf("%d", maxServerSideTrails))
	require.Empty(suite.T(), fake.created)
}

func (suite *EvaluatePolicyCommandTestSuite) TestControlRecordsADecisionWhereItIsTold() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail "+
			"--fingerprint b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), map[string]any{
		"control":     "SDLC-CTRL-0007",
		"name":        "SDLC-CTRL-0007-decision",
		"flow":        "release",
		"trail":       "my-trail",
		"fingerprint": "b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
	}, fake.created[0]["decision"])
}

// The default is computed here and sent, so the name is one the caller can
// predict.
func (suite *EvaluatePolicyCommandTestSuite) TestTheDecisionNameDefaultsToTheControl() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail"))

	require.NoError(suite.T(), err)
	decision := fake.created[0]["decision"].(map[string]any)
	require.Equal(suite.T(), "SDLC-CTRL-0007-decision", decision["name"])
	require.NotContains(suite.T(), decision, "fingerprint",
		"a decision about the trail carries no fingerprint")
}

func (suite *EvaluatePolicyCommandTestSuite) TestAGivenNameIsSentAsGiven() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail --name code-review-decision"))

	require.NoError(suite.T(), err)
	decision := fake.created[0]["decision"].(map[string]any)
	require.Equal(suite.T(), "code-review-decision", decision["name"])
}

// A pipeline sets the flow and the trail for every command it runs, so
// carrying them is not a reason to refuse a run that asked for no decision.
func (suite *EvaluatePolicyCommandTestSuite) TestADestinationWithoutAControlRecordsNothing() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, "--flow release --trail my-trail"))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+ALLOWED`, combined)
	require.NotContains(suite.T(), fake.created[0], "decision")
}

func (suite *EvaluatePolicyCommandTestSuite) TestWhatCannotBeAskedForIsRefusedBeforeAnyRequest() {
	for _, test := range []struct {
		name    string
		extra   string
		says    []string
		saysNot []string
	}{
		{"a name with no control", "--name code-review-decision",
			[]string{"--name", "--control"}, []string{"--fingerprint"}},
		{"a fingerprint with no control",
			"--fingerprint b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
			[]string{"--fingerprint", "--control"}, []string{"--name"}},
		{"a name and a fingerprint with no control",
			"--name code-review-decision " +
				"--fingerprint b5bb9d8014a0f9b1d61e21e796d78dccdf1352f23cd32812f4850b878ae4944c",
			[]string{"--name", "--fingerprint", "--control"}, nil},
		{"a control with no destination", "--control SDLC-CTRL-0007",
			[]string{"--flow", "--trail"}, nil},
		{"a control with no trail", "--control SDLC-CTRL-0007 --flow release",
			[]string{"--trail"}, nil},
		// A malformed fingerprint is ours to catch, as on every other command
		// that takes one.
		{"a fingerprint that is not a SHA256",
			"--control SDLC-CTRL-0007 --flow release --trail my-trail --fingerprint sha256:abc",
			[]string{"not a valid SHA256"}, nil},
		// Refused before the request, which with --control records a decision.
		{"an output format that does not exist",
			"--control SDLC-CTRL-0007 --flow release --trail my-trail --output tabel",
			[]string{"tabel", "table", "json"}, nil},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, test.extra))

			require.Error(suite.T(), err)
			for _, says := range test.says {
				require.Contains(suite.T(), err.Error(), says)
			}
			for _, saysNot := range test.saysNot {
				require.NotContains(suite.T(), err.Error(), saysNot,
					"the refusal names what was typed, not what was not")
			}
			require.Empty(suite.T(), fake.created, "nothing is sent")
			require.NotContains(suite.T(), combined, "RESULT")
		})
	}
}

// The environment satisfies the destination as the flags do.
func (suite *EvaluatePolicyCommandTestSuite) TestTheDestinationCanComeFromTheEnvironment() {
	suite.T().Setenv("KOSLI_FLOW", "release")
	suite.T().Setenv("KOSLI_TRAIL", "my-trail")
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL, "--control SDLC-CTRL-0007"))

	require.NoError(suite.T(), err)
	decision := fake.created[0]["decision"].(map[string]any)
	require.Equal(suite.T(), "release", decision["flow"])
	require.Equal(suite.T(), "my-trail", decision["trail"])
}

// A denial is a decision: it is recorded whether or not the caller asked the
// command to fail.
func (suite *EvaluatePolicyCommandTestSuite) TestADenialAsksForItsDecisionToo() {
	server, fake := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail --assert"))

	require.Error(suite.T(), err, "--assert still fails the step")
	require.Contains(suite.T(), err.Error(), "policy denied")
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
	require.Contains(suite.T(), fake.created[0], "decision")
}

func (suite *EvaluatePolicyCommandTestSuite) TestTheRecordedDecisionIsReported() {
	server, _ := newFakeEvaluations(suite.T(), verdictAllowedWithDecision)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "01DECISION")
}

func (suite *EvaluatePolicyCommandTestSuite) TestTheRecordedDecisionIsReportedInJson() {
	server, _ := newFakeEvaluations(suite.T(), verdictAllowedWithDecision)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail --output json"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, `"decision_attestation_id": "01DECISION"`)
}

// A refusal is reported in the API's own words, whatever it refused, so the
// command needs no opinion about each one.
func (suite *EvaluatePolicyCommandTestSuite) TestARefusalIsReportedAsTheServerPutIt() {
	for _, test := range []struct {
		name    string
		status  int
		message string
	}{
		{"an unknown control", 404, "Control 'SDLC-CTRL-0007' does not exist in org 'test-org'"},
		{"an unresolvable trail", 404, "These trails do not exist in org 'test-org': my-flow/my-trail"},
		{"an archived destination", 400, "Flow named 'release' has been archived for organization 'test-org'"},
		{"an organisation without the entitlement", 403,
			"Server-side evaluation is not enabled for this organization"},
	} {
		suite.Run(test.name, func() {
			server := newRefusingServer(suite.T(), test.status,
				fmt.Sprintf(`{"message":%q}`, test.message))

			_, combined, _, _, err := executeCommandC(suite.cmd(server.URL,
				"--control SDLC-CTRL-0007 --flow release --trail my-trail"))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), test.message)
			require.Contains(suite.T(), err.Error(), fmt.Sprintf("%d", test.status))
			require.NotContains(suite.T(), combined, "RESULT")
		})
	}
}

// A policy that could not run has decided nothing: it is never a denial, and
// nothing is recorded.
func (suite *EvaluatePolicyCommandTestSuite) TestAClassifiedFailureIsNotADenial() {
	server, _ := newFakeEvaluations(suite.T(),
		`{"id":"01EVAL","status":"failed","requested_at":1.0,"recorded_at":1.0,`+
			`"error":{"kind":"compile","message":"policy.rego:4: unexpected token"}}`)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL,
		"--control SDLC-CTRL-0007 --flow release --trail my-trail --assert"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "compile")
	require.Contains(suite.T(), err.Error(), "policy.rego:4: unexpected token")
	require.NotContains(suite.T(), err.Error(), "denied")
	require.NotContains(suite.T(), combined, "DENIED")
}

func (suite *EvaluatePolicyCommandTestSuite) TestADirectoryTravelsAsOneBundle() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate policy --context trail=my-flow/my-trail --policy testdata/policies/bundle "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	files := fake.created[0]["policy"].(map[string]any)["files"].(map[string]any)
	require.Equal(suite.T(), []string{"README.md", "lib/helpers.rego", "policy.rego"}, sortedKeys(files),
		"keyed by path relative to the directory, and nothing left out by name")
	require.Contains(suite.T(), files["policy.rego"], "package policy")
	require.Contains(suite.T(), files["lib/helpers.rego"], "package lib.helpers")
}

// The API takes at least one file.
func (suite *EvaluatePolicyCommandTestSuite) TestAnEmptyDirectoryIsRefused() {
	directory := suite.T().TempDir()
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate policy --context trail=my-flow/my-trail --policy %s "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0", directory, server.URL))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), directory)
	require.Empty(suite.T(), fake.created)
}

func (suite *EvaluatePolicyCommandTestSuite) TestABundleOverTheCapsIsRefusedWithTheCapNamed() {
	for _, test := range []struct {
		name  string
		build func(string)
		says  string
	}{
		{"too many files", func(directory string) {
			for i := 0; i <= maxPolicyBundleFiles; i++ {
				require.NoError(suite.T(), os.WriteFile(
					filepath.Join(directory, fmt.Sprintf("policy-%d.rego", i)),
					[]byte("package policy\n"), 0644))
			}
		}, fmt.Sprintf("%d", maxPolicyBundleFiles)},
		{"too many bytes", func(directory string) {
			source := make([]byte, serverPolicyMaxBytes+1)
			for i := range source {
				source[i] = 'a'
			}
			require.NoError(suite.T(), os.WriteFile(
				filepath.Join(directory, "policy.rego"), source, 0644))
		}, fmt.Sprintf("%d", serverPolicyMaxBytes)},
	} {
		suite.Run(test.name, func() {
			directory := suite.T().TempDir()
			test.build(directory)
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, _, _, _, err := executeCommandC(fmt.Sprintf(
				"evaluate policy --context trail=my-flow/my-trail --policy %s "+
					"--host %s --org test-org --api-token test-token --max-api-retries 0",
				directory, server.URL))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), test.says)
			require.Empty(suite.T(), fake.created, "nothing is sent")
		})
	}
}

// A policy comes from the machine that runs the command.
func (suite *EvaluatePolicyCommandTestSuite) TestARemotePolicyIsRefused() {
	for _, ref := range []string{"http://policies.example.com/pr.rego", "https://policies.example.com/pr.rego"} {
		suite.Run(ref, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, combined, _, _, err := executeCommandC(fmt.Sprintf(
				"evaluate policy --context trail=my-flow/my-trail --policy %s "+
					"--host %s --org test-org --api-token test-token --max-api-retries 0", ref, server.URL))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), "--policy")
			require.Empty(suite.T(), fake.created, "nothing is sent")
			require.NotContains(suite.T(), combined, "RESULT")
		})
	}
}

func (suite *EvaluatePolicyCommandTestSuite) TestADryRunSendsNothing() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate policy --context trail=my-flow/my-trail "+
			"--policy testdata/policies/allow-all.rego "+
			"--host %s --org test-org --api-token DRY_RUN --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	require.Empty(suite.T(), fake.created)
	require.Equal(suite.T(), 0, fake.reads, "nothing was created, so there is no verdict to wait for")
	require.NotContains(suite.T(), combined, "RESULT")
}

func sortedKeys(files map[string]any) []string {
	keys := make([]string, 0, len(files))
	for key := range files {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func TestEvaluatePolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluatePolicyCommandTestSuite))
}
