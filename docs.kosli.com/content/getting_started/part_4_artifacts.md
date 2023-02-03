---
title: "Part 4: Artifacts"
bookCollapseSection: false
weight: 240
---
# Part 4: Artifacts

## Report artifacts

To report an artifact to Kosli, you need its SHA256 fingerprint. You can either provide the fingerprint yourself, or let Kosli CLI calculate it for you. 
You also need to provide the name of the Kosli pipeline you want to report the artifact to.

You also should have long enough git history in your local git repo clone to let Kosli calculate the artifact's changelog (the list of commits from the new artifact back to the previous artifact in the same Kosli pipeline).
If you use shallow clone in your CI, Kosli won't be able to generate the changelog but the artifact reporting will NOT fail. Kosli collects the changelog commits on best-effort basis.

The fingerprint (sha256 checksum of the file/directory/docker image) of the artifact will be stored in Kosli. The fingerprint can't be changed, it becomes a unique identifier of the artifact in Kosli, used - among other things - to connect it with the recorded environment. Fingerprints of all the running artifacts, recorded with the Kosli CLI are also stored and compared with fingerprints of the artifacts you have built and reported to Kosli so you always know if you're running things you have no provenance for. 

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

