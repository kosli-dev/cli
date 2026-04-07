package policy

import (
	"fmt"
	"strings"
)

// FlowNameExpr returns a policy expression matching a single flow name.
func FlowNameExpr(name string) string {
	return fmt.Sprintf(`${{ flow.name == "%s" }}`, name)
}

// FlowNameInExpr returns a policy expression matching any of the given flow names.
// For a single name, it returns an equality expression instead.
func FlowNameInExpr(names []string) string {
	if len(names) == 1 {
		return FlowNameExpr(names[0])
	}
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = fmt.Sprintf(`"%s"`, n)
	}
	return fmt.Sprintf(`${{ flow.name in [%s] }}`, strings.Join(quoted, ", "))
}

// ArtifactNameMatchExpr returns a policy expression matching artifact names by regex.
func ArtifactNameMatchExpr(regex string) string {
	return fmt.Sprintf(`${{ matches(artifact.name, "%s") }}`, regex)
}

// ComparisonExpr returns a policy expression comparing a context field to a value.
func ComparisonExpr(context, op, value string) string {
	return fmt.Sprintf(`${{ %s %s "%s" }}`, context, op, value)
}

// CombineExprs joins inner expressions (without ${{ }} wrappers) with a logical operator.
func CombineExprs(op string, exprs ...string) string {
	if len(exprs) == 1 {
		return WrapExpr(exprs[0])
	}
	return fmt.Sprintf("${{ %s }}", strings.Join(exprs, " "+op+" "))
}

// WrapExpr adds the ${{ }} wrapper if not already present.
func WrapExpr(raw string) string {
	if strings.HasPrefix(raw, "${{") && strings.HasSuffix(raw, "}}") {
		return raw
	}
	return fmt.Sprintf("${{ %s }}", raw)
}
