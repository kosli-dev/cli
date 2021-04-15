# Watcher (temporary working name)

This CLI is used to report the images of workloads running in a k8s cluster back to merkely.


## Usage 

```
harvest pod image data from specific namespace or entire cluster

Usage:
  merkely [flags]

Flags:
  -a, --api-token string            the merkely API token.
  -d, --dry-run                     whether to send the request to the endpoint or just log it in stdout.
  -e, --environment string          the name of the merkely environment.
  -x, --exclude-namespace strings   the comma separated list of namespaces NOT to harvest artifacts info from. Can't be used together with --namespace.
  -h, --help                        help for merkely
  -H, --host string                 the merkely endpoint. (default "https://app.merkely.com")
  -k, --kubeconfig string           kubeconfig path for the target cluster (default "/Users/samialajrami/.kube/config")
  -n, --namespace strings           the comma separated list of namespaces to harvest artifacts info from. Can't be used together with --exclude-namespace.
  -o, --owner string                the merkely organization.
```

## Linting the code

`make lint`


## Building the code

`make build`

## Testing the code

`make test`

## Building the docker image

`make docker`