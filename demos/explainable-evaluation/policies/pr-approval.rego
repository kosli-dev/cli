# METADATA
# title: Pull-request compliance
# description: |
#   A pull request is compliant either because it was authored by an
#   automated bot account, or because at least one approver signed off.
package policy

import rego.v1

default allow := false

allow if { pr_compliant }

# METADATA
# title: PR is compliant
# scope: document

# METADATA
# title: bot-authored PR
# scope: rule
pr_compliant if {
	input.pr.author == "bot"
}

# METADATA
# title: human-authored PR has an approver
# scope: rule
pr_compliant if {
	count(input.pr.approvers) > 0
}
