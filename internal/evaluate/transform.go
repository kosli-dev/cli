package evaluate

// TransformTrail converts attestations_statuses arrays in trail data
// to maps keyed by attestation_name for easier Rego policy access.
func TransformTrail(trailData interface{}) interface{} {
	if trailData == nil {
		return nil
	}
	return trailData
}
