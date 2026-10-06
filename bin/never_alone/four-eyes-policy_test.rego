package policy_test

import data.policy
import rego.v1

# Run: opa test bin/never_alone/four-eyes-policy.rego bin/never_alone/four-eyes-policy_test.rego -v

first := "1111111111111111111111111111111111111111"

later := "2222222222222222222222222222222222222222"

merge := "3333333333333333333333333333333333333333"

params := {"repository": "o/r"}

# A commit written and signed by user.
commit(sha, ts, user) := signed_commit(sha, ts, user, user)

# A commit naming author as its author, with a verified signature by signer.
signed_commit(sha, ts, author, signer) := {
	"sha1": sha, "author": sprintf("%v <%v@example.com>", [author, author]), "author_username": author, "timestamp": ts,
	"verified": true, "signer_username": signer, "signed_by_platform": false,
}

# A submitted review by a person with write access.
review(user, sha, state, ts) := {
	"username": user, "state": state, "timestamp": ts, "commit_sha": sha,
	"author_type": "user", "has_write_access": true,
}

approval(user, sha) := review(user, sha, "APPROVED", 1000050)

pr(commits, reviews, head) := {
	"url": "https://github.com/o/r/pull/1",
	"author": "alice",
	"merge_commit": merge,
	"head_sha": head,
	"merged_at": merged_at,
	"commits": commits,
	"commit_count": count(commits),
	"reviews": reviews,
}

merged_at := 2000000

trail(name, prs) := {
	"name": name,
	"git_commit_info": {"author": "alice <alice@example.com>", "sha1": name},
	"compliance_status": {"attestations_statuses": {"pr-review": {"attestation_type": "pull_request", "pull_requests": prs}}},
}

allowed(prs) if policy.allow with input as {"trails": [trail(merge, prs)]} with data.params as params

one_commit := [commit(first, 1000000, "alice")]

# A later push dated before the approval, as a backdated author or committer date would be.
backdated_push := array.concat(one_commit, [commit(later, 500, "alice")])

test_approval_on_head_passes if {
	allowed([pr(one_commit, [approval("bob", first)], first)])
}

test_push_after_approval_fails if {
	not allowed([pr(array.concat(one_commit, [commit(later, 1000100, "alice")]), [approval("bob", first)], later)])
}

test_backdated_push_after_approval_fails if {
	not allowed([pr(backdated_push, [approval("bob", first)], later)])
}

test_reapproval_on_new_head_passes if {
	allowed([pr(backdated_push, [approval("bob", first), approval("bob", later)], later)])
}

test_second_reviewer_on_new_head_passes if {
	allowed([pr(backdated_push, [approval("bob", first), approval("carol", later)], later)])
}

test_self_approval_on_head_fails if {
	not allowed([pr(one_commit, [approval("alice", first)], first)])
}

test_dismissed_review_on_head_fails if {
	not allowed([pr(one_commit, [object.union(approval("bob", first), {"state": "DISMISSED"})], first)])
}

test_missing_head_and_reviewed_commit_fails if {
	not allowed([pr(one_commit, [object.remove(approval("bob", first), ["commit_sha"])], null)])
}

test_missing_reviewed_commit_fails if {
	not allowed([pr(one_commit, [object.remove(approval("bob", first), ["commit_sha"])], first)])
}

test_missing_head_fails if {
	not allowed([object.remove(pr(one_commit, [approval("bob", first)], first), ["head_sha"])])
}

test_empty_head_and_reviewed_commit_fails if {
	not allowed([pr(one_commit, [approval("bob", "")], "")])
}

# The PR author must be covered when the trail is not the merge commit.
test_non_merge_trail_approved_on_head_passes if {
	policy.allow with input as {"trails": [trail(later, [pr(one_commit, [approval("bob", first)], first)])]}
		with data.params as params
}

test_non_merge_trail_self_approved_by_pr_author_fails if {
	not policy.allow with input as {"trails": [trail(later, [pr([commit(first, 1000000, "dave")], [approval("alice", first)], first)])]}
		with data.params as params
}

