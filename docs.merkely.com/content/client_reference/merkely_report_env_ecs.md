---
title: "merkely report env ecs"
---

## merkely report env ecs

Report images data from AWS ECS cluster to Merkely.

### Synopsis


List the artifacts deployed in an AWS ECS cluster and their digests
and report them to Merkely.


```
merkely report env ecs env-name [flags]
```

### Examples

```

* report what's running in an entire AWS ECS cluster:
merkely report env ecs prod --api-token 1234 --owner exampleOrg

```

### Options

```
  -C, --cluster string        The name of the ECS cluster.
  -h, --help                  help for ecs
  -i, --id string             The unique identifier of the source infrastructure of the report (e.g. the ECS cluster/service name).If not set, it is defaulted based on the following order: --service-name, --cluster, environment name.
  -s, --service-name string   The name of the ECS service.
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

