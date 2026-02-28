package policy

allow if {
	input.trail.compliance_status.attestations_statuses["trail-att"].html_url
}

violations contains msg if {
	not input.trail.compliance_status.attestations_statuses["trail-att"].html_url
	msg := "trail-att missing html_url (rehydration failed)"
}
