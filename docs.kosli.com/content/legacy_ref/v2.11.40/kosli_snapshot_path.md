---
title: "kosli snapshot path"
beta: false
deprecated: false
summary: "Report a snapshot of a single artifact running in a specific filesystem path to Kosli.  "
---

# kosli snapshot path

## Synopsis

```shell
kosli snapshot path ENVIRONMENT-NAME [flags]
```

Report a snapshot of a single artifact running in a specific filesystem path to Kosli.  
You can report a directory or file artifact. For reporting multiple artifacts in one go, use "kosli snapshot paths".
You can exclude certain paths or patterns from the artifact fingerprint using `--exclude`.
The supported glob pattern syntax is documented here: https://pkg.go.dev/path/filepath#Match ,
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file, unless it is explicitly ignored itself.

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma-separated list of literal paths or glob patterns to exclude when fingerprinting the artifact.  |
|    -h, --help  |  help for path  |
|        --name string  |  The reported name of the artifact.  |
|        --path string  |  The base path for the artifact to snapshot.  |
|        --watch  |  [optional] Watch the filesystem for changes and report snapshots of artifacts running in specific filesystem paths to Kosli.  |


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

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### report one artifact running in a specific path in a filesystem

```shell
kosli snapshot path yourEnvironmentName 
	--path path/to/your/artifact/dir/or/file 
	--name yourArtifactDisplayName 

```

##### report one artifact running in a specific path in a filesystem AND exclude certain path patterns

```shell
kosli snapshot path yourEnvironmentName 
	--path path/to/your/artifact/dir 
	--name yourArtifactDisplayName 
	--exclude **/log,unwanted.txt,path/**/output.txt
```

