---
title: Reporting in Kosli
bookCollapseSection: false
weight: 420
---
# Connect your pipeline events

Kosli allows you to connect the development world (commits, builds, tests, approvals, deployments) with what’s happening in operations. There is a variety of commands that let you report all the necessary information to Kosli and - relying on automatically calculated fingerprints of your artifacts - match it with the environments.

## Create a pipeline

In order to be able to report artifacts to Kosli you need to create a Kosli [pipeline](/introducing_kosli/pipelines) first. When you create a pipeline you also define expected controls - a list of evidences you need to be reported in order for the artifact to become compliant. Use `--template` flag to provide the list of requirements. 

Later, when reporting evidences for specific control you will use the same name you used in template to identify which evidence you are reporting.

It is a normal practice to include `kosli pipeline declare` command in the same pipeline you use to build the artifact you want to report to that Kosli pipeline. None of the previously reported artifacts will be overwritten or lost. And if you change the template, by adding or removing required evidence, it won't affect the compliancy status of existing artifacts.

### Example

```
# create/update a Kosli pipeline
kosli pipeline declare \
	--pipeline yourPipelineName \
	--description yourPipelineDescription \
  --visibility private OR public \
	--template artifact,unit-test,pull-request,code-coverage \
	--api-token yourAPIToken \
	--owner yourOrgName
```

## Report artifacts

To report an artifact you need either the artifact available while running reporting command, and use `--artifact-type` flag to make it possible for the tool to calculate the fingerprint OR you need a fingerprint of the artifact calculated separately using [kosli fingerprint](/client_reference/kosli_fingerprint/) command. You also need to provide the name of the Kosli pipeline you want to report to.

You also should provide long enough git history so Kosli can generate a list of commits that are part of the new artifact (that means at least until the commit of the previously built artifact). If you use shallow clone in your CI Kosli won't be able to generate the list but the command will NOT fail.

Fingerprint (sha256 checksum of the file/directory/docker image) of the artifact will be stored in Kosli. The fingerprint can't be changed, it becomes a unique identifier of the artifact in Kosli, used - among other things - to connect it with the recorded environment. Fingerprints of all the running artifacts, recorded with Kosli CLI are also stored and compared with fingerprints of the artifacts you have built and reported to Kosli so you always know if you're running things you have no provenance of. 

Some of the required flags will be automatically resolved if you're using one of the [supported CI systems](/getting_started/use_kosli_in_ci_systems/).

### Example 

```
# Report to a Kosli pipeline that a file type artifact has been created
kosli pipeline artifact report creation FILE.tgz \
	--api-token yourApiToken \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
kosli pipeline artifact report creation \
	--api-token yourApiToken \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 
  ```

See [kosli pipeline artifact report creation](/client_reference/kosli_pipeline_artifact_report_creation/) for more details. 

## Report evidences

Whenever an event related to one of your evidences happens you should report it to Kosli. 

Currently we support following types of evidences:

### Pull request

If you use GitHub or Bitbucket you can use Kosli to verify if the merge commit you used to build your artifact comes from a pull request. Remember to add the pull request evidence to your [pipeline template](/how_to/connect/#create-a-pipeline) and use the same label for `--evidence-type` you provided in a `template` 

If there is no pull request for given commit the evidence will be reported as incompliant and the pipeline will continue. You can choose to fail the pipeline altogether in case pull request is missing - use `--assert` flag to do that.

There are two different commands for that, depending on your CI.

For GitHub: [kosli pipeline artifact report evidence github-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_github-pullrequest/) along regular flags, you need to provide:
* `--github-org`
* `--github-token` your	Github personal access token.


For Bitbucket: [kosli pipeline artifact report evidence bitbucket-pullrequest](/client_reference/kosli_pipeline_artifact_report_evidence_bitbucket-pullrequest/) along regular flags, you need to provide:
*  `--bitbucket-password` - you need to use an api token which is the "App password" you create under "Personal Settings", keep in mind that api tokens you create under "Manage account" won't work for basic auth
* `--bitbucket-username` - you cannot user your email address you use to log in, you have an actual username under "Personal Settings") 
* `--bitbucket-workspace`


### JUnit test 

If you use JUnit to run your test you can use `kosli pipeline artifact report evidence test` command to analyze the results and provide it to Kosli. Remember to add the junit test evidence to your [pipeline template](/how_to/connect/#create-a-pipeline) and use the same label for `--evidence-type` you provided in a `template` 

### Example

```
# report a JUnit test evidence about a file artifact:
kosli pipeline artifact report evidence test FILE.tgz \
	--artifact-type file \
	--evidence-type yourEvidenceType \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults

# report a JUnit test evidence about an artifact using an available Sha256 digest:
kosli pipeline artifact report evidence test \
	--sha256 yourSha256 \
	--evidence-type yourEvidenceType \
	--pipeline yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--results-dir yourFolderWithJUnitResults
```

[junit test](/client_reference/kosli_pipeline_artifact_report_evidence_test/) 

### Generic

## Report approvals

Whenever a given artifact is ready to be deployed you may need an additional manual approval from authorized person. This is something that can't alway be automated, but you can use Kosli to request such an approval, and later record it, so the information about decisions made outside of your CI system won't be lost. The list of commits between current and previous approval will be generated, which allows you to track a set of changes that are being approved.

See [kosli pipeline approval report](/client_reference/kosli_pipeline_approval_report/) and [kosli pipeline approval request](/client_reference/kosli_pipeline_approval_request/) for more details. 

{{< hint warning >}}

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

{{< /hint >}}

## Deployments

The last important piece of information, when it comes to artifacts are deployments. Whenever you (likely with the use of your CI system) deploy an artifact to an environment you should record that information to Kosli. So when you check the status of your environments you know not only what is running there, but also how did it get there. It's an easy way of detecting a manual change was introduced.