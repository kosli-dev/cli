package policywizard

import (
	"testing"

	"github.com/kosli-dev/cli/internal/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// advanceStep tests
// ---------------------------------------------------------------------------

func TestAdvance_ProvConfirm_RequiredGoesToExceptions(t *testing.T) {
	m := newTestModel()
	m.step = stepProvConfirm
	m.requireProv = true

	m.advanceStep()

	assert.Equal(t, stepProvExcConfirm, m.step)
	assert.Equal(t, targetProvException, m.exprTarget)
}

func TestAdvance_ProvConfirm_NotRequiredGoesToTrail(t *testing.T) {
	m := newTestModel()
	m.step = stepProvConfirm
	m.requireProv = false

	m.advanceStep()

	assert.Equal(t, stepTrailConfirm, m.step)
}

func TestAdvance_TrailConfirm_RequiredGoesToExceptions(t *testing.T) {
	m := newTestModel()
	m.step = stepTrailConfirm
	m.requireTrail = true

	m.advanceStep()

	assert.Equal(t, stepTrailExcConfirm, m.step)
}

func TestAdvance_TrailConfirm_NotRequiredGoesToAtt(t *testing.T) {
	m := newTestModel()
	m.step = stepTrailConfirm
	m.requireTrail = false

	m.advanceStep()

	assert.Equal(t, stepAttConfirm, m.step)
}

func TestAdvance_ProvExcConfirm_YesGoesToExprMode(t *testing.T) {
	m := newTestModel()
	m.step = stepProvExcConfirm
	m.lastConfirm = true

	m.advanceStep()

	assert.Equal(t, stepExprMode, m.step)
	assert.Equal(t, targetProvException, m.exprTarget)
}

func TestAdvance_ProvExcConfirm_NoGoesToTrail(t *testing.T) {
	m := newTestModel()
	m.step = stepProvExcConfirm
	m.lastConfirm = false

	m.advanceStep()

	assert.Equal(t, stepTrailConfirm, m.step)
}

func TestAdvance_TrailExcConfirm_YesGoesToExprMode(t *testing.T) {
	m := newTestModel()
	m.step = stepTrailExcConfirm
	m.lastConfirm = true

	m.advanceStep()

	assert.Equal(t, stepExprMode, m.step)
	assert.Equal(t, targetTrailException, m.exprTarget)
}

func TestAdvance_TrailExcConfirm_NoGoesToAtt(t *testing.T) {
	m := newTestModel()
	m.step = stepTrailExcConfirm
	m.lastConfirm = false

	m.advanceStep()

	assert.Equal(t, stepAttConfirm, m.step)
}

func TestAdvance_AttConfirm_YesGoesToDetails(t *testing.T) {
	m := newTestModel()
	m.step = stepAttConfirm
	m.lastConfirm = true

	m.advanceStep()

	assert.Equal(t, stepAttDetails, m.step)
}

func TestAdvance_AttConfirm_NoGoesToSaveFile(t *testing.T) {
	m := newTestModel()
	m.step = stepAttConfirm
	m.lastConfirm = false

	m.advanceStep()

	assert.Equal(t, stepSaveFile, m.step)
}

func TestAdvance_AttCondConfirm_YesGoesToExprMode(t *testing.T) {
	m := newTestModel()
	m.step = stepAttCondConfirm
	m.lastConfirm = true

	m.advanceStep()

	assert.Equal(t, stepExprMode, m.step)
	assert.Equal(t, targetAttCondition, m.exprTarget)
}

func TestAdvance_AttCondConfirm_NoCommitsAndLoops(t *testing.T) {
	m := newTestModel()
	m.step = stepAttCondConfirm
	m.lastConfirm = false
	m.currentAttRule = policy.AttestationRule{Type: "snyk", Name: "scan"}
	m.Policy.Artifacts = &policy.ArtifactRules{}

	m.advanceStep()

	assert.Equal(t, stepAttConfirm, m.step)
	require.Len(t, m.Policy.Artifacts.Attestations, 1)
	assert.Equal(t, "snyk", m.Policy.Artifacts.Attestations[0].Type)
}

func TestAdvance_ExprMode_AllModes(t *testing.T) {
	tests := []struct {
		mode     string
		expected wizardStep
	}{
		{"flow_name", stepExprFlowName},
		{"flow_tag", stepExprFlowTag},
		{"artifact_name", stepExprArtifactName},
		{"custom", stepExprCustomCtx},
		{"raw", stepExprRaw},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			m := newTestModel()
			m.step = stepExprMode
			m.exprMode = tt.mode
			m.advanceStep()
			assert.Equal(t, tt.expected, m.step)
		})
	}
}

func TestAdvance_ExprCustomCtx_TagKeyGoesToTagKeyStep(t *testing.T) {
	m := newTestModel()
	m.step = stepExprCustomCtx
	m.exprContext = "flow.tags.<key>"

	m.advanceStep()

	assert.Equal(t, stepExprCustomTagKey, m.step)
}

