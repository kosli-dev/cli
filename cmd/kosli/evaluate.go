package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateShortDesc = `Evaluate data against Rego policies.`

// Backtick breaks (`"` + "`x`" + `"`) are needed to embed markdown
// inline code spans inside raw string literals.
const evaluateLongDesc = evaluateShortDesc + `
Evaluate trail data or local JSON input against custom Rego policies.

Use ` + "`evaluate trail`" + ` or ` + "`evaluate trails`" + ` to fetch data from Kosli and evaluate it.
Use ` + "`evaluate input`" + ` to evaluate a local JSON file or stdin without any API calls.

The policy must use ` + "`package policy`" + ` and define an ` + "`allow`" + ` rule.
An optional ` + "`violations`" + ` rule (a set of strings) can provide human-readable denial reasons.
The command exits with code 0 when allowed and code 1 when denied.`

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
