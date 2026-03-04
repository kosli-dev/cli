package policy

allow if {
	input.trail.compliance_status.attestations_statuses.bar.attestation_name == "bar"
}

violations contains msg if {
	not input.trail.compliance_status.attestations_statuses.bar
	msg := "trail attestation 'bar' not found"
}
