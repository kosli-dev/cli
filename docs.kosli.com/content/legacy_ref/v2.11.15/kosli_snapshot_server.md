---
title: "kosli snapshot server"
beta: false
deprecated: true
summary: "Report a snapshot of artifacts running in a server environment to Kosli.  "
---

# kosli snapshot server

{{% hint danger %}}
**kosli snapshot server** is deprecated. use 'kosli snapshot paths' instead  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report a snapshot of artifacts running in a server environment to Kosli.  
You can report directory or file artifacts in one or more server paths.

When fingerprinting a 'dir' artifact, you can exclude certain paths from fingerprint calculation 
using the `--exclude` flag.
Excluded paths are relative to the DIR-PATH and can be literal paths or glob patterns.
With a directory structure like this `foo/bar/zam/file.txt` if you are calculating the fingerprint of `foo/bar` you need to
exclude `zam/file.txt` which is relative to the DIR-PATH.
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

```shell
kosli snapshot server ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma separated list of directories and files to exclude from fingerprinting. Can take glob patterns.  |
|    -h, --help  |  help for server  |
|    -p, --paths strings  |  The comma separated list of absolute or relative paths of artifact directories or files. Can take glob patterns, but be aware that each matching path will be reported as an artifact.  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

```shell
# report directory artifacts running in a server at a list of paths:
kosli snapshot server yourEnvironmentName \
	--paths a/b/c,e/f/g \
	--api-token yourAPIToken \
	--org yourOrgName  
	
# exclude certain paths when reporting directory artifacts: 
# in the example below, any path matching [a/b/c/logs, a/b/c/*/logs, a/b/c/*/*/logs]
# will be skipped when calculating the fingerprint
kosli snapshot server yourEnvironmentName \
	--paths a/b/c \
	--exclude logs,"*/logs","*/*/logs"
	--api-token yourAPIToken \
	--org yourOrgName 
	
# use glob pattern to match paths to report them as directory artifacts: 
# in the example below, any path matching "*/*/src" under top-dir/ will be reported as a separate artifact.
kosli snapshot server yourEnvironmentName \
	--paths "top-dir/*/*/src" \
	--api-token yourAPIToken \
	--org yourOrgName
```

