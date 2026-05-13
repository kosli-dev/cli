package policy

import rego.v1

# Four-eyes principle enforcement: every commit must have independent review.
# This policy evaluates per-commit attestation data from Kosli and passes only
# when all violations are resolved.
default allow = false

allow if count(violation_reasons) == 0

# Set PR attestation name
attestation_name := name if {
    name := data.params.attestation_name
    is_string(name)
} else := "pr-review"

# ---------------------------------------------------------------------------
# Attestation data
#
# Used with `kosli evaluate trails` (plural). Each trail in input.trails
# represents one commit. The PR attestation payload is at:
#   trail.compliance_status.attestations_statuses["pr-review"]
#
# Attested via: kosli attest pullrequest github --name pr-review --commit <sha>
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
	"noreply@github.com"
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
# Helpers — multi-PR support
# ---------------------------------------------------------------------------

# Check if any associated PR has independent approval for the commit.
has_any_pr_approval(trail, attest) if {
	some pr in attest.pull_requests
	has_independent_approval(trail, pr)
}

# ---------------------------------------------------------------------------
# Violation reasons — detection only, no sprintf
#
# allow is derived from this set. Keeping detection logic here (no sprintf)
# means a formatting failure in the violations rules below cannot silently
# empty this set and flip allow to true.
# ---------------------------------------------------------------------------

# Guard: if input.trails is absent or not an array, every other rule silently
# skips iteration and violation_reasons stays empty, making allow=true.
# object.get ensures the argument to is_array is always defined
# (avoids undefined-arg propagation that would make `not is_array(undefined)`
# silently skip the rule).
violation_reasons contains "missing_trails_input" if {
	trails := object.get(input, "trails", null)
	not is_array(trails)
}

# Missing attestation: no PR review data collected for this commit.
violation_reasons contains {"type": "missing_attestation", "trail": trail.name} if {
	some trail in input.trails
	not trail.compliance_status.attestations_statuses[attestation_name]
}

# Unverifiable identity: commit author has no resolvable GitHub account and is not a known service account.
violation_reasons contains {"type": "unverifiable_identity", "pr_url": pr.url, "sha": c.sha1} if {
	some trail in input.trails
	attest := pr_attest(trail)
	some pr in attest.pull_requests
	some c in pr.commits
	object.get(c, "author_username", null) == null
	not is_service_account(trail)
	not is_web_flow_commit(c)
}

# Missing PR: commit has no associated merged pull request (non-service-account commits must come through a PR).
violation_reasons contains {"type": "missing_pr", "trail": trail.name} if {
	some trail in input.trails
	not is_service_account(trail)
	attest := pr_attest(trail)
	count(attest.pull_requests) == 0
}

# Missing approval: commit has an associated PR but no independent approval from someone other than the authors.
violation_reasons contains {"type": "missing_approval", "trail": trail.name} if {
	some trail in input.trails
	not is_service_account(trail)
	attest := pr_attest(trail)
	count(attest.pull_requests) > 0
	not has_any_pr_approval(trail, attest)
}

# ---------------------------------------------------------------------------
# Violations — message formatting only
#
# allow does NOT depend on this set. A sprintf failure here cannot affect
# the compliance decision; it only affects the human-readable output.
# ---------------------------------------------------------------------------

violations contains "Policy error: input.trails is missing or not an array — cannot evaluate" if {
	"missing_trails_input" in violation_reasons
}

violations contains msg if {
	some r in violation_reasons
	r.type == "missing_attestation"
	msg := sprintf("Trail %v: %v attestation is missing", [r.trail, attestation_name])
}

violations contains msg if {
	some r in violation_reasons
	r.type == "unverifiable_identity"
	msg := sprintf(
		"PR %v: commit %v has no linked GitHub account — identity unverifiable",
		[r.pr_url, substring(r.sha, 0, 7)],
	)
}

violations contains msg if {
	some r in violation_reasons
	r.type == "missing_pr"
	msg := sprintf("Commit %v: no associated PR found", [substring(r.trail, 0, 7)])
}

violations contains msg if {
	some r in violation_reasons
	r.type == "missing_approval"
	msg := sprintf(
		"Commit %v: no independent approval after latest code commit",
		[substring(r.trail, 0, 7)],
	)
}
