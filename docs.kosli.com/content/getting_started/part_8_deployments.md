---
title: "Part 8: Deployments"
bookCollapseSection: false
weight: 270
---
# Part 8: Deployments

The last important piece of information, when it comes to artifacts are deployments. Whenever you (likely with the use of your CI system) deploy an artifact to an environment you should record that information to Kosli. So when you check the status of your environments you know not only what is running there, but also how did it get there. It's an easy way of detecting a manual change was introduced.

### Example
```
$ kosli expect deployment project-a-app.bin \
    --artifact-type file \
    --build-url aaahttps://exampleci.com \
    --environment server-prod \
    --pipeline project-a 
    
deployment of artifact 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf was reported to: server-prod
```