# Azure Service Apps and Function Apps reporting

Azure API/SDK and portal do not provide SHA256 image digests for docker containers.
According to [Azure support](https://learn.microsoft.com/en-us/answers/questions/1366756/how-do-you-find-the-sha256-digest-of-a-running-app#comment-1371459),

`"Yes as far as I can see App service doesn't store the digest anywhere other than the docker logs that are accessible to you - unless you use the image hash value as part of the identifier."`

Thus, to get the SHA256 digest of a running container inside of a Service and Function app, we use the algorithm below.

## Algorithm

### Pre-requisites

To use Azure CLI, you need to have Azure CLI installed and logged in to your Azure account:

```bash
az login
```

To list accounts that you are logged into:

```bash
az account list --all
```

Accounts are only refreshed when you login, so if you have recently added a new subscription, you need to login again.

Get a list of resource groups in a subscription.
CLI command:

```bash
az group list --subscription <subscription_id | subscription_name>
```

### Get a list of apps in a resource group of a subscription.

CLI command:

```bash
# To get a list of web apps
az webapp list --resource-group <YourResourceGroupId> --subscription <YourSubscriptionId>
# To get a list of function apps
az functionapp list --resource-group <YourResourceGroupId> --subscription <YourSubscriptionId>
```

You will get an output similar to this:
```json
[
    {
        ...
        "name": "api-service", # app name
        ...
        "siteConfig": {
            ...
            "linuxFxVersion": "DOCKER|tookyregistry.azurecr.io/tookyregistry/tooky/api-image:3d346858a44df6820eaef8195008459f979f0526",
            ...
        },
        ...
        "state": "Running",
        ...
    }
]
```

### Check the state of an Azure app
Only apps with state "Running" are considered for reporting, the rest are ignored.
Only apps with `linuxFxVersion starting with "DOCKER|"` are considered for reporting, the rest are ignored.

### Get docker image name and tag of an Azure app
Docker image name and tag are extracted from `linuxFxVersion`.
For example, if `linuxFxVersion` is `DOCKER|tookyregistry.azurecr.io/tookyregistry/tooky/api-image:3d346858a44df6820eaef8195008459f979f0526`,
then the docker image name is `tookyregistry.azurecr.io/tookyregistry/tooky/api-image`
and the image tag is `3d346858a44df6820eaef8195008459f979f0526`.

### Get WebSite container logs for a running Azure app

CLI command:

```bash
az webapp log download --name <YourAppName> --resource-group <YourResourceGroupId> --subscription <YourSubscriptionId>
```

If successfull it will display
```bash
Downloaded logs to webapp_logs.zip
```

Extract the logs
```bash
unzip webapp_logs.zip
```

Find the last docker log.
```bash
ls -trl LogFiles/*docker.log
...
LogFiles/2023_09_28_10-30-0-141_docker.log
```


### Get docker information from log file

The log file will contain one or more of these blocks
```bash
...
2023-09-28T12:27:30.909Z INFO  - 3a9444c255ce Extracting 1KB / 1KB
2023-09-28T12:27:31.086Z INFO  - 3a9444c255ce Pull complete
2023-09-28T12:27:31.201Z INFO  -  Digest: sha256:1b7c84fc8a533a34ed6e8553976c6b68d97adaa1dbe6499265e7a76ac75801d4
2023-09-28T12:27:31.250Z INFO  -  Status: Downloaded newer image for tookyregistry.azurecr.io/tookyregistry/tooky/api-image6@sha256:1b7c84fc8a533a34ed6e8553976c6b68d97adaa1dbe6499265e7a76ac75801d4
2023-09-28T12:27:31.282Z INFO  - Pull Image successful, Time taken: 1 Minutes and 8 Seconds
2023-09-28T12:27:33.104Z INFO  - Starting container for site
2023-09-28T12:27:33.104Z INFO  - docker run -d -p 6693:3000 --name api-service_0_5b07493a -e WEBSITES_ENABLE_APP_SERVICE_STORAGE=false -e WEBSITES_PORT=3000 -e WEBSITE_SITE_NAME=api-service -e WEBSITE_AUTH_ENABLED=False -e WEBSITE_ROLE_INSTANCE_ID=0 -e WEBSITE_HOSTNAME=api-service.azurewebsites.net -e WEBSITE_INSTANCE_ID=e3848c4a19ed5120ac06e6c4552adf58a74475463871a23ea40c8f269e489576 -e HTTP_LOGGING_ENABLED=1 -e WEBSITE_USE_DIAGNOSTIC_SERVER=False tookyregistry.azurecr.io/tookyregistry/tooky/api-image@sha256:1b7c84fc8a533a34ed6e8553976c6b68d97adaa1dbe6499265e7a76ac75801d4  

2023-09-28T12:27:36.389Z INFO  - Initiating warmup request to container api-service_0_5b07493a for site api-service
2023-09-28T12:27:37.414Z INFO  - Container api-service_0_5b07493a for site api-service initialized successfully and is ready to serve requests.
...
```

The import lines are
```bash
2023-09-28T12:27:31.201Z INFO  -  Digest: sha256:1b7c84fc8a533a34ed6e8553976c6b68d97adaa1dbe6499265e7a76ac75801d4
2023-09-28T12:27:37.414Z INFO  - Container api-service_0_5b07493a for site api-service initialized successfully and is ready to serve requests.
```

The second line informs us that a container started successfully. When we have that we know that the previous `Digest: sha256`
line was the sha256 of this container.

### Which `Digest:` line to trust

The docker log is not written by the platform alone. It also carries the container's stdout and stderr, so a
container can print a line containing `Digest: sha256:...` and it lands in the same text. A container that does
this after it starts must not be able to choose the fingerprint the snapshot reports
([kosli-dev/server#6881](https://github.com/kosli-dev/server/issues/6881)).

Three things separate the platform's lines from the container's:

1. **The start line names the site and the image.** The platform's `docker run` line is the start marker. It
   names the site (`--name <site>_...` and `-e WEBSITE_SITE_NAME=<site>`), so only this site's starts count,
   and it names the reference the platform ran. When the configured repository is pinned there
   (`<repository>@sha256:<digest>`, as in the sample above) the digest is read from it. Start proof and digest
   are then the same platform line, so nothing written earlier in the log can stand in for it. This also
   covers a restart with no pull in the window. Only the configured repository is accepted, because a
   configured startup command follows the image on the same line.
2. **Order.** The container's output can only appear after the platform's `docker run` line. When the run line
   names a tag instead, the digest is the last `Digest:` line *before* the last container start, and anything
   after the start is ignored. The latest start wins, because the log window can hold several deployments and
   the running container is the most recent one. Taking the first match instead would report a stale
   deployment.
3. **Shape.** Platform lines begin with a UTC timestamp carrying exactly three fractional digits, then a level
   and a dash (`2023-09-28T12:27:31.201Z INFO  - ...`). Container output is timestamped by Docker at nanosecond
   precision and gets no level or dash (`2023-09-28T12:27:35.123456789Z ...`). Only lines of the platform's
   shape are considered. This is what stops a container that is being replaced, and is still running while the
   platform pulls its successor, from writing a fake digest into the gap between the pull and the start when
   the run line names a tag.

Nothing in the log can fail the snapshot. Lines are walked without a length cap, so a container writing an
arbitrarily long line cannot abort reporting, and a line that matches a platform pattern in shape but not in
content (a timestamp that is not an instant, a digest token of the wrong length) is skipped. When no platform
digest can be found the app is reported without a fingerprint and a warning names it.

#### Assumptions about the platform's log

Every logs-mode app now depends on these, and each fails closed to "no fingerprint plus a warning" rather than
to a wrong fingerprint:

1. The platform logs a `docker run` line for every container start.
2. That line names the site as `--name <site>_<instance>_...` or `-e WEBSITE_SITE_NAME=<site>`, compared
   ignoring case since app names are hostnames.
3. Platform lines carry the `<timestamp with three fractional digits> <LEVEL>  - ` prefix, and container output
   does not.
4. When the platform pins the image by digest on the run line, it spells the repository the same way the app's
   configuration does, up to Docker Hub short-form expansion and registry host case. The line is logged
   unquoted and split on whitespace, so an app setting value or startup command can produce a token that also
   names that repository; the run line is trusted only when every such token names the same digest, a tag
   counting as a competing claim, and the pulled digest is used otherwise.

They hold for the Linux Web App capture above. They have **not** been verified against a containerised Function
App or a Windows-container app; a capture from either is the way to close that.

None of this makes the log a source of truth: the registry is. `--digests-source acr` is the default, and the CLI
prints a warning whenever `--digests-source logs` is used.


## Findings

A script was periodically executed over a period of time. It involved multiple function apps operating as Docker containers and several servers in the setup:

1. Retention period for the logs are 10 days. One of the function app had a shorter retention period. Don't know why.
2. Deployments and errors (which leads to a restart of the container) generates sha256 events in the log.
3. Scaling events does not end up in the logs.
4. If a function app has had no events for 10 days, it is no longer possible to get the sha256 from the log.
5. When we started to run this script we could get the sha256 from most of the function apps in most of the environments. But not for all.
6. Docker restart with `az functionapp restart` triggers a new pull, restart of the container and it makes an entry in the log.


<!-- 
Notes:
az webapp log deployment list --name arstan-service --resource-group KosliExperiment \
    --subscription 1f4973e6-11b3-4259-be2f-92bd3fe0a5cf

az webapp log deployment list --name tsha256  --resource-group EnvironmentReportingExperiment \
 --subscription 96cdee58-1fa8-419d-a65a-7233b3465632
 -->
