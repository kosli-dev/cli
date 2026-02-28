package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

type EvaluateTrailsCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	trailName             string
	trailName2            string
}

func (suite *EvaluateTrailsCommandTestSuite) SetupTest() {
	suite.flowName = "evaluate-trails"
	suite.trailName = "test-trail-1"
	suite.trailName2 = "test-trail-2"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	BeginTrail(suite.trailName2, suite.flowName, "", suite.T())
}

func (suite *EvaluateTrailsCommandTestSuite) TestEvaluateTrailsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing trail name argument fails",
			cmd:       fmt.Sprintf(`evaluate trails --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
		{
			wantError: true,
			name:      "missing --flow flag fails",
			cmd:       fmt.Sprintf(`evaluate trails %s %s`, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\", \"policy\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --policy flag fails",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"policy\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --org orgX`, suite.trailName, suite.flowName),
		},
		{
			wantError: true,
			name:      "evaluating a non-existing trail fails",
			cmd:       fmt.Sprintf(`evaluate trails non-existent --flow %s --policy testdata/policies/allow-all.rego %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Trail with name 'non-existent' does not exist for Organization '%s' and Flow '%s'\n", global.Org, suite.flowName),
		},
		{
			name: "with --policy allow-all exits 0",
			cmd:  fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy deny-all exits 1",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/deny-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			name:       "with --output json and --policy prints allow/violations JSON",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}},
		},
		{
			name:   "with --output table and --policy prints ALLOWED text",
			cmd:    fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output table %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden: "Policy evaluation: ALLOWED\n",
		},
		{
			wantError: true,
			name:      "with --output table and deny --policy prints DENIED text with violations",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/deny-all.rego --output table %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Policy evaluation: DENIED\nViolations:\n  - always denied\nError: policy denied: [always denied]\n",
		},
		{
			wantError: true,
			name:      "with --output invalid returns error",
			cmd:       fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output invalid %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: unsupported output format: invalid\n",
		},
		{
			name:       "with --policy and --show-input --output json includes input in JSON output",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}, {"input.trails.[0].name", suite.trailName}},
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *EvaluateTrailsCommandTestSuite) TestEvaluateTrailsEnrichment() {
	trailName := "test-trails-enrichment"
	fingerprint := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	BeginTrail(trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "bar", suite.T())
	CreateArtifactOnTrail(suite.flowName, trailName, "cli", fingerprint, "file1", suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "foo", true, suite.T())

	tests := []cmdTestCase{
		{
			name:       "trails in output have map-keyed attestations_statuses",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trails.[0].compliance_status.attestations_statuses.bar.attestation_name", "bar"}},
		},
		{
			name: "rego policy can reference enriched attestation fields",
			cmd:  fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/check-trails-attestation-name.rego %s`, trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *EvaluateTrailsCommandTestSuite) TestEvaluateTrailsRehydration() {
	trailName := "test-trails-rehydration"
	fingerprint := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	BeginTrail(trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "trail-att", suite.T())
	CreateArtifactOnTrail(suite.flowName, trailName, "cli", fingerprint, "file1", suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "art-att", true, suite.T())

	tests := []cmdTestCase{
		{
			name:       "rehydrated fields present on trail-level attestations",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trails.[0].compliance_status.attestations_statuses.trail-att.html_url", "not-nil"}},
		},
		{
			name: "rego policy can reference rehydrated field",
			cmd:  fmt.Sprintf(`evaluate trails %s --flow %s --policy testdata/policies/check-trails-rehydrated-field.rego %s`, trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *EvaluateTrailsCommandTestSuite) TestEvaluateTrailsAttestationsFilter() {
	trailName := "test-trails-filter"
	fingerprint := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	BeginTrail(trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "trail-att", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "trail-other", suite.T())
	CreateArtifactOnTrail(suite.flowName, trailName, "cli", fingerprint, "file1", suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "art-att", true, suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "art-other", true, suite.T())

	tests := []cmdTestCase{
		{
			name:       "--attestations filters attestations across all trails",
			cmd:        fmt.Sprintf(`evaluate trails %s --flow %s --attestations trail-att --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trails.[0].compliance_status.attestations_statuses.trail-att.attestation_name", "trail-att"}},
		},
		{
			name: "rego policy referencing filtered-in attestation passes",
			cmd:  fmt.Sprintf(`evaluate trails %s --flow %s --attestations trail-att --policy testdata/policies/check-trails-filtered-attestation.rego %s`, trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestEvaluateTrailsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateTrailsCommandTestSuite))
}