# The policy reads no author name, so a bot-like or service-account name is reviewed like any other.
test_bot_named_trail_author_needs_review if {
	t := object.union(trail(merge, [pr(backdated_push, [approval("bob", first)], later)]), {"git_commit_info": {"author": "alice[bot] <a@example.com>", "sha1": merge}})
	not policy.allow with input as {"trails": [t]} with data.params as params
}

fork_pr(p) := object.union(p, {"url": "https://github.com/alice/fork/pull/1"})

test_fork_pr_approval_does_not_count if {
	not allowed([
		pr(backdated_push, [approval("bob", first)], later),
		fork_pr(pr(backdated_push, [approval("alice-alt", later)], later)),
	])
}

test_unrelated_fork_pr_does_not_block if {
	allowed([pr(one_commit, [approval("bob", first)], first), fork_pr(pr(one_commit, [], first))])
}

test_enterprise_host_in_repo_passes if {
	allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"url": "https://ghe.example.com/o/r/pull/1"})])
}

test_malformed_pr_url_fails if {
	not allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"url": "https://github.com/o/r/pull/1/files"})])
}

test_repository_compared_case_insensitively if {
	policy.allow with input as {"trails": [trail(merge, [pr(one_commit, [approval("bob", first)], first)])]}
		with data.params as {"repository": "O/R"}
}

test_missing_repository_param_fails_and_says_why if {
	inp := {"trails": [trail(merge, [pr(one_commit, [approval("bob", first)], first)])]}
	not policy.allow with input as inp
	v := policy.violations with input as inp
	some msg in v
	contains(msg, "data.params.repository is not set")
}

test_unapproved_commit_violation_names_the_repository if {
	v := policy.violations with input as {"trails": [trail(merge, [pr(backdated_push, [approval("bob", first)], later)])]}
		with data.params as params
	v == {sprintf("Commit %v: no PR in o/r has an independent approval on its final commit before merge", [merge])}
}

test_null_head_and_reviewed_commit_fails if {
	not allowed([pr(one_commit, [object.union(approval("bob", first), {"commit_sha": null})], null)])
}

test_pr_url_compared_case_insensitively if {
	allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"url": "https://github.com/O/R/pull/1"})])
}

test_non_pull_request_url_fails if {
	not allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"url": "https://github.com/o/r/issues/1"})])
}

test_unresolved_commit_author_blocks_approval if {
	not allowed([pr(array.concat(one_commit, [{"sha1": later, "author": "x <x@example.com>", "timestamp": 1000000}]), [approval("bob", later)], later)])
}

test_ghost_commit_author_blocks_approval if {
	not allowed([pr(array.concat(one_commit, [commit(later, 1000000, "ghost")]), [approval("bob", later)], later)])
}

test_ghost_approver_does_not_count if {
	not allowed([pr(one_commit, [approval("ghost", first)], first)])
}

test_empty_commit_author_blocks_approval if {
	not allowed([pr(array.concat(one_commit, [commit(later, 1000000, "")]), [approval("bob", later)], later)])
}

test_unsigned_commit_blocks_approval if {
	not allowed([pr([object.remove(commit(first, 1000000, "alice"), ["verified", "signer_username", "signed_by_platform"])], [approval("bob", first)], first)])
}

test_invalid_signature_blocks_approval if {
	not allowed([pr([object.union(commit(first, 1000000, "alice"), {"verified": false})], [approval("bob", first)], first)])
}

test_signature_without_known_signer_blocks_approval if {
	not allowed([pr([object.remove(commit(first, 1000000, "alice"), ["signer_username"])], [approval("bob", first)], first)])
}

test_ghost_signer_blocks_approval if {
	not allowed([pr([signed_commit(first, 1000000, "alice", "ghost")], [approval("bob", first)], first)])
}

test_commit_signed_by_platform_passes if {
	allowed([pr([object.union(commit(first, 1000000, "alice"), {"signer_username": "web-flow", "signed_by_platform": true})], [approval("bob", first)], first)])
}

# Naming someone else as the author doesn't let the signer approve their own work.
test_signer_cannot_approve_own_commit_under_another_author if {
	not allowed([pr([signed_commit(first, 1000000, "bob", "alice")], [approval("alice", first)], first)])
}

