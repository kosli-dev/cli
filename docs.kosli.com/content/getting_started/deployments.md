---
title: "Part 9: Deployments"
bookCollapseSection: false
weight: 290
---
# Part 9: Deployments

Whenever you (likely with the use of your CI system) deploy an artifact to an environment you should record that information to Kosli to declare that the artifact is expected to run in the specific environment.

When you check the status of your environments you want to know that what is running there was **expected** to run there. Reporting deployment expectation is an easy way of detecting unauthorized workloads and manual changes.

{{< hint info >}}

You should report deployment expectation right **before** you start the actual deployment.

{{< /hint >}}

See [kosli expect deployment](/client_reference/kosli_expect_deployment/) for usage details and examples.