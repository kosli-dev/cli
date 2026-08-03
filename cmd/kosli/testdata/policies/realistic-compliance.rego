package policy

# A policy shaped like one a user would actually write against trail data:
# default-deny, required attestations overridable via --params, and
# human-readable violation messages.

default allow := false

default required := ["pull-request", "unit-tests"]

required := data.params.required_attestations if { data.params.required_attestations }

attestations := input.trail.compliance_status.attestations_statuses

allow if {
	count(violations) == 0
}

violations contains msg if {
	some name in required
	not attestations[name]
	msg := sprintf("required attestation %q is missing", [name])
}

violations contains msg if {
	some name in required
	attestations[name].compliant == false
	msg := sprintf("attestation %q is not compliant", [name])
}
