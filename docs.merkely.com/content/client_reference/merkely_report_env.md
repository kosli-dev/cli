---
title: "merkely report env"
---

## merkely report env

Report running artifacts in an environment to Merkely.

### Synopsis


Report actual deployments in an environment back to Merkely.
This allows Merkely to determine Runtime compliance status of the environment.


### Options

```
  -h, --help   help for env
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

* [merkely report](/client_reference/merkely_report/)	 - Report compliance events to Merkely.
* [merkely report env ecs](/client_reference/merkely_report_env_ecs/)	 - Report images data from AWS ECS cluster to Merkely.
* [merkely report env k8s](/client_reference/merkely_report_env_k8s/)	 - Report images data from specific namespace(s) or entire cluster to Merkely.
* [merkely report env server](/client_reference/merkely_report_env_server/)	 - Report directory artifacts data in the given list of paths to Merkely.

