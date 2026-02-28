package policy

allow if {
	input.trails[0].compliance_status.attestations_statuses["trail-att"]
}

violations contains msg if {
	not input.trails[0].compliance_status.attestations_statuses["trail-att"]
	msg := "trail attestation 'trail-att' not found"
}