test_independent_approval_covers_author_and_signer if {
	allowed([pr([signed_commit(first, 1000000, "bob", "alice")], [approval("carol", first)], first)])
}

test_unsigned_commit_violation_says_why if {
	unsigned := object.remove(commit(first, 1000000, "alice"), ["verified", "signer_username", "signed_by_platform"])
	v := policy.violations with input as {"trails": [trail(merge, [pr([unsigned], [approval("bob", first)], first)])]}
		with data.params as params
	some msg in v
	contains(msg, "no verified signature")
}

test_invalid_github_signature_blocks_approval if {
	not allowed([pr([object.union(commit(first, 1000000, "alice"), {"signer_username": "web-flow", "signed_by_platform": true, "verified": false})], [approval("bob", first)], first)])
}

test_non_merge_trail_with_old_approval_fails if {
	not policy.allow with input as {"trails": [trail(later, [pr(backdated_push, [approval("bob", first)], later)])]}
		with data.params as params
}

test_unlinked_author_with_known_signer_blocks_approval if {
	not allowed([pr([object.remove(commit(first, 1000000, "alice"), ["author_username"])], [approval("bob", first)], first)])
}

test_non_merge_trail_pr_author_needs_independent_approval if {
	not policy.allow with input as {"trails": [trail(later, [object.union(pr(one_commit, [approval("carol", first)], first), {"author": "carol"})])]}
		with data.params as params
}

test_dismissed_review_on_head_fails_on_non_merge_trail if {
	not policy.allow with input as {"trails": [trail(later, [pr(one_commit, [object.union(approval("bob", first), {"state": "DISMISSED"})], first)])]}
		with data.params as params
}

test_bot_approval_does_not_count if {
	not allowed([pr(one_commit, [object.union(approval("bob", first), {"author_type": "bot"})], first)])
}

test_approval_from_other_actor_type_does_not_count if {
	not allowed([pr(one_commit, [object.union(approval("bob", first), {"author_type": "other"})], first)])
}

test_approval_without_write_access_does_not_count if {
	not allowed([pr(one_commit, [object.union(approval("bob", first), {"has_write_access": false})], first)])
}

test_approval_with_unknown_write_access_does_not_count if {
	not allowed([pr(one_commit, [object.remove(approval("bob", first), ["has_write_access"])], first)])
}

test_later_request_for_changes_withdraws_approval if {
	not allowed([pr(one_commit, [approval("bob", first), review("bob", first, "CHANGES_REQUESTED", 1000060)], first)])
}

test_same_second_request_for_changes_withdraws_approval if {
	not allowed([pr(one_commit, [approval("bob", first), review("bob", first, "CHANGES_REQUESTED", 1000050)], first)])
}

test_later_dismissal_withdraws_approval if {
	not allowed([pr(one_commit, [approval("bob", first), review("bob", first, "DISMISSED", 1000060)], first)])
}

test_request_for_changes_without_time_withdraws_approval if {
	not allowed([pr(one_commit, [approval("bob", first), object.remove(review("bob", first, "CHANGES_REQUESTED", 0), ["timestamp"])], first)])
}

test_earlier_request_for_changes_does_not_withdraw if {
	allowed([pr(one_commit, [review("bob", first, "CHANGES_REQUESTED", 1000040), approval("bob", first)], first)])
}

test_later_comment_does_not_withdraw if {
	allowed([pr(one_commit, [approval("bob", first), review("bob", first, "COMMENTED", 1000060)], first)])
}

test_another_reviewers_request_for_changes_does_not_withdraw if {
	allowed([pr(one_commit, [approval("bob", first), review("carol", first, "CHANGES_REQUESTED", 1000060)], first)])
}

test_approval_without_time_does_not_count if {
	not allowed([pr(one_commit, [object.remove(approval("bob", first), ["timestamp"])], first)])
}

# Attestations from a CLI that records approvers only, with no reviews.
test_pr_without_reviews_fails if {
	not allowed([object.union(object.remove(pr(one_commit, [], first), ["reviews"]), {"approvers": [approval("bob", first)]})])
}

test_request_for_changes_with_null_time_withdraws_approval if {
	not allowed([pr(one_commit, [approval("bob", first), object.union(review("bob", first, "CHANGES_REQUESTED", 0), {"timestamp": null})], first)])
}

