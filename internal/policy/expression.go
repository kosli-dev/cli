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
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return FlowNameExpr(names[0])
	}
	quoted := make([]string, len(names))
	for i, n := range names {
		quoted[i] = fmt.Sprintf(`"%s"`, n)
	}
	return fmt.Sprintf(`${{ flow.name in [%s] }}`, strings.Join(quoted, ", "))
}

// FlowTagExpr returns a policy expression comparing a flow tag to a value.
func FlowTagExpr(key, op, value string) string {
	return fmt.Sprintf(`${{ flow.tags.%s %s "%s" }}`, key, op, value)
}

// ArtifactNameMatchExpr returns a policy expression matching artifact names by regex.
func ArtifactNameMatchExpr(regex string) string {
	return MatchesExpr("artifact.name", regex)
}

// MatchesExpr returns a policy expression using the matches() function form.
func MatchesExpr(context, regex string) string {
	return fmt.Sprintf(`${{ matches(%s, "%s") }}`, context, regex)
}

// ExistsExpr returns a policy expression checking that a context field is not null.
func ExistsExpr(context string) string {
	return fmt.Sprintf(`${{ exists(%s) }}`, context)
}

// ComparisonExpr returns a policy expression comparing a context field to a value.
// The value is always quoted as a string. For operators like > or <, the policy
// engine must handle string-to-numeric coercion if needed.
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

// UnwrapExpr strips the ${{ }} wrapper, returning the inner expression.
// Tolerates varying whitespace inside the delimiters.
func UnwrapExpr(expr string) string {
	s := strings.TrimSpace(expr)
	s = strings.TrimPrefix(s, "${{")
	s = strings.TrimSuffix(s, "}}")
	return strings.TrimSpace(s)
}

// NegateExpr prefixes a raw (unwrapped) expression with the not operator.
func NegateExpr(raw string) string {
	return fmt.Sprintf("not %s", raw)
}
