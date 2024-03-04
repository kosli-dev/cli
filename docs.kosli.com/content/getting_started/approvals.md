---
title: "Part 9: Approvals"
bookCollapseSection: false
weight: 300
---
# Part 9: Approvals

When an artifact is ready to be deployed to a given [environment](/getting_started/environments/), an approval may be reported to Kosli. An approval can be requested which will require a manual action, or reported automatically. This will be recorded in Kosli so the decision made outside your CI system won't be lost.

When an approval is created for an artifact to a specific environment with the `--environment` flag, Kosli will generate a list of commits to be approved. By default, this list will contain all commits between `HEAD` and the commit of the most recent artifact coming from the same [flow](/getting_started/flows/) found in the given environment. The list can also be specified by providing values for `--newest-commit` and `--oldest-commit`. If you are providing these commits yourself, keep in mind that `--oldest-commit` has to be an ancestor of `--newest-commit`.

See [request approval](/client_reference/kosli_request_approval/) and [report approval](/client_reference/kosli_report_approval/) for usage details and examples. 
