package evaluate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransformTrail(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := TransformTrail(nil)
		require.Nil(t, result)
	})

	t.Run("non-map input passes through unchanged", func(t *testing.T) {
		input := "just a string"
		result := TransformTrail(input)
		require.Equal(t, input, result)
	})

	t.Run("trail with no compliance_status passes through", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "my-trail",
		}
		result := TransformTrail(input)
		resultMap := result.(map[string]interface{})
		require.Equal(t, "my-trail", resultMap["name"])
		require.Nil(t, resultMap["compliance_status"])
	})

	t.Run("empty attestations_statuses array becomes empty map", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": []interface{}{},
			},
		}
		result := TransformTrail(input)
		resultMap := result.(map[string]interface{})
		cs := resultMap["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"]
		require.IsType(t, map[string]interface{}{}, as)
		require.Empty(t, as)
	})

	t.Run("single trail-level attestation becomes map entry keyed by name", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": []interface{}{
					map[string]interface{}{
						"attestation_name": "bar",
						"is_compliant":     true,
					},
				},
			},
		}
		result := TransformTrail(input)
		resultMap := result.(map[string]interface{})
		cs := resultMap["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		bar := as["bar"].(map[string]interface{})
		require.Equal(t, "bar", bar["attestation_name"])
		require.Equal(t, true, bar["is_compliant"])
	})
}
