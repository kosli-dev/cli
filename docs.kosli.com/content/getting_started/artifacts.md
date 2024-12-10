---
title: "Part 6: Artifacts"
bookCollapseSection: false
weight: 260
---
# Part 6: Artifacts

In software processes, you typically generate one or more artifacts that are deployed or distributed, such as docker images, archives, binaries, etc. You can ensure traceability for the creation of these artifacts by attesting them to Kosli, thereby establishing a binary provenance for each one.

## Binary provenance

Binary provenance for artifacts refers to the ability to trace and verify the origins, history, and journey of the artifacts throughout their lifecycle. This involves recording immutable attestations about the artifact creation, risk controls performed on it, deployments, and execution/usage.

Artifacts are uniquely identified by their SHA256 fingerprints. When attesting an artifact to Kosli, you have the option to either provide the fingerprint manually or allow Kosli CLI to calculate it automatically for you.

By leveraging the artifact's fingerprint, Kosli can establish connections between the creation of the artifact and its runtime-related events, such as when the artifact starts or ceases execution within a specific environment.

By establishing and maintaining binary provenance for artifacts, Kosli enables you to:

1. **Track Changes**: Trace how your Flow artifacts change over time.
2. **Identify Sources**: Understand where your artifacts originated from, which can help in identifying vulnerabilities or issues.
3. **Monitor Compliance**: Ensure that the artifacts adhere to your compliance requirements.
4. **Enable Audits**: Access audit packages on demand allowing audits and investigations into the software supply chain.
5. **Enhance Trust**: Build trust among users, customers, and stakeholders by providing transparent and verified information about the software's history.

## Attesting artifacts

To attest an artifact, you can run a command similar to the one below:

```shell
$ kosli attest artifact project-a-app.bin \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/ProjectA/ProjectAApp/commit/e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--git-commit e67f2f2b121f9325ebf166b7b3c707f73cb48b14 \
	--flow project-a \
	--trail trail-1 \
	--name backend
```
See [kosli attest artifact](/client_reference/kosli_attest_artifact/) for more details. 


## The --dry-run flag

All Kosli CLI commands accept the `--dry-run` [boolean flag](/faq/#boolean-flags). 
When this flag is used, a CLI command:
* Does not communicate with Kosli at all
* Prints the payload it would have sent
* Exits with a zero status code

We recommend using the `KOSLI_DRY_RUN` environment variable to automatically set the `--dry-run` flag. 
This will allow you to instantly turn off all Kosli CLI commands if Kosli is down, as detailed in
[this tutorial](/tutorials/what_do_i_do_if_kosli_is_down/).

The `--dry-run` flag is also useful when trying commands locally. For example:

```shell
$ kosli attest artifact cyberdojo/differ:dde3b2a \
  --artifact-type=docker \
  --org=cyber-dojo \
  --flow=differ-ci \
  --trail=$(git rev-parse HEAD) \
  --dry-run \
  ...

 {
    "fingerprint": "0f53b5b9e7c266defe6984deafe039b116295b2df4a409ba6288c403f2451a9f",
    "filename": "cyberdojo/differ:dde3b2a",
    "git_commit": "fbb9e8000e2344323040e348a54b33ecbf67f273",
    "git_commit_info": {
        "sha1": "fbb9e8000e2344323040e348a54b33ecbf67f273",
        "message": "improve coverage report info (#2796)",
        "author": "Jon Jagger \u003cjon@jaggersoft.com\u003e",
        "timestamp": 1733724563,
        "branch": "master",
        "url": "https://github.com/kosli-dev/server/commit/fbb9e8000e2344323040e348a54b33ecbf67f273"
    },
    "build_url": "https://github.com/cyber-dojo/differ/actions/runs/11777650898",
    "commit_url": "https://github.com/cyber-dojo/differ/commit/dde3b2a7dab8e4567038e4c66ac68f0f01d0f704",
    "repo_url": "https://github.com/kosli-dev/server",
    "template_reference_name": "differ",
    "trail_name": "dde3b2a7dab8e4567038e4c66ac68f0f01d0f704"
}

$ echo $?
0
```

