package policy

allow if {
	input.trails[0].compliance_status.attestations_statuses["trail-att"].html_url
}

violations contains msg if {
	not input.trails[0].compliance_status.attestations_statuses["trail-att"].html_url
	msg := "trail-att missing html_url (rehydration failed)"
}
