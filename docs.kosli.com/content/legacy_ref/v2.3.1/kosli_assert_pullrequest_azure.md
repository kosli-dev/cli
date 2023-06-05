---
title: "kosli assert pullrequest azure"
experimental: false
---

# kosli assert pullrequest azure

## Synopsis

Assert a Azure DevOps pull request for a git commit exists.
The command exits with non-zero exit code 
if no pull requests were found for the commit.

```shell
kosli assert pullrequest azure [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --azure-org-url string  |  Azure organization url. E.g. "https://dev.azure.com/myOrg" (defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --azure-token string  |  Azure Personal Access token.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for azure  |
|        --project string  |  Azure project.(defaulted if you are running in Azure Devops pipelines: https://docs.kosli.com/ci-defaults ).  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

kosli assert pullrequest azure \
	--azure-token yourAzureToken \
	--azure-org-url yourAzureOrgUrl \
	--commit yourGitCommit \
	--project yourAzureDevopsProject \
	--repository yourAzureDevOpsGitRepository

```

