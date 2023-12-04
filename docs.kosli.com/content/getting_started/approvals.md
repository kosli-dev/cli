---
title: "Part 7: Approvals"
bookCollapseSection: false
weight: 260
---
# Part 7: Approvals

Whenever an artifact is ready to be deployed to a given [environment](/getting_started/environments/), an additional approval may be created in Kosli. An approval can be requested which will require a manually action, or reported automatically. This will be recorded into Kosli so the decision made outside of your CI system won't be lost.

When an approval is created for an artifact to a specific environment with the `--environment` flag, Kosli will generate a list of commits to be approved. By default, this list will contain all commits between `HEAD` and the commit of the most recent artifact coming from the same [flow](/getting_started/flows/) found in the given environment. The list can also be specified by providing values for `--newest-commit` and `--oldest-commit`. If you are providing these commits yourself, keep in mind that `--oldest-commit` has to be an ancestor of `--newest-commit`.


## Example

{{< tabs "approvals" "col-no-wrap" >}}

{{< tab "v2" >}}
```
$ kosli report approval project-a-app.bin \
  --artifact-type file \
  --environment production \
  --flow project-a 
  
approval created for artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```

See [kosli report approval](/client_reference/kosli_report_approval/) and [kosli request approval](/client_reference/kosli_request_approval/) for more details. 
{{< /tab >}}

{{< tab "v0.1.x" >}}
```
$ kosli pipeline approval report project-a-app.bin \
  --artifact-type file \
  --newest-commit $(git rev-parse HEAD) \
  --oldest-commit $(git rev-parse HEAD~1) \
  --pipeline project-a 
  
approval created for artifact: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```

See [kosli pipeline approval report](/legacy_ref/v0.1.41/kosli_pipeline_approval_report/) and [kosli pipeline approval request](/legacy_ref/v0.1.41/kosli_pipeline_approval_request/) for more details. 
{{< /tab >}}

{{< /tabs >}}
