---
title: "merkely"
---

## merkely

The Merkely evidence reporting CLI.

### Synopsis

The Merkely evidence reporting CLI.

Environment variables:
| Name                               | Description                                                                       |
|------------------------------------|-----------------------------------------------------------------------------------|
| $MERKELY_API_TOKEN                 | set the Merkely API token.                                                        |
| $MERKELY_OWNER                     | set the Merkely Pipeline Owner.                                                   |
| $MERKELY_HOST                      | set the Merkely host.                                                             |
| $MERKELY_DRY_RUN                   | indicate whether or not Merkely CLI is running in Dry Run mode.                   |
| $MERKELY_MAX_API_RETRIES           | set the maximum number of API calling retries when the API host is not reachable. |
| $MERKELY_CONFIG_FILE               | set the path to Merkely config file where you can set your options.               |


### Options

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -h, --help                  help for merkely
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely organization.
  -v, --verbose               Print verbose logs to stdout.
```

### SEE ALSO

* [merkely control](/client_reference/merkely_control/)	 - Check if artifact is allowed to be deployed.
* [merkely create](/client_reference/merkely_create/)	 - Create objects in Merkely.
* [merkely fingerprint](/client_reference/merkely_fingerprint/)	 - Print the SHA256 fingerprint of an artifact.
* [merkely report](/client_reference/merkely_report/)	 - Report compliance events to Merkely.
* [merkely version](/client_reference/merkely_version/)	 - Print the client version information

