---
title: "merkely environment report k8s"
---

## merkely environment report k8s

Report images data from specific namespace(s) or entire cluster to Merkely.

### Synopsis


List the artifacts deployed in the k8s environment and their digests 
and report them to Merkely. 


```shell
merkely environment report k8s [-n namespace | -x namespace]... [-k /path/to/kube/config] [-i infrastructure-identifier] env-name [flags]
```

### Examples

```shell

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
merkely environment report k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config 
(with global flags defined in environment or in a config file):
export MERKELY_API_TOKEN=yourAPIToken
export MERKELY_OWNER=yourOrgName

merkely environment report k8s yourEnvironmentName

# report what is running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
merkely environment report k8s yourEnvironmentName \
    --exclude-namespace kube-system,utilities \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
merkely environment report k8s yourEnvironmentName \
	--namespace your-namespace \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a cluster using kubeconfig at a custom path:
merkely environment report k8s yourEnvironmentName \
	--kubeconfig /path/to/kube/config \
	--api-token yourAPIToken \
	--owner yourOrgName

```

### Options

```
  -x, --exclude-namespace strings   The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace.
  -h, --help                        help for k8s
  -k, --kubeconfig string           The kubeconfig path for the target cluster. (default "$HOME/.kube/config")
  -n, --namespace strings           The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely user or organization.
  -v, --verbose               Print verbose logs to stdout.
```

