---
title: "kosli snapshot azure"
beta: false
deprecated: false
---

# kosli snapshot azure

## Synopsis

Report a snapshot of running Azure web apps and function apps in an Azure resource group to Kosli.  
The reported data includes Azure app names, container image digests and creation timestamps.

To authenticate to Azure, you need to create Azure service principal with a secret  
and provide these Azure credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AZURE_CLIENT_ID).  
The service principal needs to have the following permissions:  
  1) Microsoft.Web/sites/Read  
  2) Microsoft.ContainerRegistry/registries/pull/read  

	

```shell
kosli snapshot azure ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --azure-client-id string  |  Azure client ID.  |
|        --azure-client-secret string  |  Azure client secret.  |
|        --azure-resource-group-name string  |  Azure resource group name.  |
|        --azure-subscription-id string  |  Azure subscription ID.  |
|        --azure-tenant-id string  |  Azure tenant ID.  |
|        --digests-source string  |  [defaulted] Where to get the digests from. Valid values are 'acr' and 'logs'. (default "acr")  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for azure  |
|        --zip  |  Download logs from Azure as zip files  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples Use Cases

**Use Azure Container Registry to get the digests for artifacts in a snapshot**

```shell
kosli snapshot azure yourEnvironmentName \
	--azure-client-id yourAzureClientID \
	--azure-client-secret yourAzureClientSecret \
	--azure-tenant-id yourAzureTenantID \
	--azure-subscription-id yourAzureSubscriptionID \
	--azure-resource-group-name yourAzureResourceGroupName \
	--digests-source acr \
	--api-token yourAPIToken \
	--org yourOrgName

```

**Use Docker logs of Azure apps to get the digests for artifacts in a snapshot**

```shell
kosli snapshot azure yourEnvironmentName \
	--azure-client-id yourAzureClientID \
	--azure-client-secret yourAzureClientSecret \
	--azure-tenant-id yourAzureTenantID \
	--azure-subscription-id yourAzureSubscriptionID \
	--azure-resource-group-name yourAzureResourceGroupName \
	--digests-source logs \
	--api-token yourAPIToken \
	--org yourOrgName

```

**Report digest of an Azure Function app**

```shell
kosli snapshot azure yourEnvironmentName \
	--azure-client-id yourAzureClientID \
	--azure-client-secret yourAzureClientSecret \
	--azure-tenant-id yourAzureTenantID \
	--azure-subscription-id yourAzureSubscriptionID \
	--azure-resource-group-name yourAzureResourceGroupName \
	--api-token yourAPIToken \
	--org yourOrgName
```

