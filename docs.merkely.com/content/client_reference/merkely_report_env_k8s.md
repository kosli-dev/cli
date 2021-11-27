---
title: "merkely report env k8s"
---

## merkely report env k8s

Report images data from specific namespace(s) or entire cluster to Merkely.

### Synopsis


List the artifacts deployed in the k8s environment and their digests
and report them to Merkely.


```
merkely report env k8s [-n namespace | -x namespace]... [-k /path/to/kube/config] [-i infrastructure-identifier] env-name [flags]
```

### Examples

```

* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config:
merkely report env k8s prod --api-token 1234 --owner exampleOrg --id prod-cluster

* report what's running in an entire cluster using kubeconfig at $HOME/.kube/config
(with global flags defined in environment or in  a config file):
merkely report env k8s prod

* report what's running in an entire cluster excluding some namespaces using kubeconfig at $HOME/.kube/config:
merkely report env k8s prod -x kube-system,utilities

* report what's running in a given namespace in the cluster using kubeconfig at $HOME/.kube/config:
merkely report env k8s prod -n prod-namespace

* report what's running in a cluster using kubeconfig at a custom path:
merkely report env k8s prod -k /path/to/kube/config

```

### Options

```
  -x, --exclude-namespace strings   The comma separated list of namespaces regex patterns NOT to report artifacts info from. Can't be used together with --namespace.
  -h, --help                        help for k8s
  -i, --id string                   The unique identifier of the source infrastructure of the report (e.g. the K8S cluster/namespace name). If not set, it is defaulted to environment name.
  -k, --kubeconfig string           The kubeconfig path for the target cluster. (default "$HOME/.kube/config")
  -n, --namespace strings           The comma separated list of namespaces regex patterns to report artifacts info from. Can't be used together with --exclude-namespace.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely organization.
  -v, --verbose               Print verbose logs to stdout.
```

### SEE ALSO

* [merkely report env](/client_reference/merkely_report_env/)	 - Report running artifacts in an environment to Merkely.

