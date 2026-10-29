package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Runs against the local server, which has no validator, so a version it
// publishes stays pending: enough to see a pinned version reach the server and
// be refused there, though never enough to see one run. It cannot queue the
// validation either, so a publish stores the version and still answers 503.
type EvaluatePolicyPinnedTestSuite struct {
	suite.Suite
	identifier string
}

func (suite *EvaluatePolicyPinnedTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	// Each run publishes afresh, so a reused identifier would reach a later
	// version on every run.
	suite.identifier = fmt.Sprintf("pinned-%d", time.Now().UnixNano())
	CreateFlow("pinned-policy-flow", suite.T())
	BeginTrail("pinned-policy-trail", "pinned-policy-flow", "", suite.T())
	// No retries: each retry of a 503 would publish a further version.
	_, _, _, _, publishErr := executeCommandC(fmt.Sprintf(
		"publish rego-policy %s --name 'Pinned policy' --path testdata/policies/allow-all.rego "+
			"--input-schema trail "+
			"--host %s --org %s --api-token %s --max-api-retries 0",
		suite.identifier, global.Host, global.Org, global.ApiToken))
	published, err := regopolicies.NewClient(kosliClient, global.Host, global.ApiToken, false).
		GetVersion(global.Org, suite.identifier, 1)
	require.NoError(suite.T(), err, "publish: %v", publishErr)
	require.Equal(suite.T(), "pending", published.Status)
}

func (suite *EvaluatePolicyPinnedTestSuite) evaluate(reference string) error {
	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"evaluate policy --context trail=pinned-policy-flow/pinned-policy-trail --identifier %s "+
			"--host %s --org %s --api-token %s --max-api-retries 0",
		reference, global.Host, global.Org, global.ApiToken))
	return err
}

func (suite *EvaluatePolicyPinnedTestSuite) TestAPendingPinnedVersionIsRefusedByTheServer() {
	err := suite.evaluate(suite.identifier + "@v1")

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), fmt.Sprintf(
		"Version 1 of rego policy '%s' is pending and cannot be evaluated", suite.identifier))
}

func (suite *EvaluatePolicyPinnedTestSuite) TestAVersionThePolicyDoesNotHaveIsRefusedByTheServer() {
	err := suite.evaluate(suite.identifier + "@v2")

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), fmt.Sprintf(
		"Rego policy '%s' has no version 2", suite.identifier))
}

func TestEvaluatePolicyPinnedTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluatePolicyPinnedTestSuite))
}
