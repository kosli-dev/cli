---
title: "merkely create environment"
---

## merkely create environment

Create a Merkely environment

### Synopsis


Create a Merkely environment.


```
merkely create environment [flags]
```

### Examples

```

* create a Merkely environment:
merkely create environment --api-token 1234 --owner test --name newEnv --type K8S --description "my new env"

```

### Options

```
  -d, --description string   [optional] The environment description.
  -h, --help                 help for environment
  -n, --name string          The name of environment.
  -t, --type string          The type of environment. Valid options are: [K8S, ECS, server]
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

* [merkely create](/client_reference/merkely_create/)	 - Create objects in Merkely.

