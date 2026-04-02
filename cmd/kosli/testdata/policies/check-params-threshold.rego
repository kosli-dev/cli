package policy

import rego.v1

default allow := false

default threshold := 10

threshold := data.params.threshold if { data.params.threshold }

allow if { input.score >= threshold }

violations contains msg if {
	input.score < threshold
	msg := sprintf("score %d is below threshold %d", [input.score, threshold])
}
