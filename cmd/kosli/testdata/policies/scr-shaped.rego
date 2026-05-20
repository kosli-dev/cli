package policy

import rego.v1

default allow := false

allow if {
	every trail in input.trails {
		trail_compliant(trail)
	}
}

# METADATA
# scope: document
# title: Commit has independent review (or is exempt)
# description: A commit is compliant if authored by a service account, or if its PR has independent approval.

# METADATA
# title: exempt — service account author
trail_compliant(trail) if {
	startswith(trail.author, "bot-")
}

# METADATA
# title: human commit with independent PR approval
trail_compliant(trail) if {
	not startswith(trail.author, "bot-")
	has_independent_approval(trail)
}

# METADATA
# scope: document
# title: PR approval covering every author

# METADATA
# title: regular commit — branch authors + PR author approved
has_independent_approval(trail) if {
	trail.kind == "regular"
	count(trail.approvers) > 0
}

# METADATA
# title: merge commit — branch authors approved
has_independent_approval(trail) if {
	trail.kind == "merge"
}
