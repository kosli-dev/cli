# METADATA
# title: Trail source-code-review compliance
# description: |
#   Every trail in input.trails is compliant if it was authored by a
#   service account, or if its PR has at least one approver.
package policy

import rego.v1

default allow := false

allow if {
	every trail in input.trails {
		trail_compliant(trail)
	}
}

# METADATA
# title: Trail has a compliant author or approval
# scope: document

# METADATA
# title: service-account author
# scope: rule
trail_compliant(trail) if {
	trail.author == "ci-bot"
}

# METADATA
# title: human commit with PR approval
# scope: rule
trail_compliant(trail) if {
	count(trail.pr.approvers) > 0
}
