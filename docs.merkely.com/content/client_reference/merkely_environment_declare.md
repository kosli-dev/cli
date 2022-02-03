---
title: "merkely environment declare"
---

## merkely environment declare

Declare or update a Merkely environment

### Synopsis


Declare or update a Merkely environment.


```shell
merkely environment declare [flags]
```

### Examples

```shell

* declare (or update) a Merkely environment:
merkely environment declare --api-token 1234 --owner test --name newEnv --environment-type K8S --description "my new env"

```

### Options

```
  -d, --description string        [optional] The environment description.
  -t, --environment-type string   The type of environment. Valid options are: [K8S, ECS, server]
  -h, --help                      help for declare
  -n, --name string               The name of environment.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely user or organization.
  -v, --verbose               Print verbose logs to stdout.
```

