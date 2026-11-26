package main

import (
	"io"

	"github.com/spf13/cobra"
)

// evaluateTrailDeprecation is shared by evaluate trail and evaluate trails,
// which evaluate client-side; evaluate policy evaluates the same trails in Kosli.
const evaluateTrailDeprecation = "use 'kosli evaluate policy --context trail=FLOW-NAME/TRAIL-NAME' instead, which evaluates the policy in Kosli"

// evaluateInputDeprecation names the tool that evaluates local JSON input
// against a Rego policy without Kosli.
const evaluateInputDeprecation = "use 'opa eval' directly instead"

const evaluateShortDesc = `Evaluate data against Rego policies.`

// Backtick breaks (`"` + "`x`" + `"`) are needed to embed markdown
// inline code spans inside raw string literals.
const evaluateLongDesc = evaluateShortDesc + `

Evaluate trail data or local JSON input against custom Rego policies.

Use ` + "`evaluate policy`" + ` to evaluate a policy against trails in Kosli.
` + "`evaluate trail`" + ` and ` + "`evaluate trails`" + ` are deprecated; use ` + "`evaluate policy`" + ` instead.
` + "`evaluate input`" + ` is deprecated; use ` + "`opa eval`" + ` directly to evaluate local JSON input.

The policy must use ` + "`package policy`" + ` and define an ` + "`allow`" + ` rule.
An optional ` + "`violations`" + ` rule (a set of strings) can provide human-readable denial reasons.
Use ` + "`--output-rule`" + ` to add other rules of the policy, like a report, to the JSON output.

By default a deny exits with code 1 so the command can gate a pipeline.
Pass ` + "`--no-assert`" + ` to use the command as a policy decision point: it prints
the verdict and exits 0 even on deny, leaving the asserting to a downstream
step. ` + "`--assert`" + ` is the current default; pass it explicitly to lock in the
assert-on-deny behaviour across future releases, where the default will flip
to ` + "`--no-assert`" + `.

Use ` + "`--params`" + ` to pass configuration data (thresholds, expected counts, etc.)
to your policy. Params are available as ` + "`data.params`" + ` in Rego, keeping policy
logic reusable across environments with different tolerances.`

func newEvaluateCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "evaluate",
		Short: evaluateShortDesc,
		Long:  evaluateLongDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newEvaluatePolicyCmd(out),
		newEvaluateTrailCmd(out),
		newEvaluateTrailsCmd(out),
		newEvaluateInputCmd(out),
	)

	return cmd
}
