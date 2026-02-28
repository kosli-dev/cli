package evaluate

import "strings"

// TransformTrail converts attestations_statuses arrays in trail data
// to maps keyed by attestation_name for easier Rego policy access.
func TransformTrail(trailData interface{}) interface{} {
	if trailData == nil {
		return nil
	}
	trailMap, ok := trailData.(map[string]interface{})
	if !ok {
		return trailData
	}

	cs, ok := trailMap["compliance_status"].(map[string]interface{})
	if !ok {
		return trailData
	}

	if arr, ok := cs["attestations_statuses"].([]interface{}); ok {
		cs["attestations_statuses"] = attestationsArrayToMap(arr)
	}

	if artifacts, ok := cs["artifacts_statuses"].(map[string]interface{}); ok {
		for _, artData := range artifacts {
			artMap, ok := artData.(map[string]interface{})
			if !ok {
				continue
			}
			if arr, ok := artMap["attestations_statuses"].([]interface{}); ok {
				artMap["attestations_statuses"] = attestationsArrayToMap(arr)
			}
		}
	}

	return trailData
}

// CollectAttestationIDs extracts all non-null attestation_id values
// from the already-transformed (map-keyed) trail data.
func CollectAttestationIDs(trailData interface{}) []string {
	var ids []string
	walkTrailAttestations(trailData, func(_ string, as map[string]interface{}) {
		ids = append(ids, collectIDsFromAttestationMap(as)...)
	})
	return ids
}

// RehydrateTrail merges attestation detail data into the already-transformed
// trail data. Fields from details are added where the key doesn't already exist.
func RehydrateTrail(trailData interface{}, details map[string]interface{}) interface{} {
	if len(details) == 0 {
		return trailData
	}
	walkTrailAttestations(trailData, func(_ string, as map[string]interface{}) {
		rehydrateAttestationMap(as, details)
	})
	return trailData
}

// FilterAttestations limits which attestations appear in the trail data.
// When filters is nil or empty, all attestations are included unchanged.
// Plain names (e.g. "pull-request") filter trail-level attestations.
// Dot-qualified names (e.g. "cli.unit-test") filter artifact-level attestations.
func FilterAttestations(trailData interface{}, filters []string) interface{} {
	if len(filters) == 0 {
		return trailData
	}

	trailFilters, artifactFilters := parseAttestationFilters(filters)

	walkTrailAttestations(trailData, func(artifactName string, as map[string]interface{}) {
		if artifactName == "" {
			filterMap(as, trailFilters)
		} else if allowed, exists := artifactFilters[artifactName]; exists {
			filterMap(as, allowed)
		} else {
			for k := range as {
				delete(as, k)
			}
		}
	})

	return trailData
}

// walkTrailAttestations navigates the trail data structure and calls fn
// for each attestations_statuses map found. The artifactName is "" for
// trail-level attestations, or the artifact name for artifact-level ones.
func walkTrailAttestations(trailData interface{}, fn func(artifactName string, as map[string]interface{})) {
	trailMap, ok := trailData.(map[string]interface{})
	if !ok {
		return
	}
	cs, ok := trailMap["compliance_status"].(map[string]interface{})
	if !ok {
		return
	}

	if as, ok := cs["attestations_statuses"].(map[string]interface{}); ok {
		fn("", as)
	}

	if artifacts, ok := cs["artifacts_statuses"].(map[string]interface{}); ok {
		for artName, artData := range artifacts {
			artMap, ok := artData.(map[string]interface{})
			if !ok {
				continue
			}
			if as, ok := artMap["attestations_statuses"].(map[string]interface{}); ok {
				fn(artName, as)
			}
		}
	}
}

func parseAttestationFilters(filters []string) (trailFilters map[string]bool, artifactFilters map[string]map[string]bool) {
	trailFilters = make(map[string]bool)
	artifactFilters = make(map[string]map[string]bool)
	for _, f := range filters {
		if dotIdx := strings.IndexByte(f, '.'); dotIdx >= 0 {
			artName := f[:dotIdx]
			attName := f[dotIdx+1:]
			if artifactFilters[artName] == nil {
				artifactFilters[artName] = make(map[string]bool)
			}
			artifactFilters[artName][attName] = true
		} else {
			trailFilters[f] = true
		}
	}
	return
}

func filterMap(m map[string]interface{}, keep map[string]bool) {
	for k := range m {
		if !keep[k] {
			delete(m, k)
		}
	}
}

func rehydrateAttestationMap(attestations map[string]interface{}, details map[string]interface{}) {
	for _, v := range attestations {
		entry, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		id, ok := entry["attestation_id"].(string)
		if !ok {
			continue
		}
		detail, ok := details[id].(map[string]interface{})
		if !ok {
			continue
		}
		for k, dv := range detail {
			if _, exists := entry[k]; !exists {
				entry[k] = dv
			}
		}
	}
}

func collectIDsFromAttestationMap(m map[string]interface{}) []string {
	var ids []string
	for _, v := range m {
		entry, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if id, ok := entry["attestation_id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func attestationsArrayToMap(arr []interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, entry := range arr {
		entryMap, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := entryMap["attestation_name"].(string)
		if !ok {
			continue
		}
		result[name] = entryMap
	}
	return result
}
