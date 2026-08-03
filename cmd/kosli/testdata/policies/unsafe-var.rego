package policy

# A policy with an unsafe variable: `y` is never bound by any expression.
# OPA rejects this at compile time, so `kosli evaluate` must surface an
# actionable error (with policy.rego:LINE) rather than a verdict.
default allow := false

allow if {
	x := y
	x == 7
}
