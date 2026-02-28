package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type EvaluateTrailCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	flowName              string
	trailName             string
}

func (suite *EvaluateTrailCommandTestSuite) SetupTest() {
	suite.flowName = "evaluate-trail"
	suite.trailName = "test-trail-1"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user-shared",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

func (suite *EvaluateTrailCommandTestSuite) TestEvaluateTrailCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "missing trail name argument fails",
			cmd:       fmt.Sprintf(`evaluate trail --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "providing more than one argument fails",
			cmd:       fmt.Sprintf(`evaluate trail %s xxx --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 2\n",
		},
		{
			wantError: true,
			name:      "missing --flow flag fails",
			cmd:       fmt.Sprintf(`evaluate trail %s %s`, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\", \"policy\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --policy flag fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"policy\" not set\n",
		},
		{
			wantError: true,
			name:      "missing --api-token fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --org orgX`, suite.trailName, suite.flowName),
		},
		{
			wantError: true,
			name:      "evaluating a non-existing trail fails",
			cmd:       fmt.Sprintf(`evaluate trail non-existent --flow %s --policy testdata/policies/allow-all.rego %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    fmt.Sprintf("Error: Trail with name 'non-existent' does not exist for Organization '%s' and Flow '%s'\n", global.Org, suite.flowName),
		},
		{
			name: "with --policy allow-all exits 0",
			cmd:  fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy deny-all exits 1",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy non-existent file fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/does-not-exist.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			wantError: true,
			name:      "with --policy invalid rego fails",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/invalid.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
		},
		{
			name:       "with --policy allow-all --output json prints JSON with allow true",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}},
		},
		{
			wantError:   true,
			name:        "with --policy deny-all --output json prints JSON with allow false and violations",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `(?s)"allow":\s*false.*"violations":\s*\[.*"always denied"`,
		},
		{
			name:        "with --policy allow-all --output table prints allowed text",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output table %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
		{
			wantError:   true,
			name:        "with --policy deny-all --output table prints denied text with violations",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego --output table %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `RESULT:\s+DENIED\nVIOLATIONS:\s+always denied`,
		},
		{
			name:        "with --policy allow-all and no --output defaults to table output",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
		{
			wantError: true,
			name:      "with --output invalid returns an error",
			cmd:       fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output invalid %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: unsupported output format: invalid\n",
		},
		{
			name:       "with --policy allow-all --output json --show-input includes input in JSON",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"allow", true}, {"input.trail.name", suite.trailName}},
		},
		{
			wantError:   true,
			name:        "with --policy deny-all --output json --show-input includes input alongside allow and violations",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/deny-all.rego --output json --show-input %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `(?s)"allow":\s*false.*"input":\s*\{.*"trail".*"violations":\s*\[.*"always denied"`,
		},
		{
			name:        "with --policy allow-all --output table --show-input ignores show-input",
			cmd:         fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output table --show-input %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: `RESULT:\s+ALLOWED`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *EvaluateTrailCommandTestSuite) TestEvaluateTrailEnrichment() {
	trailName := "test-trail-with-attestations"
	fingerprint := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	BeginTrail(trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "bar", suite.T())
	CreateArtifactOnTrail(suite.flowName, trailName, "cli", fingerprint, "file1", suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "foo", true, suite.T())

	tests := []cmdTestCase{
		{
			name:       "output has map-keyed trail-level attestations_statuses",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trail.compliance_status.attestations_statuses.bar.attestation_name", "bar"}},
		},
		{
			name:       "output has map-keyed artifact-level attestations_statuses",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trail.compliance_status.artifacts_statuses.cli.attestations_statuses.foo.attestation_name", "foo"}},
		},
		{
			name: "rego policy can reference attestation by name",
			cmd:  fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/check-attestation-name.rego %s`, trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *EvaluateTrailCommandTestSuite) TestEvaluateTrailRehydration() {
	trailName := "test-trail-rehydration"
	fingerprint := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	BeginTrail(trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "trail-att", suite.T())
	CreateArtifactOnTrail(suite.flowName, trailName, "cli", fingerprint, "file1", suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "art-att", true, suite.T())

	tests := []cmdTestCase{
		{
			name:       "rehydrated trail-level attestation has html_url from detail",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trail.compliance_status.attestations_statuses.trail-att.html_url", "not-nil"}},
		},
		{
			name:       "rehydrated artifact-level attestation has html_url from detail",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trail.compliance_status.artifacts_statuses.cli.attestations_statuses.art-att.html_url", "not-nil"}},
		},
		{
			name: "rego policy can reference rehydrated field",
			cmd:  fmt.Sprintf(`evaluate trail %s --flow %s --policy testdata/policies/check-rehydrated-field.rego %s`, trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *EvaluateTrailCommandTestSuite) TestEvaluateTrailAttestationsFilter() {
	trailName := "test-trail-filter"
	fingerprint := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	BeginTrail(trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "trail-att", suite.T())
	CreateGenericTrailAttestation(suite.flowName, trailName, "trail-other", suite.T())
	CreateArtifactOnTrail(suite.flowName, trailName, "cli", fingerprint, "file1", suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "art-att", true, suite.T())
	CreateGenericArtifactAttestation(suite.flowName, trailName, fingerprint, "art-other", true, suite.T())

	tests := []cmdTestCase{
		{
			name:       "--attestations trail-att keeps only trail-att in trail-level",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --attestations trail-att --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trail.compliance_status.attestations_statuses.trail-att.attestation_name", "trail-att"}},
		},
		{
			name:       "--attestations cli.art-att keeps only art-att in cli's attestations",
			cmd:        fmt.Sprintf(`evaluate trail %s --flow %s --attestations cli.art-att --policy testdata/policies/allow-all.rego --output json --show-input %s`, trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{{"input.trail.compliance_status.artifacts_statuses.cli.attestations_statuses.art-att.attestation_name", "art-att"}},
		},
		{
			name: "rego policy referencing filtered-in attestation passes",
			cmd:  fmt.Sprintf(`evaluate trail %s --flow %s --attestations trail-att --policy testdata/policies/check-filtered-attestation.rego %s`, trailName, suite.flowName, suite.defaultKosliArguments),
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestEvaluateTrailCommandTestSuite(t *testing.T) {
	suite.Run(t, new(EvaluateTrailCommandTestSuite))
}
