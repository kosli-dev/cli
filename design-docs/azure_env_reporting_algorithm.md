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
Only apps with linuxFxVersion starting with "DOCKER|" are considered for reporting, the rest are ignored.

### Get docker image name and tag of an Azure app
Docker image name and tag are extracted from linuxFxVersion.
For example, if linuxFxVersion is "DOCKER|tookyregistry.azurecr.io/tookyregistry/tooky/api-image:3d346858a44df6820eaef8195008459f979f0526",
then the docker image name is "tookyregistry.azurecr.io/tookyregistry/tooky/api-image" 
and the image tag is "3d346858a44df6820eaef8195008459f979f0526".

### Get WebSite container logs for a running Azure app

CLI command:

```bash
az webapp log download --name <YourAppName> --resource-group <YourResourceGroupId> --subscription <YourSubscriptionId>

az webapp log deployment list --name arstan-service --resource-group KosliExperiment --subscription 1f4973e6-11b3-4259-be2f-92bd3fe0a5cf

```

If successfull it will display
```bash
Downloaded logs to webapp_logs.zip
``

