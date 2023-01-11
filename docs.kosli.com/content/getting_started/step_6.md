---
title: "Step 6: Report artifacts to Kosli"
bookCollapseSection: false
weight: 280
---

# Step 6: Report artifacts to Kosli

Typically, you would build an artifact in your CI system. 
The quickstart-docker repository contains a `docker-compose.yml` file which uses an [nginx](https://nginx.org/) docker image 
which you will be using as your artifact in this tutorial instead.

Pull the docker image - Kosli CLI needs the artifact to be locally present to 
generate a "fingerprint" to identify it:

```shell {.command}
docker-compose pull
```

You can check that this has worked by typing: 
```shell {.command}
docker images nginx
```
The output should look like this:
```plaintext {.light-console}
REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        1.21      8f05d7383593   5 months ago   134MB
```

Now you can report the artifact to Kosli. 
This tutorial uses a dummy value for the `--build-url` flag, in a real installation 
this would be a link to a build service (e.g. Github Actions).

```shell {.command}
kosli pipeline artifact report creation nginx:1.21 \
    --pipeline quickstart-nginx \
    --artifact-type docker \
    --build-url https://example.com \
    --commit-url https://github.com/kosli-dev/quickstart-docker-example/commit/9f14efa0c91807da9a8b1d1d6332c5b3aa24a310 \
    --git-commit 9f14efa0c91807da9a8b1d1d6332c5b3aa24a310
```

You can verify that you have reported the artifact in your *quickstart-nginx* pipeline:

```shell {.command}
kosli artifact ls quickstart-nginx
```

```plaintext {.light-console}
COMMIT   ARTIFACT                                                                       STATE      CREATED_AT
9f14efa  Name: nginx:1.21                                                               COMPLIANT  Tue, 01 Nov 2022 15:46:59 CET
         Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514                 
```
