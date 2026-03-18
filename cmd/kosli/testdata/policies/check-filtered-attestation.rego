package policy

allow if {
	input.trail.compliance_status.attestations_statuses["trail-att"]
}

violations contains msg if {
	not input.trail.compliance_status.attestations_statuses["trail-att"]
	msg := "trail attestation 'trail-att' not found"
}