func TestAdvance_ExprCustomCtx_DirectGoesToOp(t *testing.T) {
	m := newTestModel()
	m.step = stepExprCustomCtx
	m.exprContext = "flow.name"

	m.advanceStep()

	assert.Equal(t, stepExprCustomOp, m.step)
}

func TestAdvanceAfterExpr_AllTargets(t *testing.T) {
	tests := []struct {
		target   exprTarget
		expected wizardStep
	}{
		{targetProvException, stepProvExcConfirm},
		{targetTrailException, stepTrailExcConfirm},
		{targetAttCondition, stepAttConfirm},
	}
	for _, tt := range tests {
		m := newTestModel()
		m.exprTarget = tt.target
		m.advanceAfterExpr()
		assert.Equal(t, tt.expected, m.step)
	}
}

// ---------------------------------------------------------------------------
// applyFormValues tests
// ---------------------------------------------------------------------------

func TestApply_ProvConfirm_True_SetsProvenance(t *testing.T) {
	m := newTestModel()
	m.step = stepProvConfirm

	m.applyFormValues(formValues{confirm: true})

	assert.True(t, m.requireProv)
	require.NotNil(t, m.Policy.Artifacts)
	require.NotNil(t, m.Policy.Artifacts.Provenance)
	assert.True(t, m.Policy.Artifacts.Provenance.Required)
}

func TestApply_ProvConfirm_False_NoArtifacts(t *testing.T) {
	m := newTestModel()
	m.step = stepProvConfirm

	m.applyFormValues(formValues{confirm: false})

	assert.False(t, m.requireProv)
	assert.Nil(t, m.Policy.Artifacts)
}

func TestApply_TrailConfirm_True_SetsTrailCompliance(t *testing.T) {
	m := newTestModel()
	m.step = stepTrailConfirm

	m.applyFormValues(formValues{confirm: true})

	assert.True(t, m.requireTrail)
	require.NotNil(t, m.Policy.Artifacts)
	require.NotNil(t, m.Policy.Artifacts.TrailCompliance)
	assert.True(t, m.Policy.Artifacts.TrailCompliance.Required)
}

func TestApply_AttDetails_SetsCurrentRule(t *testing.T) {
	m := newTestModel()
	m.step = stepAttDetails

	m.applyFormValues(formValues{attType: "snyk", attName: "security-scan"})

	assert.Equal(t, "snyk", m.currentAttRule.Type)
	assert.Equal(t, "security-scan", m.currentAttRule.Name)
	assert.Empty(t, m.validationErr)
}

func TestApply_AttDetails_EmptyNameDefaultsToWildcard(t *testing.T) {
	m := newTestModel()
	m.step = stepAttDetails

	m.applyFormValues(formValues{attType: "snyk", attName: ""})

	assert.Equal(t, "*", m.currentAttRule.Name)
}

func TestApply_AttDetails_WildcardTypeAndName_Rejected(t *testing.T) {
	m := newTestModel()
	m.step = stepAttDetails

	m.applyFormValues(formValues{attType: "*", attName: "*"})

	assert.Contains(t, m.validationErr, "name must not be *")
	assert.Equal(t, policy.AttestationRule{}, m.currentAttRule)
}

func TestApply_AttDetails_WildcardTypeEmptyName_Rejected(t *testing.T) {
	m := newTestModel()
	m.step = stepAttDetails

	m.applyFormValues(formValues{attType: "*", attName: ""})

	assert.Contains(t, m.validationErr, "name must not be *")
}

func TestApply_ExprMode_StoresMode(t *testing.T) {
	m := newTestModel()
	m.step = stepExprMode

	m.applyFormValues(formValues{str: "flow_tag"})

	assert.Equal(t, "flow_tag", m.exprMode)
}

func TestApply_ExprFlowName_AppliesExpression(t *testing.T) {
	m := newTestModel()
	m.step = stepExprFlowName
	m.exprTarget = targetProvException
	m.Policy.Artifacts = &policy.ArtifactRules{
		Provenance: &policy.BooleanRule{Required: true},
	}

	m.applyFormValues(formValues{str: "prod"})

	require.Len(t, m.Policy.Artifacts.Provenance.Exceptions, 1)
	assert.Equal(t, `${{ flow.name == "prod" }}`, m.Policy.Artifacts.Provenance.Exceptions[0].If)
}

func TestApply_ExprFlowTag_StoresTagKey(t *testing.T) {
	m := newTestModel()
	m.step = stepExprFlowTag

	m.applyFormValues(formValues{str: "risk-level"})

	assert.Equal(t, "risk-level", m.exprTagKey)
}

func TestApply_ExprFlowTagOp_AppliesExpression(t *testing.T) {
	m := newTestModel()
	m.step = stepExprFlowTagOp
	m.exprTarget = targetTrailException
	m.exprTagKey = "team"
	m.Policy.Artifacts = &policy.ArtifactRules{
		TrailCompliance: &policy.BooleanRule{Required: true},
	}

	m.applyFormValues(formValues{operator: "==", str: "backend"})

	require.Len(t, m.Policy.Artifacts.TrailCompliance.Exceptions, 1)
	assert.Equal(t, `${{ flow.tags.team == "backend" }}`, m.Policy.Artifacts.TrailCompliance.Exceptions[0].If)
}

