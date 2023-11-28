---
title: "Part 7: Approvals"
bookCollapseSection: false
weight: 260
---
# Part 7: Approvals

Whenever a given artifact is ready to be deployed you may need an additional approval. In Kosli, an Approval means an artifact is ready to be deployed to a given [environment](/getting_started/environments/). An approval can be manually or automatically, and recorded into Kosli so the decision made outside of your CI system won't be lost.

When an Approval is created for an artifact to a specific environment with the `--environment` flag, Kosli will generate a list of commits to be approved. By default, this list will contain all commits between `HEAD` and the commit of the latest artifact coming from the same [flow](/getting_started/flows/) found in the given environment. The list can also be specified by providing values for `--newest-commit` and `--oldest-commit`.

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

{{< hint warning >}}

## Quick note about a commit list

If you are providing the oldest and newest commits yourself, keep in mind that `--oldest-commit` has to be an ancestor of `--newest-commit`. 

It's easy to verify locally in the repository using:
```shell {.command}
git merge-base --is-ancestor <oldest-commit> <newest-commit>
echo $?
```

`echo $?` checks the exit code of previous command so it's important you run it directly after `git merge-base <...>` command.  

Exit code 0 means `oldest-commit` is an ancestor of `newest-commit` and your kosli approval command will work. If the exit code is different than 0 then we won't be able to generate a list of commits needed for an approval and the command will fail.

To be able to trace back the history of your commits we need a complete repository history to be available - in your CI pipelines it'll likely mean you have to explicitly check out the whole history (many CI tools checkout just a latest version by default).

In GitHub Actions you'd need to modify the checkout step by adding fetch-depth option (zero means full depth):

```
steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 0
```

{{< /hint >}}
