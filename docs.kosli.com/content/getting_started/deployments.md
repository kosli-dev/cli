---
title: "Part 8: Deployment Expectations"
bookCollapseSection: false
weight: 270
---
# Part 8: Deployment Expectations

The last important piece of information, when it comes to artifacts are deployments. Whenever you (likely with the use of your CI system) deploy an artifact to an environment you should record that information to Kosli.

So when you check the status of your environments you know, not only what is running there, but also that it was expected to run  there. It's an easy way of detecting unauthorized workloads and manual changes.

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
See [kosli expect deployment](/legacy_ref/v0.1.37/kosli_expect_deployment/) for more details.
{{< /tab >}}

{{< /tabs >}}
