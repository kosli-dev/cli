package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/kosli-dev/cli/internal/evaluate"
	"github.com/stretchr/testify/require"
)

func TestEvaluateResultKeysListsEveryKeyTheOutputCanHave(t *testing.T) {
	var out bytes.Buffer
	result := &evaluate.Result{Allow: true}
	err := printEvaluateResult(&out, result, map[string]any{}, "json", true, map[string]any{}, false, "decision-id")
	require.NoError(t, err)

	var printed map[string]any
	require.NoError(t, json.Unmarshal(out.Bytes(), &printed))
	for key := range printed {
		require.Contains(t, evaluateResultKeys, key)
	}
}
