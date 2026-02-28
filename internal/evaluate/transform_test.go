package evaluate

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

	t.Run("multiple trail-level attestations all present in map", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": []interface{}{
					map[string]interface{}{"attestation_name": "alpha"},
					map[string]interface{}{"attestation_name": "beta"},
					map[string]interface{}{"attestation_name": "gamma"},
				},
			},
		}
		result := TransformTrail(input)
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		require.Len(t, as, 3)
		require.Contains(t, as, "alpha")
		require.Contains(t, as, "beta")
		require.Contains(t, as, "gamma")
	})

	t.Run("artifact-level attestations_statuses array becomes map", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": []interface{}{
							map[string]interface{}{"attestation_name": "foo", "is_compliant": true},
						},
					},
				},
			},
		}
		result := TransformTrail(input)
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		cli := cs["artifacts_statuses"].(map[string]interface{})["cli"].(map[string]interface{})
		as := cli["attestations_statuses"].(map[string]interface{})
		foo := as["foo"].(map[string]interface{})
		require.Equal(t, "foo", foo["attestation_name"])
		require.Equal(t, true, foo["is_compliant"])
	})

	t.Run("both trail-level and artifact-level transform in one call", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": []interface{}{
					map[string]interface{}{"attestation_name": "trail-att"},
				},
				"artifacts_statuses": map[string]interface{}{
					"art1": map[string]interface{}{
						"attestations_statuses": []interface{}{
							map[string]interface{}{"attestation_name": "art-att"},
						},
					},
				},
			},
		}
		result := TransformTrail(input)
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		trailAs := cs["attestations_statuses"].(map[string]interface{})
		require.Contains(t, trailAs, "trail-att")
		art1 := cs["artifacts_statuses"].(map[string]interface{})["art1"].(map[string]interface{})
		artAs := art1["attestations_statuses"].(map[string]interface{})
		require.Contains(t, artAs, "art-att")
	})

	t.Run("multiple artifacts each get their attestations transformed", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"artifacts_statuses": map[string]interface{}{
					"art1": map[string]interface{}{
						"attestations_statuses": []interface{}{
							map[string]interface{}{"attestation_name": "a1"},
						},
					},
					"art2": map[string]interface{}{
						"attestations_statuses": []interface{}{
							map[string]interface{}{"attestation_name": "a2"},
						},
					},
				},
			},
		}
		result := TransformTrail(input)
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		arts := cs["artifacts_statuses"].(map[string]interface{})
		art1As := arts["art1"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		require.Contains(t, art1As, "a1")
		art2As := arts["art2"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		require.Contains(t, art2As, "a2")
	})

	t.Run("entry without attestation_name is skipped", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": []interface{}{
					map[string]interface{}{"attestation_name": "good"},
					map[string]interface{}{"something_else": "no name"},
				},
			},
		}
		result := TransformTrail(input)
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		require.Len(t, as, 1)
		require.Contains(t, as, "good")
	})
}

func TestCollectAttestationIDs(t *testing.T) {
	t.Run("nil input returns empty slice", func(t *testing.T) {
		ids := CollectAttestationIDs(nil)
		assert.Empty(t, ids)
	})

	t.Run("trail with no compliance_status returns empty slice", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "my-trail",
		}
		ids := CollectAttestationIDs(input)
		assert.Empty(t, ids)
	})

	t.Run("collects ID from trail-level attestation", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
						"attestation_id":   "att-uuid-001",
					},
				},
			},
		}
		ids := CollectAttestationIDs(input)
		assert.Equal(t, []string{"att-uuid-001"}, ids)
	})

	t.Run("skips entries with null or missing attestation_id", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"has-id": map[string]interface{}{
						"attestation_name": "has-id",
						"attestation_id":   "att-uuid-001",
					},
					"null-id": map[string]interface{}{
						"attestation_name": "null-id",
						"attestation_id":   nil,
					},
					"missing-id": map[string]interface{}{
						"attestation_name": "missing-id",
					},
				},
			},
		}
		ids := CollectAttestationIDs(input)
		assert.Equal(t, []string{"att-uuid-001"}, ids)
	})

	t.Run("collects IDs from artifact-level attestation", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"foo": map[string]interface{}{
								"attestation_name": "foo",
								"attestation_id":   "att-uuid-002",
							},
						},
					},
				},
			},
		}
		ids := CollectAttestationIDs(input)
		assert.Equal(t, []string{"att-uuid-002"}, ids)
	})

	t.Run("collects from both trail-level and artifact-level", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"trail-att": map[string]interface{}{
						"attestation_name": "trail-att",
						"attestation_id":   "att-trail-001",
					},
				},
				"artifacts_statuses": map[string]interface{}{
					"art1": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"art-att": map[string]interface{}{
								"attestation_name": "art-att",
								"attestation_id":   "att-art-001",
							},
						},
					},
				},
			},
		}
		ids := CollectAttestationIDs(input)
		assert.Len(t, ids, 2)
		assert.Contains(t, ids, "att-trail-001")
		assert.Contains(t, ids, "att-art-001")
	})
}

func TestRehydrateTrail(t *testing.T) {
	t.Run("nil details map leaves trail unchanged", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
						"attestation_id":   "att-uuid-001",
					},
				},
			},
		}
		result := RehydrateTrail(input, nil)
		resultMap := result.(map[string]interface{})
		cs := resultMap["compliance_status"].(map[string]interface{})
		bar := cs["attestations_statuses"].(map[string]interface{})["bar"].(map[string]interface{})
		assert.Equal(t, "bar", bar["attestation_name"])
		assert.Equal(t, "att-uuid-001", bar["attestation_id"])
		assert.Nil(t, bar["origin_url"])
	})

	t.Run("empty details map leaves trail unchanged", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
						"attestation_id":   "att-uuid-001",
					},
				},
			},
		}
		result := RehydrateTrail(input, map[string]interface{}{})
		bar := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})["bar"].(map[string]interface{})
		assert.Equal(t, "bar", bar["attestation_name"])
		assert.Nil(t, bar["origin_url"])
	})

	t.Run("merges detail fields into trail-level attestation", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
						"attestation_id":   "att-uuid-001",
						"is_compliant":     true,
					},
				},
			},
		}
		details := map[string]interface{}{
			"att-uuid-001": map[string]interface{}{
				"origin_url": "https://github.com/org/repo/pull/42",
				"user_data":  map[string]interface{}{"key": "value"},
			},
		}
		result := RehydrateTrail(input, details)
		bar := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})["bar"].(map[string]interface{})
		assert.Equal(t, "https://github.com/org/repo/pull/42", bar["origin_url"])
		assert.Equal(t, map[string]interface{}{"key": "value"}, bar["user_data"])
		assert.Equal(t, true, bar["is_compliant"])
	})
}
