---
title: "merkely environment report server"
---

## merkely environment report server

Report directory or file artifacts data in the given list of paths to Merkely.

### Synopsis


List the artifacts deployed in a server environment and their digests 
and report them to Merkely. 


```shell
merkely environment report server [-p /path/of/artifacts/directory] [-i infrastructure-identifier] env-name [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for server  |
|    -p, --paths strings  |  The comma separated list of artifact directories.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The merkely API token.  |
|    -c, --config-file string  |  [optional] The merkely config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The merkely endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The merkely user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

```shell

# report directory artifacts running in a server at a list of paths:
merkely environment report server yourEnvironmentName \
	--paths a/b/c, e/f/g \
	--api-token yourAPIToken \
	--owner yourOrgName  

```

