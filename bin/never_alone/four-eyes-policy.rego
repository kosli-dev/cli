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

# ---------------------------------------------------------------------------
# Compliance - a trail is compliant if any of these positive conditions hold
# ---------------------------------------------------------------------------

# A compliant initial-commit attestation substitutes for a PR review.
trail_compliant(trail) if {
	attest := initial_commit_attest(trail)
	attest.is_compliant == true
}

# Every approver on the PR has a resolvable identity.
# To be considered resolved, there must be at least one approver, and
# every approver's username must be a non-empty string and not "ghost".
all_approvers_resolved(pr) if {
	count(pr.approvers) > 0
	every a in pr.approvers {
		is_string(a.username)
		a.username != ""
		a.username != "ghost"
	}
}

# Compliance helper: ignore unresolved commit authors if all approvers are resolved.
authors_resolved_or_approvers_resolved(pr) if {
	all_authors_resolved(pr)
}

authors_resolved_or_approvers_resolved(pr) if {
	all_approvers_resolved(pr)
}

# Commits are compliant when an associated PR has independent approval
# covering every author after the latest code commit.
trail_compliant(trail) if {
	attest := pr_attest(trail)
	some pr in attest.pull_requests
	authors_resolved_or_approvers_resolved(pr)
	has_independent_approval(trail, pr)
}

# ---------------------------------------------------------------------------
# Attestation data
#
# Used with `kosli evaluate trails` (plural). Each trail in input.trails
# represents one commit. The PR attestation payload is found by type, not by
# name, so any attestation with attestation_type == "pull_request" qualifies.
#
# Attested via: kosli attest pullrequest github --name <name> --commit <sha>
# ---------------------------------------------------------------------------

# Extract PR attestation payload from a trail by type.
pr_attest(trail) := attest if {
	some attest in trail.compliance_status.attestations_statuses
	attest.attestation_type == "pull_request"
}

# Extract initial-commit attestation from a trail.
initial_commit_attest(trail) := attest if {
	some attest in trail.compliance_status.attestations_statuses
	attest.attestation_type == "custom:initial-commit-by-verified-committer"
}

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# GitHub usernames of all PR branch commit authors whose identity was resolved.
pr_commit_authors(pr) := {c.author_username |
	some c in pr.commits
	is_string(c.author_username)
	c.author_username != ""
}

# Approver usernames that can satisfy four-eyes constraints for a cutoff.
approved_approvers_after_cutoff(pr, cutoff) := {a.username |
	some a in pr.approvers
	a.state == "APPROVED"
	is_string(a.username)
	a.username != ""
	a.username != "ghost"
	a.timestamp > cutoff
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
	c.author_username != ""
	c.author_username != "ghost"
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
	eligible_approvers := approved_approvers_after_cutoff(pr, cutoff)
	count(all_authors) > 0

	# At least one approver must exist to satisfy four-eyes.
	count(pr.approvers) > 0
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
	count(pr.approvers) > 0
	every author in all_authors {
		some approver in eligible_approvers
		approver != author
	}
}

# Merge commit fallback: if branch commits are all web-flow/unresolved,
# treat PR author as the code author requiring independent approval.
has_independent_approval(trail, pr) if {
	is_merge_commit(trail, pr)
	cutoff := latest_commit_ts(pr)
	count(pr_commit_authors(pr)) == 0
	is_string(pr.author)
	pr.author != ""
	eligible_approvers := approved_approvers_after_cutoff(pr, cutoff)

	# At least one approver must exist to satisfy four-eyes.
	count(pr.approvers) > 0
	every author in {pr.author} {
		some approver in eligible_approvers
		approver != author
	}
}

# ---------------------------------------------------------------------------
# Web-flow / bot commit tolerance
#
# These patterns identify PR-branch commits (not trails) whose author identity
# cannot be resolved to a GitHub account - e.g. web-flow edits or bot-signed
# commits made in the course of an otherwise human-authored PR. They do not
# exempt a trail from the PR-review requirement; every trail must still have
# an associated PR with independent approval.
# ---------------------------------------------------------------------------

service_account_patterns := {
	"^svc_[a-zA-Z0-9_-]+ <[^>]+>$", # Anchored to svc_ names
	"^.*?\\[bot\\] <[^>]+>$", # Anchored to bot names
}

# PR commit author is unresolvable (web-flow edits, Copilot co-auth).
is_web_flow_commit(c) if {
	some pattern in service_account_patterns
	regex.match(pattern, object.get(c, "author", ""))
}

is_web_flow_commit(c) if {
	regex.match("^GitHub <noreply@github.com>$", object.get(c, "author", ""))
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
	not pr_attest(trail)
	not initial_commit_attest(trail)
	msg := sprintf("Trail %v: pull_request attestation is missing", [trail.name])
}

# Non-compliant initial-commit attestation.
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	attest := initial_commit_attest(trail)
	attest.is_compliant != true
	msg := sprintf("Trail %v: initial-commit-by-verified-committer attestation is non-compliant", [trail.name])
}

# Unverifiable identity: commit author has no resolvable GitHub account
# and is not a tolerated web-flow/bot commit.
# Matches both null and empty string author_username (e.g. GitHub "ghost" users).
violations contains msg if {
	some trail in input.trails
	not trail_compliant(trail)
	attest := pr_attest(trail)
	some pr in attest.pull_requests
	not all_approvers_resolved(pr)
	some c in pr.commits
	username := object.get(c, "author_username", null)
	not is_web_flow_commit(c)
	_is_unresolved_username(username)
	msg := sprintf(
		"PR %v: commit %v has no linked GitHub account - identity unverifiable",
		[pr.url, c.sha1],
	)
}

_is_unresolved_username(u) if u == null
_is_unresolved_username(u) if u == ""
_is_unresolved_username(u) if u == "ghost"

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

# True if any associated PR has both resolved authors (or resolved approvers) and independent approval.
# Used only for violation messaging to distinguish "missing approval" from
# "unverifiable identity".
any_pr_fully_approved(trail, attest) if {
	some pr in attest.pull_requests
	authors_resolved_or_approvers_resolved(pr)
	has_independent_approval(trail, pr)
}