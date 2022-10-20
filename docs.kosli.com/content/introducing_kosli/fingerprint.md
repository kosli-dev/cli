---
title: 'Fingerprint'
weight: 35
---
# Fingerprint

Every time artifact is reported to Kosli a SHA256 digest of it is calculated. It doesn't matter if the artifact is a single file, a directory or a docker image - we can always calculate SHA256. 

Fingerprints are used to connect the information recorded in Kosli - about environments, deployments and approval - to matching artifact. 

You can also use Kosli CLI to calculated the fingerprint of any artifact locally. See [kosli fingerprint](/client_reference/kosli_fingerprint/) for more details.