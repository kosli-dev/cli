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

func TestFlowNameInExpr_Single(t *testing.T) {
	result := FlowNameInExpr([]string{"prod"})
	assert.Equal(t, `${{ flow.name == "prod" }}`, result)
}

func TestArtifactNameMatchExpr(t *testing.T) {
	result := ArtifactNameMatchExpr("^datadog:.*")
	assert.Equal(t, `${{ matches(artifact.name, "^datadog:.*") }}`, result)
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
