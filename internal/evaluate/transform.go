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

	return trailData
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
