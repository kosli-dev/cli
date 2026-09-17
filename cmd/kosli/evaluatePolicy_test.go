package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// The evaluation is never run here, so every test drives the stub the
// server-side suite already stands up: this repository's test environment has
// no evaluator to reach a verdict with.
type EvaluatePolicyCommandTestSuite struct {
	suite.Suite
}

func (suite *EvaluatePolicyCommandTestSuite) cmd(host, extra string) string {
	return fmt.Sprintf(
		"evaluate policy --flow my-flow --trail my-trail "+
			"--policy testdata/policies/allow-all.rego "+
			"--host %s --org test-org --api-token test-token --max-api-retries 0 %s",
		host, extra)
}

// The command evaluates where the policy runs, so the flags that only make
// sense on this machine are not offered at all. Hiding them would still leave
// them reachable; they are simply not there.
func (suite *EvaluatePolicyCommandTestSuite) TestItOffersOnlyItsOwnFlags() {
	_, combined, _, _, err := executeCommandC("evaluate policy --help")

	require.NoError(suite.T(), err)
	for _, flag := range []string{"--flow", "--trail", "--policy", "--params", "--output"} {
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
		{"flow", "evaluate policy --trail my-trail --policy testdata/policies/allow-all.rego"},
		{"trail", "evaluate policy --flow my-flow --policy testdata/policies/allow-all.rego"},
		{"policy", "evaluate policy --flow my-flow --trail my-trail"},
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
	context := created["context"].(map[string]interface{})
	require.Equal(suite.T(), []interface{}{
		map[string]interface{}{"flow": "my-flow", "trail": "my-trail"},
	}, context["trails"])

	files := created["policy"].(map[string]interface{})["files"].(map[string]interface{})
	require.Len(suite.T(), files, 1)
	require.Contains(suite.T(), files, "allow-all.rego", "the policy travels under its own name")
	require.Contains(suite.T(), files["allow-all.rego"], "package policy")
}

// No decision is asked for yet, and the block must be absent rather than sent
// empty: the body forbids what it does not name.
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
		want map[string]interface{}
	}{
		{"inline json", `--params '{"min_approvers":2}'`, map[string]interface{}{"min_approvers": float64(2)}},
		{"a file", "--params @testdata/evaluate/params-low-threshold.json", map[string]interface{}{"threshold": float64(3)}},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

			_, _, _, _, err := executeCommandC(suite.cmd(server.URL, test.flag))

			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.want, fake.created[0]["params"])
		})
	}
}

// The field is required by the API and defaults to empty, so no params must
// travel as an empty object rather than as a null.
func (suite *EvaluatePolicyCommandTestSuite) TestNoParamsTravelAsAnEmptyObject() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, _, _, _, err := executeCommandC(suite.cmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), map[string]interface{}{}, fake.created[0]["params"])
}

// A caller moving here from `evaluate trail` must not have to re-parse.
func (suite *EvaluatePolicyCommandTestSuite) TestItPrintsTheSameJsonAsEvaluateTrail() {
	server, _ := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, "--output json"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), localAllowedJSON, combined)
}

// Asserting arrives with --assert in the next slice. Until then a denial is
// reported in full and the command exits 0, so nothing half-built is exposed.
func (suite *EvaluatePolicyCommandTestSuite) TestADenialPrintsInFullAndExitsZero() {
	server, _ := newFakeEvaluations(suite.T(), verdictDenied)

	_, combined, _, _, err := executeCommandC(suite.cmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Regexp(suite.T(), `RESULT:\s+DENIED`, combined)
	require.Contains(suite.T(), combined, "change is not approved")
}

func (suite *EvaluatePolicyCommandTestSuite) TestADryRunSendsNothing() {
	server, fake := newFakeEvaluations(suite.T(), verdictAllowed)

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate policy --flow my-flow --trail my-trail "+
			"--policy testdata/policies/allow-all.rego "+
			"--host %s --org test-org --api-token DRY_RUN --max-api-retries 0", server.URL))

	require.NoError(suite.T(), err)
	require.Empty(suite.T(), fake.created)
	require.Equal(suite.T(), 0, fake.reads, "nothing was created, so there is no verdict to wait for")
	require.NotContains(suite.T(), combined, "RESULT")
}

func TestEvaluatePolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluatePolicyCommandTestSuite))
}
