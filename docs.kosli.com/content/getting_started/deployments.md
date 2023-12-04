---
title: "Part 8: Deployments"
bookCollapseSection: false
weight: 270
---
# Part 8: Deployments

The last important piece of information, when it comes to artifacts are deployments. Whenever you (likely with the use of your CI system) deploy an artifact to an environment you should record that information to Kosli.

When you check the status of your environments you want to know that what is running there was **expected** to run there. Reporting deployment expectation is an easy way of detecting unauthorized workloads and manual changes.

{{< hint info >}}

Run [kosli expect deployment](/client_reference/kosli_expect_deployment/) right **before** you start the actual deployment.

{{< /hint >}}

### Example

{{< tabs "deployments" "col-no-wrap" >}}

{{< tab "v2" >}}
```
$ kosli expect deployment project-a-app.bin \
    --artifact-type file \
    --build-url aaahttps://exampleci.com \
    --environment server-prod \
    --flow project-a

deployment of artifact 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf was reported to: server-prod
```
See [kosli expect deployment](/client_reference/kosli_expect_deployment/) for more details.
{{< /tab >}}

{{< tab "v0.1.x" >}}
```
$ kosli expect deployment project-a-app.bin \
    --artifact-type file \
    --build-url aaahttps://exampleci.com \
    --environment server-prod \
    --pipeline project-a

expect deployment of artifact 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf was reported to: server-prod
```
See [kosli expect deployment](/legacy_ref/v0.1.41/kosli_expect_deployment/) for more details.
{{< /tab >}}

{{< /tabs >}}
