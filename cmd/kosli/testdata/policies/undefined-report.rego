package policy

allow = true

report := {"compliant": true} if {
	input.never_there
}
