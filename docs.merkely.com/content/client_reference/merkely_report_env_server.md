---
title: "merkely report env server"
---

## merkely report env server

Report directory artifacts data in the given list of paths to Merkely.

### Synopsis


List the artifacts deployed in a server environment and their digests
and report them to Merkely.


```
merkely report env server [-p /path/of/artifacts/directory] [-i infrastructure-identifier] env-name [flags]
```

### Examples

```

* report directory artifacts running in a server at a list of paths:
merkely report env server prod --api-token 1234 --owner exampleOrg --id prod-server --paths a/b/c, e/f/g

```

### Options

```
  -h, --help            help for server
  -i, --id string       The unique identifier of the source infrastructure of the report (e.g. the K8S cluster/namespace name). If not set, it is defaulted to environment name.
  -p, --paths strings   The comma separated list of artifact directories.
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

