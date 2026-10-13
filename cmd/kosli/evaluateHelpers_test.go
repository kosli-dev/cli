package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/kosli-dev/cli/internal/evaluations"
	"github.com/stretchr/testify/require"
)

func TestEvaluateResultKeysListsEveryKeyTheOutputCanHave(t *testing.T) {
	var out bytes.Buffer
	result := &evaluate.Result{Allow: true}
	cited := &evaluations.PolicyCitation{Identifier: "pr-approved", Version: 3, Digest: "sha256:abc123"}
	err := printEvaluateResult(&out, result, map[string]any{}, "json", true, map[string]any{}, false, "decision-id", cited)
	require.NoError(t, err)

	var printed map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &printed))
	for key := range printed {
		// Only evaluate policy prints the policy it ran, and it drops an output
		// of that name itself, so --output-rule policy stays open to evaluate trail.
		if key == "policy" {
			continue
		}
		require.Contains(t, evaluateResultKeys, key)
	}
	require.Contains(t, printed, "policy")
}
