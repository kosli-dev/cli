package policy

import rego.v1

# Four-eyes principle enforcement: every commit must have independent review.
# This policy evaluates per-commit attestation data from Kosli.
#
# Positive-assertion model: allow is true only when input.trails is a non-empty
# array AND every trail explicitly satisfies trail_compliant. Any failure to
# evaluate (malformed input, helper bug, missing field) leaves trails outside
# the compliant set and allow stays false. There is no defensive guard rule
# because the structure is fail-closed by construction.
default allow := false

allow if {
	is_array(input.trails)
	count(input.trails) > 0
	every trail in input.trails {
		trail_compliant(trail)
	}
}

# Set PR attestation name
attestation_name := name if {
	name := data.params.attestation_name
	is_string(name)
} else := "pr-review"

# ---------------------------------------------------------------------------
# Compliance
# ---------------------------------------------------------------------------

# A commit is compliant when an associated PR has independent approval
# covering every author after the latest code commit. There is no exemption
# based on the git author string: it is user-controlled, so matching on it
# would let anyone skip review.
trail_compliant(trail) if {
	attest := pr_attest(trail)
	some pr in attest.pull_requests
	all_authors_resolved(pr)
	has_independent_approval(trail, pr)
}

# ---------------------------------------------------------------------------
# Attestation data
#
# Used with `kosli evaluate trails` (plural). Each trail in input.trails
# represents one commit. The PR attestation payload is at:
#   trail.compliance_status.attestations_statuses[attestation_name]
#
# Attested via: kosli attest pullrequest github --name <attestation_name> --commit <sha>
# ---------------------------------------------------------------------------

# Extract PR attestation payload from a trail.
pr_attest(trail) := trail.compliance_status.attestations_statuses[attestation_name]

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# A username is resolved when it is a non-empty GitHub login. "ghost" is the
# placeholder GitHub shows for deleted accounts, so it identifies no one.
is_resolved_username(u) if {
	is_string(u)
	u != ""
	u != "ghost"
}

# GitHub usernames of all PR branch commit authors.
pr_commit_authors(pr) := {c.author_username |
	some c in pr.commits
	is_resolved_username(c.author_username)
}

# Approver usernames that can satisfy four-eyes for commits up to cutoff.
approved_approvers_after_cutoff(pr, cutoff) := {a.username |
	some a in pr.approvers
	a.state == "APPROVED"
	is_resolved_username(a.username)
	a.timestamp > cutoff
}

# Latest Unix timestamp among PR branch commits.
latest_commit_ts(pr) := max({c.timestamp | some c in pr.commits})

# Every commit on the PR has an author linked to a GitHub account.
all_authors_resolved(pr) if {
	every c in pr.commits {
		is_resolved_username(object.get(c, "author_username", null))
	}
}

# A commit is the merge commit when the PR's merge_commit field matches the
# trail name (which is the commit SHA). Covers squash, regular, and rebase merges.
is_merge_commit(trail, pr) if {
	trail.name == pr.merge_commit
}

# Regular commit: PR branch authors + PR author all need independent approval after last code commit.
has_independent_approval(trail, pr) if {
	not is_merge_commit(trail, pr)
	cutoff := latest_commit_ts(pr)
	all_authors := pr_commit_authors(pr) | {pr.author}
	eligible_approvers := approved_approvers_after_cutoff(pr, cutoff)
	count(all_authors) > 0

	# At least one approver must exist to satisfy four-eyes.
	count(eligible_approvers) > 0
	every author in all_authors {
		some approver in eligible_approvers
		approver != author
	}
}

# Merge commit: only PR branch commit authors need independent approval.
# The merge button clicker did not write code and requires no separate review.
has_independent_approval(trail, pr) if {
	is_merge_commit(trail, pr)
	cutoff := latest_commit_ts(pr)
	all_authors := pr_commit_authors(pr)
	eligible_approvers := approved_approvers_after_cutoff(pr, cutoff)
	count(all_authors) > 0

	# At least one approver must exist to satisfy four-eyes.
	count(eligible_approvers) > 0
	every author in all_authors {
		some approver in eligible_approvers
		approver != author
	}
}

# ---------------------------------------------------------------------------
# Violations - human-readable diagnostic output
#
# These are derived for debugging and reporting only. allow does NOT depend
# on this set: a sprintf failure here cannot affect the compliance decision.
# A trail appears in violations if and only if it is not in trail_compliant.
# ---------------------------------------------------------------------------

violations contains "Policy error: input.trails is missing or not an array - cannot evaluate" if {
	not is_array(object.get(input, "trails", null))
}

violations contains "Policy error: input.trails is empty - nothing to evaluate" if {
	is_array(input.trails)
	count(input.trails) == 0
}

# Missing attestation: no PR review data collected for this commit.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	not trail.compliance_status.attestations_statuses[attestation_name]
	msg := sprintf("Trail %v: %v attestation is missing", [trail.name, attestation_name])
}

# Unverifiable identity: commit author has no resolvable GitHub account
# (missing, empty, or "ghost").
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	attest := pr_attest(trail)
	some pr in attest.pull_requests
	some c in pr.commits
	not is_resolved_username(object.get(c, "author_username", null))
	msg := sprintf(
		"PR %v: commit %v has no linked GitHub account - identity unverifiable",
		[pr.url, c.sha1],
	)
}

# Missing PR: commit has no associated merged PR.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	attest := pr_attest(trail)
	count(attest.pull_requests) == 0
	msg := sprintf("Commit %v: no associated PR found", [trail.name])
}

# Missing approval: commit has an associated PR but no PR satisfies the
# independent-approval requirement.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	attest := pr_attest(trail)
	count(attest.pull_requests) > 0
	not any_pr_fully_approved(trail, attest)
	msg := sprintf(
		"Commit %v: no independent approval after latest code commit",
		[trail.name],
	)
}

# True if any associated PR has both resolved authors and independent approval.
# Used only for violation messaging to distinguish "missing approval" from
# "unverifiable identity".
any_pr_fully_approved(trail, attest) if {
	some pr in attest.pull_requests
	all_authors_resolved(pr)
	has_independent_approval(trail, pr)
}
