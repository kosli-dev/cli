---
title: "kosli snapshot k8s"
beta: false
deprecated: false
summary: "Report a snapshot of running pods in a K8S cluster or namespace(s) to Kosli.  "
---

# kosli snapshot k8s

## Synopsis

```shell
kosli snapshot k8s ENVIRONMENT-NAME [flags]
```

Report a snapshot of running pods in a K8S cluster or namespace(s) to Kosli.  
Skip `--namespaces` and `--namespaces-regex` to report all pods in all namespaces in a cluster.
The reported data includes pod container images digests and creation timestamps. You can customize the scope of reporting
to include or exclude namespaces.

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude-namespaces strings  |  [optional] The comma separated list of namespaces names to exclude from reporting artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --namespaces or --namespaces-regex.  |
|        --exclude-namespaces-regex strings  |  [optional] The comma separated list of namespaces regex patterns to exclude from reporting artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --namespaces or --namespaces-regex.  |
|    -h, --help  |  help for k8s  |
|    -k, --kubeconfig string  |  [defaulted] The kubeconfig path for the target cluster. (default "$HOME/.kube/config")  |
|    -n, --namespaces strings  |  [optional] The comma separated list of namespaces names to report artifacts info from. Can't be used together with --exclude-namespaces or --exclude-namespaces-regex.  |
|        --namespaces-regex strings  |  [optional] The comma separated list of namespaces regex patterns to report artifacts info from. Requires cluster-wide read permissions for pods and namespaces. Can't be used together with --exclude-namespaces --exclude-namespaces-regex.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

**report what is running in an entire cluster using kubeconfig at $HOME/.kube/config**

```shell
kosli snapshot k8s yourEnvironmentName 

```

**report what is running in an entire cluster using kubeconfig at $HOME/.kube/config**

```shell
(with global flags defined in environment or in a config file):
export KOSLI_API_TOKEN=yourAPIToken
export KOSLI_ORG=yourOrgName

kosli snapshot k8s yourEnvironmentName

```

**report what is running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config**

```shell
kosli snapshot k8s yourEnvironmentName 
    --exclude-namespaces kube-system,utilities 

```

**report what is running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config**

```shell
kosli snapshot k8s yourEnvironmentName 
	--namespaces your-namespace 

```

**report what is running in a cluster using kubeconfig at a custom path**

```shell
kosli snapshot k8s yourEnvironmentName 
	--kubeconfig /path/to/kube/config 
```

