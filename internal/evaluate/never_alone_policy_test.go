package evaluate

import (
	"context"
	"testing"

	"github.com/open-policy-agent/opa/v1/tester"
	"github.com/stretchr/testify/require"
)

// TestNeverAlonePolicy runs the Rego tests for the four-eyes policy this
// repository's CI evaluates, so they run wherever go test does.
func TestNeverAlonePolicy(t *testing.T) {
	results, err := tester.Run(context.Background(),
		"../../bin/never_alone/four-eyes-policy.rego",
		"../../bin/never_alone/four-eyes-policy_test.rego",
	)
	require.NoError(t, err)
	require.NotEmpty(t, results, "no Rego tests were found")
	for _, r := range results {
		require.Truef(t, r.Pass(), "%s failed (error: %v)", r.Name, r.Error)
	}
}
