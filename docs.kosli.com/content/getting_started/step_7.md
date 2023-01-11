---
title: "Step 7: Report expected deployment of the artifact"
bookCollapseSection: false
weight: 290
---

# Step 7: Report expected deployment of the artifact

Before you run the nginx docker image (the artifact) on your docker host, you need to report 
to Kosli your intention of deploying that image. This allows Kosli to match what you 
expect to run in your environment with what is actually running, and flag any mismatches.  

```shell {.command}
kosli expect deployment nginx:1.21 \
    --pipeline quickstart-nginx \
    --artifact-type docker \
    --build-url https://example.com \
    --environment quickstart \
    --description "quickstart-nginx artifact deployed to quickstart env"
```

You can verify the deployment with:

```shell {.command}
kosli deployment ls quickstart-nginx
```

```plaintext {.light-console}
ID   ARTIFACT                                                                       ENVIRONMENT  REPORTED_AT
1    Name: nginx:1.21                                                               quickstart   Tue, 01 Nov 2022 15:48:47 CET
     Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514  
```

Now run the artifact:
```shell {.command}
docker-compose up -d
```

You can confirm the container is running:
```shell {.command}
docker ps
```
The output should include an entry similar to this:
```plaintext {.light-console}
CONTAINER ID  IMAGE      COMMAND                 CREATED         STATUS         PORTS                  NAMES
6330e545b532  nginx:1.21 "/docker-entrypoint.…"  35 seconds ago  Up 34 seconds  0.0.0.0:8080->80/tcp   quickstart-nginx
```
