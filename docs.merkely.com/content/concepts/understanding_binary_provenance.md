---
title: 'Understanding Binary Provenance'
weight: 4
---

## Understanding Binary Provenance

From a security and change management perspective,  the strongest form of trust in your production environment is to identify what is running and know where it came from.

This is important because all the controls, audit trails and tools are worth nothing if you can simply switch the binaries at any stage in your delivery process.

To implement binary provenance, there are two problems to solve:

1. How to uniquely identify your software
2. How to trace this identity to a provenance audit trail

Traditionally, software artifacts (docker images, zip files, etc) are identified using labels, such as, semantic versions, filenames, or other metadata.  We prefer to identify each software artifact by its binary content.

> **What is Content-Addressable Storage?**
>
> Content-Addressable Storage is a fancy phrase with a simple meaning: “the content of the binary determines its identity”.  Cryptographic hash functions such as the sha256 digest give you a one-way algorithm for identifying any collection of bytes.  This digest acts as a unique fingerprint.  
> Any change to the software package would result in a different fingerprint.

![Diagram of sha256 fingerprint](/images/fingerprint.png)

There are many advantages to this approach:

* Any tampering to the image produces a different fingerprint
* There is no need for specific naming or versioning
* There is no need for metadata to be tracked along with binaries

In addition, this approach significantly simplifies your devops pipelines.

Now that we can identify software, then we need to solve the second problem: how do we trace this identity to a provenance audit trail?  For this we need somewhere secure to store our evidence, be that a database, a git repository, or Merkely (which happens to be both 😇).
