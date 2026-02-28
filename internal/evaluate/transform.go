package evaluate

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
	trailMap, ok := trailData.(map[string]interface{})
	if !ok {
		return nil
	}
	cs, ok := trailMap["compliance_status"].(map[string]interface{})
	if !ok {
		return nil
	}

	var ids []string
	if as, ok := cs["attestations_statuses"].(map[string]interface{}); ok {
		ids = append(ids, collectIDsFromAttestationMap(as)...)
	}
	if artifacts, ok := cs["artifacts_statuses"].(map[string]interface{}); ok {
		for _, artData := range artifacts {
			artMap, ok := artData.(map[string]interface{})
			if !ok {
				continue
			}
			if as, ok := artMap["attestations_statuses"].(map[string]interface{}); ok {
				ids = append(ids, collectIDsFromAttestationMap(as)...)
			}
		}
	}
	return ids
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
