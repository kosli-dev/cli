package policy

allow if {
	input.trails[0].compliance_status.attestations_statuses.bar.attestation_name == "bar"
}

violations contains msg if {
	not input.trails[0].compliance_status.attestations_statuses.bar
	msg := "trail attestation 'bar' not found in first trail"
}
