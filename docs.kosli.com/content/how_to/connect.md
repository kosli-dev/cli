---
title: Connect
bookCollapseSection: false
weight: 40
---
# Connect what you build with what's running

Kosli allows you to connect the development world (commits, builds, tests, approvals, deployments) with what’s happening in operations. There is a variety of commands that let you report all the necessary information to Kosli and - relying on automatically calculated fingerprints of your artifacts - match it with the environments.

## Artifacts

Whenever you build an artifact you can report it to Kosli using our CLI. Fingerprint (sha256 checksum of the file/directory/docker image) of the artifact will be calculated and stored in Kosli. The fingerprint can't be changed, it becomes a unique identifier of the artifact in Kosli, used - among other things - to connect it with the reported environment. Fingerprints of all the running artifacts, recorded with Kosli CLI are also stored and compared with fingerprints of the artifacts you have built and reported to Kosli so you always know if you're running things you have no provenance of. 

See [kosli pipeline artifact report creation](/client_reference/kosli_pipeline_artifact_report_creation/) for more details. 

## Approvals

Whenever a given artifact is ready to be deployed you may need an additional manual approval from authorized person. This is something that can't alway be automated, but you can use Kosli to request such an approval, and later record it, so the information about decisions made outside of your CI system won't be lost. The list of commits between current and previous approval will be generated, which allows you to track a set of changes that are being approved.

See [kosli pipeline approval report](/client_reference/kosli_pipeline_approval_report/) and [kosli pipeline approval request](/client_reference/kosli_pipeline_approval_request/) for more details. 

### Quick note about a commit list

When reporting or requesting an approval one has to keep in mind that `oldest-commit` has to be an ancestor of `newest-commit`. 

It's easy to verify locally in the repository using:
```shell {.command}
git merge-base --is-ancestor <oldest-commit> <newest-commit>
echo $?
```

`echo $?` checks the exit code of previous command so it's important you run it directly after `git merge-base <...>` command.  

Exit code 0 means `oldest-commit` is an ancestor of `newest-commit` and your kosli approval command will work. If the exit code is different than 0 then we won't be able to generate a list of commits needed for an approval and the command will fail.

To be able to trace back the history of your commits we need a complete repository history to be available - in your CI pipelines it'll likely mean you have to explicitly check out the whole history (many CI tools checkout just a latest version by default).

In GitHub Actions you'd need to modify the checkout step by adding fetch-depth option:

```
steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 0
```

## Deployments

The last important piece of information, when it comes to artifacts are deployments. Whenever you (likely with the use of your CI system) deploy an artifact to an environment you should record that information to Kosli. So when you check the status of your environments you know not only what is running there, but also how did it get there. It's an easy way of detecting a manual change was introduced.