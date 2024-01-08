---
title: "Part 6: Artifacts"
bookCollapseSection: false
weight: 260
---
# Part 6: Artifacts

In software processes, you typically generate one or more artifacts that are deployed or distributed, such as docker images, archives, binaries, and more. You can ensure traceability for the creation of these artifacts by attesting them to Kosli, thereby establishing an binary provenance for each one.

## Binary provenance

Binary provenance for artifacts refers to the ability to trace and authenticate the origins, history, and journey of the artifacts throughout their lifecycle. This involves recording immutable attestations about the artifact creation, risk controls performed on it, distribution and execution/usage.

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



