---
title: "kosli environment report k8s"
---

## kosli environment report k8s

Report images data from specific namespace(s) or entire cluster to Kosli.

### Synopsis


List the artifacts deployed in the k8s environment and their digests 
and report them to Kosli. 


```shell
kosli environment report k8s ENVIRONMENT-NAME [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -x, --exclude-namespace strings  |  [conditional] The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace.  |
|    -h, --help  |  help for k8s  |
|    -k, --kubeconfig string  |  [defaulted] The kubeconfig path for the target cluster. (default "$HOME/.kube/config")  |
|    -n, --namespace strings  |  [conditional] The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|    -D, --dry-run  |  [optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  [optional] Print verbose logs to stdout.  |


### Examples

```shell

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config 
(with global flags defined in environment or in a config file):
export KOSLI_API_TOKEN=yourAPIToken
export KOSLI_OWNER=yourOrgName

kosli environment report k8s yourEnvironmentName

# report what is running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
    --exclude-namespace kube-system,utilities \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
	--namespace your-namespace \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a cluster using kubeconfig at a custom path:
kosli environment report k8s yourEnvironmentName \
	--kubeconfig /path/to/kube/config \
	--api-token yourAPIToken \
	--owner yourOrgName

```

