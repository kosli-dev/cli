package main

import (
	"io"

	"github.com/spf13/cobra"
)

const evaluateShortDesc = `Evaluate Kosli trail data against OPA/Rego policies.`

// Backtick breaks (`"` + "`x`" + `"`) are needed to embed markdown
// inline code spans inside raw string literals.
const evaluateLongDesc = evaluateShortDesc + `
Fetch trail data from Kosli and evaluate it against custom policies written
in Rego, the policy language used by Open Policy Agent (OPA).
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
	)

	return cmd
}
