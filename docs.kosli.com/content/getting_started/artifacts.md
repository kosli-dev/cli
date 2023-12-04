---
title: "Part 5: Artifacts"
bookCollapseSection: false
weight: 240
---
# Part 5: Artifacts

## Report artifacts

To report an artifact to Kosli, you need its SHA256 fingerprint. You can either provide the fingerprint yourself, or let Kosli CLI calculate it for you - we'll need the artifact available while running reporting command to do that. 
You also need to provide the name of the Kosli flow you want to report the artifact to.

You should also have long enough git history in your local git repo clone to let Kosli calculate the artifact's changelog (the list of commits from the new artifact back to the previous artifact reported to the same Kosli flow).  
If you use shallow clone in your CI, Kosli won't be able to generate the changelog but the artifact reporting will **NOT** fail. Kosli collects the changelog commits on best-effort basis.

The fingerprint (SHA256 checksum of the file/directory/docker image) of the artifact will be stored in Kosli. The fingerprint can't be changed, it becomes a unique identifier of the artifact in Kosli, used - among other things - to connect it with the data reported from your runtime environments. 

Fingerprints of all the **running** artifacts, recorded with the Kosli CLI are also stored and **compared with** fingerprints of the artifacts you have built and **reported** to Kosli so you always know if you're running things you have built or if you have no provenance for them. 

Some of the required flags will be automatically resolved if you're using one of the [supported CI systems](/integrations/ci_cd/).

### Example 

{{< tabs "commands" "col-no-wrap" >}}

{{< tab "v2" >}}
```
$ kosli report artifact project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/ProjectA/ProjectAApp/commit/e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--git-commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--flow project-a 

artifact project-a-app.bin was reported with fingerprint: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli report artifact](/client_reference/kosli_report_artifact/) for more details. 

{{< /tab >}}

{{< tab "v0.1.x" >}}
```
$ kosli pipeline artifact report creation project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/ProjectA/ProjectAApp/commit/e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--git-commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--pipeline project-a 

artifact project-a-app.bin was reported with fingerprint: 53c97572093cc107c0caa2906d460ccd65083a4c626f68689e57aafa34b14cbf
```
See [kosli pipeline artifact report creation](/legacy_ref/v0.1.41/kosli_pipeline_artifact_report_creation/) for more details. 

{{< /tab >}}

{{< /tabs >}}



