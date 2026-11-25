package main

import (
	"bytes"
	"testing"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

// TestEveryOutputCommandDocumentsItsJSON fails when a command gains --output
// without saying what shapes its JSON, so the reference page cannot omit it.
func TestEveryOutputCommandDocumentsItsJSON(t *testing.T) {
	root, err := newRootCmd(&bytes.Buffer{}, &bytes.Buffer{}, nil)
	require.NoError(t, err)

	var walk func(cmd *cobra.Command)
	walk = func(cmd *cobra.Command) {
		for _, sub := range cmd.Commands() {
			walk(sub)
		}
		operations := docgen.SplitOperations(cmd.Annotations[docgen.JSONOutputAPIAnnotation])
		source, hasSource := cmd.Annotations[docgen.JSONOutputSourceAnnotation]
		documented := len(operations) > 0 || hasSource

		if cmd.LocalFlags().Lookup("output") == nil {
			require.Falsef(t, documented, "%s documents JSON output but has no --output flag", cmd.CommandPath())
			return
		}
		require.Truef(t, documented,
			"%s has --output: set the %s annotation to the API operation it prints, or %s to %q",
			cmd.CommandPath(), docgen.JSONOutputAPIAnnotation, docgen.JSONOutputSourceAnnotation, docgen.JSONOutputSourceCLI)

		switch source {
		case "":
			require.NotEmptyf(t, operations, "%s: no API operation", cmd.CommandPath())
		case docgen.JSONOutputSourceDerived:
			require.NotEmptyf(t, operations, "%s: derived JSON needs the API operations it derives from", cmd.CommandPath())
		case docgen.JSONOutputSourceCLI:
			require.Emptyf(t, operations, "%s: CLI-built JSON names no API operation", cmd.CommandPath())
		default:
			require.Failf(t, "unknown source", "%s: %s is %q", cmd.CommandPath(), docgen.JSONOutputSourceAnnotation, source)
		}
	}
	walk(root)
}
