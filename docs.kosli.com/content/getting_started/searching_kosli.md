---
title: "Searching Kosli"
bookCollapseSection: false
weight: 292
---

# Searching Kosli

Now that you have reported your artifact and what's running in our runtime environment,
you can use the `kosli search` command to find everything Kosli knows about an artifact or a git commit.

For example, you can give Kosli search the git commit SHA which you used when you reported the artifact: 

```shell {.command}
kosli search 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
```

```plaintext {.light-console}
Search result resolved to commit 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Name:              nginx:1.21
Fingerprint:       2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514
Has provenance:    true
Pipeline:          quickstart-nginx
Git commit:        9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Commit URL:        https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
Build URL:         https://example.com
Compliance state:  COMPLIANT
History:
    Artifact created                             Tue, 01 Nov 2022 15:46:59 CET
    Deployment #1 to quickstart environment      Tue, 01 Nov 2022 15:48:47 CET
    Started running in quickstart#1 environment  Tue, 01 Nov 2022 15:55:49 CET
```

Visit [Search](/how_to/search/) section to learn more