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
# Compliance — a trail is compliant if any of these positive conditions hold
# ---------------------------------------------------------------------------

# Service-account commits are exempt from PR review.
trail_compliant(trail) if {
	is_service_account(trail)
}

# Human-authored commits are compliant when an associated PR has independent
# approval covering every author after the latest code commit.
trail_compliant(trail) if {
	not is_service_account(trail)
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

# GitHub usernames of all PR branch commit authors whose identity was resolved.
pr_commit_authors(pr) := {u |
	some c in pr.commits
	u := c.author_username
	u != null
}

# Latest Unix timestamp among PR branch commits.
latest_commit_ts(pr) := max({c.timestamp | some c in pr.commits})

# Every commit on the PR has a resolvable author (or is a known service-account
# style commit like web-flow / Copilot co-auth that we tolerate).
all_authors_resolved(pr) if {
	every c in pr.commits {
		author_resolved_or_exempt(c)
	}
}

author_resolved_or_exempt(c) if {
	is_string(c.author_username)
}

author_resolved_or_exempt(c) if {
	is_web_flow_commit(c)
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
	count(all_authors) > 0

	# At least one approver must exist to satisfy four-eyes.
	count(pr.approvers) > 0
	every author in all_authors {
		some approver in pr.approvers
		approver.state == "APPROVED"
		is_string(approver.username)
		approver.username != author
		approver.timestamp > cutoff
	}
}

# Merge commit: only PR branch commit authors need independent approval.
# The merge button clicker did not write code and requires no separate review.
has_independent_approval(trail, pr) if {
	is_merge_commit(trail, pr)
	cutoff := latest_commit_ts(pr)
	all_authors := pr_commit_authors(pr)
	count(all_authors) > 0

	# At least one approver must exist to satisfy four-eyes.
	count(pr.approvers) > 0
	every author in all_authors {
		some approver in pr.approvers
		approver.state == "APPROVED"
		is_string(approver.username)
		approver.username != author
		approver.timestamp > cutoff
	}
}

# ---------------------------------------------------------------------------
# Service account exemption
#
# Matched against trail.git_commit_info.author, which is "Name <email>" format.
# Patterns work against the full string, e.g.:
#   "github-actions[bot] <41898282+github-actions[bot]@users.noreply.github.com>"
# ---------------------------------------------------------------------------

service_account_patterns := {
	"svc_.*",
	".*\\[bot\\]",
	"noreply@github.com",
}

# Commit author is a service account (CI, GitHub Actions, dependabot, etc).
is_service_account(trail) if {
	some pattern in service_account_patterns
	regex.match(pattern, trail.git_commit_info.author)
}

# PR commit author is unresolvable (web-flow edits, Copilot co-auth).
is_web_flow_commit(c) if {
	some pattern in service_account_patterns
	regex.match(pattern, object.get(c, "author", ""))
}

# ---------------------------------------------------------------------------
# Violations — human-readable diagnostic output
#
# These are derived for debugging and reporting only. allow does NOT depend
# on this set: a sprintf failure here cannot affect the compliance decision.
# A trail appears in violations if and only if it is not in trail_compliant.
# ---------------------------------------------------------------------------

violations contains "Policy error: input.trails is missing or not an array — cannot evaluate" if {
	not is_array(object.get(input, "trails", null))
}

violations contains "Policy error: input.trails is empty — nothing to evaluate" if {
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
# and is not a known service account or web-flow commit.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	attest := pr_attest(trail)
	some pr in attest.pull_requests
	some c in pr.commits
	object.get(c, "author_username", null) == null
	not is_service_account(trail)
	not is_web_flow_commit(c)
	msg := sprintf(
		"PR %v: commit %v has no linked GitHub account — identity unverifiable",
		[pr.url, substring(c.sha1, 0, 7)],
	)
}

# Missing PR: non-service-account commit has no associated merged PR.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	not is_service_account(trail)
	attest := pr_attest(trail)
	count(attest.pull_requests) == 0
	msg := sprintf("Commit %v: no associated PR found", [substring(trail.name, 0, 7)])
}

# Missing approval: commit has an associated PR but no PR satisfies the
# independent-approval requirement.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	not is_service_account(trail)
	attest := pr_attest(trail)
	count(attest.pull_requests) > 0
	not any_pr_fully_approved(trail, attest)
	msg := sprintf(
		"Commit %v: no independent approval after latest code commit",
		[substring(trail.name, 0, 7)],
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
