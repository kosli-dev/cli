package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateShortDesc = `[BETA] Evaluate data against Rego policies.`

// Backtick breaks (`"` + "`x`" + `"`) are needed to embed markdown
// inline code spans inside raw string literals.
const evaluateLongDesc = evaluateShortDesc + `

This command is in BETA. Behaviour, flags, and the policy input shape may
change without notice. Pin a CLI version if you depend on it from CI.

Evaluate trail data or local JSON input against custom Rego policies.

Use ` + "`evaluate trail`" + ` or ` + "`evaluate trails`" + ` to fetch data from Kosli and evaluate it.
Use ` + "`evaluate input`" + ` to evaluate a local JSON file or stdin without any API calls.

The policy must use ` + "`package policy`" + ` and define an ` + "`allow`" + ` rule.
An optional ` + "`violations`" + ` rule (a set of strings) can provide human-readable denial reasons.

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
		newEvaluateTrailCmd(out),
		newEvaluateTrailsCmd(out),
		newEvaluateInputCmd(out),
	)

	return cmd
}
