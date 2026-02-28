package evaluate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransformTrail(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := TransformTrail(nil)
		assert.Nil(t, result)
	})

	t.Run("non-map input passes through unchanged", func(t *testing.T) {
		input := "just a string"
		result := TransformTrail(input)
		assert.Equal(t, input, result)
	})

	t.Run("trail with no compliance_status passes through", func(t *testing.T) {
		input := map[string]interface{}{
			"name": "my-trail",
		}
		result := TransformTrail(input)
		resultMap := result.(map[string]interface{})
		assert.Equal(t, "my-trail", resultMap["name"])
		assert.Nil(t, resultMap["compliance_status"])
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
		assert.Empty(t, as)
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
		assert.Equal(t, "bar", bar["attestation_name"])
		assert.Equal(t, true, bar["is_compliant"])
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
		assert.Len(t, as, 3)
		assert.Contains(t, as, "alpha")
		assert.Contains(t, as, "beta")
		assert.Contains(t, as, "gamma")
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
		assert.Equal(t, "foo", foo["attestation_name"])
		assert.Equal(t, true, foo["is_compliant"])
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
		assert.Contains(t, trailAs, "trail-att")
		art1 := cs["artifacts_statuses"].(map[string]interface{})["art1"].(map[string]interface{})
		artAs := art1["attestations_statuses"].(map[string]interface{})
		assert.Contains(t, artAs, "art-att")
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
		assert.Contains(t, art1As, "a1")
		art2As := arts["art2"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		assert.Contains(t, art2As, "a2")
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
		assert.Len(t, as, 1)
		assert.Contains(t, as, "good")
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

func TestFilterAttestations(t *testing.T) {
	t.Run("nil filters returns trail unchanged", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
					},
				},
			},
		}
		result := FilterAttestations(input, nil)
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		assert.Contains(t, as, "bar")
	})

	t.Run("empty filters slice returns trail unchanged", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
					},
				},
			},
		}
		result := FilterAttestations(input, []string{})
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		assert.Contains(t, as, "bar")
	})

	t.Run("plain name keeps only matching trail-level attestation", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
					},
					"baz": map[string]interface{}{
						"attestation_name": "baz",
					},
				},
			},
		}
		result := FilterAttestations(input, []string{"bar"})
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		assert.Contains(t, as, "bar")
		assert.NotContains(t, as, "baz")
	})

	t.Run("plain name removes non-matching trail-level attestations", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"alpha": map[string]interface{}{"attestation_name": "alpha"},
					"beta":  map[string]interface{}{"attestation_name": "beta"},
					"gamma": map[string]interface{}{"attestation_name": "gamma"},
				},
			},
		}
		result := FilterAttestations(input, []string{"beta"})
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		as := cs["attestations_statuses"].(map[string]interface{})
		assert.Len(t, as, 1)
		assert.Contains(t, as, "beta")
	})

	t.Run("dot-qualified name keeps only matching artifact-level attestation", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"foo": map[string]interface{}{"attestation_name": "foo"},
							"bar": map[string]interface{}{"attestation_name": "bar"},
						},
					},
				},
			},
		}
		result := FilterAttestations(input, []string{"cli.foo"})
		cli := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["artifacts_statuses"].(map[string]interface{})["cli"].(map[string]interface{})
		as := cli["attestations_statuses"].(map[string]interface{})
		assert.Len(t, as, 1)
		assert.Contains(t, as, "foo")
	})

	t.Run("dot-qualified name: unmentioned artifact gets empty attestations_statuses", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"foo": map[string]interface{}{"attestation_name": "foo"},
						},
					},
					"server": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"bar": map[string]interface{}{"attestation_name": "bar"},
						},
					},
				},
			},
		}
		result := FilterAttestations(input, []string{"cli.foo"})
		arts := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["artifacts_statuses"].(map[string]interface{})
		// cli keeps foo
		cliAs := arts["cli"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		assert.Len(t, cliAs, 1)
		assert.Contains(t, cliAs, "foo")
		// server gets empty attestations_statuses
		serverAs := arts["server"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		assert.Empty(t, serverAs)
	})

	t.Run("mixed filters: trail-level and artifact-level both applied", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"trail-att":   map[string]interface{}{"attestation_name": "trail-att"},
					"trail-other": map[string]interface{}{"attestation_name": "trail-other"},
				},
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"art-att":   map[string]interface{}{"attestation_name": "art-att"},
							"art-other": map[string]interface{}{"attestation_name": "art-other"},
						},
					},
				},
			},
		}
		result := FilterAttestations(input, []string{"trail-att", "cli.art-att"})
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		trailAs := cs["attestations_statuses"].(map[string]interface{})
		assert.Len(t, trailAs, 1)
		assert.Contains(t, trailAs, "trail-att")
		cliAs := cs["artifacts_statuses"].(map[string]interface{})["cli"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		assert.Len(t, cliAs, 1)
		assert.Contains(t, cliAs, "art-att")
	})

	t.Run("filters with no matches leave all attestations_statuses empty", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{"attestation_name": "bar"},
				},
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"foo": map[string]interface{}{"attestation_name": "foo"},
						},
					},
				},
			},
		}
		result := FilterAttestations(input, []string{"nonexistent"})
		cs := result.(map[string]interface{})["compliance_status"].(map[string]interface{})
		trailAs := cs["attestations_statuses"].(map[string]interface{})
		assert.Empty(t, trailAs)
		cliAs := cs["artifacts_statuses"].(map[string]interface{})["cli"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})
		assert.Empty(t, cliAs)
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

	t.Run("does not overwrite existing fields", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"attestations_statuses": map[string]interface{}{
					"bar": map[string]interface{}{
						"attestation_name": "bar",
						"attestation_id":   "att-uuid-001",
						"status":           "COMPLETE",
					},
				},
			},
		}
		details := map[string]interface{}{
			"att-uuid-001": map[string]interface{}{
				"status":     "DIFFERENT",
				"origin_url": "https://example.com",
			},
		}
		result := RehydrateTrail(input, details)
		bar := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})["bar"].(map[string]interface{})
		assert.Equal(t, "COMPLETE", bar["status"], "existing field should not be overwritten")
		assert.Equal(t, "https://example.com", bar["origin_url"], "new field should be added")
	})

	t.Run("merges detail fields into artifact-level attestation", func(t *testing.T) {
		input := map[string]interface{}{
			"compliance_status": map[string]interface{}{
				"artifacts_statuses": map[string]interface{}{
					"cli": map[string]interface{}{
						"attestations_statuses": map[string]interface{}{
							"foo": map[string]interface{}{
								"attestation_name": "foo",
								"attestation_id":   "att-uuid-002",
								"is_compliant":     true,
							},
						},
					},
				},
			},
		}
		details := map[string]interface{}{
			"att-uuid-002": map[string]interface{}{
				"origin_url": "https://example.com/artifact",
			},
		}
		result := RehydrateTrail(input, details)
		foo := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["artifacts_statuses"].(map[string]interface{})["cli"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})["foo"].(map[string]interface{})
		assert.Equal(t, "https://example.com/artifact", foo["origin_url"])
		assert.Equal(t, true, foo["is_compliant"])
	})

	t.Run("attestation with no matching detail is left unchanged", func(t *testing.T) {
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
			"att-uuid-999": map[string]interface{}{
				"origin_url": "https://example.com",
			},
		}
		result := RehydrateTrail(input, details)
		bar := result.(map[string]interface{})["compliance_status"].(map[string]interface{})["attestations_statuses"].(map[string]interface{})["bar"].(map[string]interface{})
		assert.Equal(t, "bar", bar["attestation_name"])
		assert.Equal(t, true, bar["is_compliant"])
		assert.Nil(t, bar["origin_url"], "no matching detail means no new fields")
	})
}