test_comment_alone_is_not_an_approval if {
	not allowed([pr(one_commit, [review("bob", first, "COMMENTED", 1000050)], first)])
}

# A commit GitHub signed names web-flow as signer; the named author still needs an independent approval.
test_platform_signed_commit_author_cannot_self_approve if {
	not allowed([pr([object.union(commit(first, 1000000, "alice"), {"signer_username": "web-flow", "signed_by_platform": true})], [approval("alice", first)], first)])
}

test_platform_signed_commit_without_signer_passes if {
	allowed([pr([object.union(object.remove(commit(first, 1000000, "alice"), ["signer_username"]), {"signed_by_platform": true})], [approval("bob", first)], first)])
}

test_approval_after_merge_fails if {
	not allowed([pr(one_commit, [review("bob", first, "APPROVED", merged_at + 1)], first)])
}

test_approval_at_merge_time_passes if {
	allowed([pr(one_commit, [review("bob", first, "APPROVED", merged_at)], first)])
}

test_unmerged_pr_fails if {
	not allowed([object.remove(pr(one_commit, [approval("bob", first)], first), ["merged_at"])])
}

test_pr_listing_fewer_commits_than_github_counts_fails if {
	p := object.union(pr(one_commit, [approval("bob", first)], first), {"commit_count": 251})
	not allowed([p])
	v := policy.violations with input as {"trails": [trail(merge, [p])]} with data.params as params
	some msg in v
	contains(msg, "lists 1 commits but GitHub counts 251")
}

test_pr_without_commit_count_fails if {
	not allowed([object.remove(pr(one_commit, [approval("bob", first)], first), ["commit_count"])])
}

# An AI agent's commit names the person who asked for it as co-author.
agent_commit := object.union(commit(first, 1000000, "Copilot"), {"signer_username": "web-flow", "signed_by_platform": true, "co_author_usernames": ["alice"]})

test_co_author_cannot_self_approve if {
	not allowed([pr([agent_commit], [approval("alice", first)], first)])
}

test_co_authored_commit_with_independent_approval_passes if {
	allowed([pr([agent_commit], [approval("bob", first)], first)])
}

test_null_merge_time_and_commit_count_fail if {
	not allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"merged_at": null})])
	not allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"commit_count": null})])
	not allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"commit_count": "1"})])
}

test_wrong_type_merge_time_fails if {
	not allowed([object.union(pr(one_commit, [review("bob", first, "APPROVED", merged_at + 1)], first), {"merged_at": "0"})])
}

test_pr_listing_more_commits_than_github_counts_fails if {
	not allowed([object.union(pr(one_commit, [approval("bob", first)], first), {"commit_count": 0})])
}

test_co_authors_as_a_string_fails if {
	not allowed([pr([object.union(commit(first, 1000000, "Copilot"), {"co_author_usernames": "alice"})], [approval("bob", first)], first)])
}

test_missing_commit_count_says_to_reattest if {
	p := object.remove(pr(one_commit, [approval("bob", first)], first), ["commit_count"])
	v := policy.violations with input as {"trails": [trail(merge, [p])]} with data.params as params
	some msg in v
	contains(msg, "no commit count recorded")
}

test_co_author_entries_must_be_accounts if {
	not allowed([pr([object.union(commit(first, 1000000, "Copilot"), {"co_author_usernames": [["alice"]]})], [approval("alice", first)], first)])
	not allowed([pr([object.union(commit(first, 1000000, "Copilot"), {"co_author_usernames": ["ghost"]})], [approval("bob", first)], first)])
}

test_non_merge_trail_needs_a_resolved_pr_author if {
	every author in [null, "", "ghost"] {
		not policy.allow with input as {"trails": [trail(later, [object.union(pr(one_commit, [approval("bob", first)], first), {"author": author})])]}
			with data.params as params
	}
}

test_null_commit_count_gives_no_count_mismatch_message if {
	p := object.union(pr(one_commit, [approval("bob", first)], first), {"commit_count": null})
	v := policy.violations with input as {"trails": [trail(merge, [p])]} with data.params as params
	every msg in v {
		not contains(msg, "GitHub counts")
	}
}
