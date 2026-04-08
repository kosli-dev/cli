package policy

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlowNameExpr(t *testing.T) {
	result := FlowNameExpr("prod")
	assert.Equal(t, `${{ flow.name == "prod" }}`, result)
}

func TestFlowNameInExpr(t *testing.T) {
	result := FlowNameInExpr([]string{"runner", "saver"})
	assert.Equal(t, `${{ flow.name in ["runner", "saver"] }}`, result)
}

func TestFlowNameInExpr_Empty(t *testing.T) {
	assert.Equal(t, "", FlowNameInExpr(nil))
	assert.Equal(t, "", FlowNameInExpr([]string{}))
}

func TestFlowNameInExpr_Single(t *testing.T) {
	result := FlowNameInExpr([]string{"prod"})
	assert.Equal(t, `${{ flow.name == "prod" }}`, result)
}

func TestFlowTagExpr(t *testing.T) {
	result := FlowTagExpr("risk-level", "==", "high")
	assert.Equal(t, `${{ flow.tags.risk-level == "high" }}`, result)
}

func TestFlowTagExpr_DottedKey(t *testing.T) {
	result := FlowTagExpr("key.with.dots", "!=", "bad")
	assert.Equal(t, `${{ flow.tags.key.with.dots != "bad" }}`, result)
}

func TestArtifactNameMatchExpr(t *testing.T) {
	result := ArtifactNameMatchExpr("^datadog:.*")
	assert.Equal(t, `${{ matches(artifact.name, "^datadog:.*") }}`, result)
}

func TestMatchesExpr(t *testing.T) {
	result := MatchesExpr("flow.name", "^prod")
	assert.Equal(t, `${{ matches(flow.name, "^prod") }}`, result)
}

func TestExistsExpr(t *testing.T) {
	result := ExistsExpr("flow")
	assert.Equal(t, `${{ exists(flow) }}`, result)
}

func TestComparisonExpr(t *testing.T) {
	result := ComparisonExpr("flow.tags.risk", "==", "high")
	assert.Equal(t, `${{ flow.tags.risk == "high" }}`, result)
}

func TestCombineExprs(t *testing.T) {
	e1 := `flow.name == "prod"`
	e2 := `artifact.name == "svc"`
	result := CombineExprs("and", e1, e2)
	assert.Equal(t, `${{ flow.name == "prod" and artifact.name == "svc" }}`, result)
}

func TestCombineExprs_Single(t *testing.T) {
	result := CombineExprs("or", `flow.name == "prod"`)
	assert.Equal(t, `${{ flow.name == "prod" }}`, result)
}

func TestWrapExpr_AddsWrapper(t *testing.T) {
	result := WrapExpr(`flow.name == "prod"`)
	assert.Equal(t, `${{ flow.name == "prod" }}`, result)
}

func TestWrapExpr_Idempotent(t *testing.T) {
	result := WrapExpr(`${{ flow.name == "prod" }}`)
	assert.Equal(t, `${{ flow.name == "prod" }}`, result)
}

func TestUnwrapExpr(t *testing.T) {
	result := UnwrapExpr(`${{ flow.name == "prod" }}`)
	assert.Equal(t, `flow.name == "prod"`, result)
}

func TestUnwrapExpr_AlreadyRaw(t *testing.T) {
	result := UnwrapExpr(`flow.name == "prod"`)
	assert.Equal(t, `flow.name == "prod"`, result)
}

func TestNegateExpr(t *testing.T) {
	result := NegateExpr(`flow.name == "prod"`)
	assert.Equal(t, `not(flow.name == "prod")`, result)
}

func TestCombineAndNegate(t *testing.T) {
	a := UnwrapExpr(FlowNameExpr("prod"))
	b := NegateExpr(UnwrapExpr(ArtifactNameMatchExpr("^datadog:.*")))
	result := CombineExprs("and", a, b)
	assert.Equal(t, `${{ flow.name == "prod" and not(matches(artifact.name, "^datadog:.*")) }}`, result)
}
