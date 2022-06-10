---
title: "kosli environment report server"
---

## kosli environment report server

Report directory or file artifacts data in the given list of paths to Kosli.

### Synopsis


List the artifacts deployed in a server environment and their digests 
and report them to Kosli. 


```shell
kosli environment report server [-p /path/of/artifacts/directory] [-i infrastructure-identifier] env-name [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for server  |
|    -p, --paths strings  |  The comma separated list of artifact directories.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

```shell

# report directory artifacts running in a server at a list of paths:
kosli environment report server yourEnvironmentName \
	--paths a/b/c, e/f/g \
	--api-token yourAPIToken \
	--owner yourOrgName  

```