func TestApply_ExprArtifactName_AppliesExpression(t *testing.T) {
	m := newTestModel()
	m.step = stepExprArtifactName
	m.exprTarget = targetProvException
	m.Policy.Artifacts = &policy.ArtifactRules{
		Provenance: &policy.BooleanRule{Required: true},
	}

	m.applyFormValues(formValues{str: "^datadog:.*"})

	require.Len(t, m.Policy.Artifacts.Provenance.Exceptions, 1)
	assert.Equal(t, `${{ matches(artifact.name, "^datadog:.*") }}`, m.Policy.Artifacts.Provenance.Exceptions[0].If)
}

func TestApply_ExprRaw_AppliesWrappedExpression(t *testing.T) {
	m := newTestModel()
	m.step = stepExprRaw
	m.exprTarget = targetProvException
	m.Policy.Artifacts = &policy.ArtifactRules{
		Provenance: &policy.BooleanRule{Required: true},
	}

	m.applyFormValues(formValues{str: `flow.name == "prod"`})

	require.Len(t, m.Policy.Artifacts.Provenance.Exceptions, 1)
	assert.Equal(t, `${{ flow.name == "prod" }}`, m.Policy.Artifacts.Provenance.Exceptions[0].If)
}

func TestApply_ExprCustomCtx_StoresContext(t *testing.T) {
	m := newTestModel()
	m.step = stepExprCustomCtx

	m.applyFormValues(formValues{str: "artifact.name"})

	assert.Equal(t, "artifact.name", m.exprContext)
}

func TestApply_ExprCustomTagKey_BuildsContext(t *testing.T) {
	m := newTestModel()
	m.step = stepExprCustomTagKey

	m.applyFormValues(formValues{str: "risk-level"})

	assert.Equal(t, "flow.tags.risk-level", m.exprContext)
}

func TestApply_ExprCustomOp_AppliesExpression(t *testing.T) {
	m := newTestModel()
	m.step = stepExprCustomOp
	m.exprTarget = targetProvException
	m.exprContext = "artifact.name"
	m.Policy.Artifacts = &policy.ArtifactRules{
		Provenance: &policy.BooleanRule{Required: true},
	}

	m.applyFormValues(formValues{operator: "==", str: "myapp"})

	require.Len(t, m.Policy.Artifacts.Provenance.Exceptions, 1)
	assert.Equal(t, `${{ artifact.name == "myapp" }}`, m.Policy.Artifacts.Provenance.Exceptions[0].If)
}

func TestApply_AttCondition_CommitsAttestation(t *testing.T) {
	m := newTestModel()
	m.step = stepExprFlowName
	m.exprTarget = targetAttCondition
	m.currentAttRule = policy.AttestationRule{Type: "snyk", Name: "scan"}
	m.Policy.Artifacts = &policy.ArtifactRules{}

	m.applyFormValues(formValues{str: "prod"})

	require.Len(t, m.Policy.Artifacts.Attestations, 1)
	assert.Equal(t, "snyk", m.Policy.Artifacts.Attestations[0].Type)
	assert.Equal(t, `${{ flow.name == "prod" }}`, m.Policy.Artifacts.Attestations[0].If)
}

func TestApply_ExprCustomOp_MatchesUsesFunction(t *testing.T) {
	m := newTestModel()
	m.step = stepExprCustomOp
	m.exprTarget = targetProvException
	m.exprContext = "flow.name"
	m.Policy.Artifacts = &policy.ArtifactRules{
		Provenance: &policy.BooleanRule{Required: true},
	}

	m.applyFormValues(formValues{operator: "matches", str: "^prod"})

	require.Len(t, m.Policy.Artifacts.Provenance.Exceptions, 1)
	assert.Equal(t, `${{ matches(flow.name, "^prod") }}`, m.Policy.Artifacts.Provenance.Exceptions[0].If)
}

func TestApply_SaveFile_SetsOutputFile(t *testing.T) {
	m := newTestModel()
	m.step = stepSaveFile

	m.applyFormValues(formValues{str: "my-policy.yaml"})

	assert.Equal(t, "my-policy.yaml", m.OutputFile)
}

func TestAdvance_SaveFile_GoesToDone(t *testing.T) {
	m := newTestModel()
	m.step = stepSaveFile

	m.advanceStep()

	assert.Equal(t, stepDone, m.step)
}

func TestApply_StoresLastConfirm(t *testing.T) {
	m := newTestModel()
	m.step = stepProvConfirm

	m.applyFormValues(formValues{confirm: true})

	assert.True(t, m.lastConfirm)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newTestModel() *Model {
	m := NewModel(&Context{})
	return &m
}
