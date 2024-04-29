---
title: "kosli snapshot paths"
beta: false
deprecated: false
---

# kosli snapshot paths

## Synopsis

Report a snapshot of artifacts running from specific filesystem paths to Kosli.  
You can report directory or file artifacts in one or more filesystem paths. 
Artifacts names and the paths to include and exclude when fingerprinting them can either:
- be defined on command line using [`--name`, `--path`, `--exclude`] (suitable for reporting one artifact)
- OR, be defined in a paths file which can be provided using `--paths-file`.

Paths files can be in YAML, JSON or TOML formats.
They specify a list of artifacts to fingerprint. For each artifact, the file specifies a base path to look for the artifact in 
and (optionally) a list of paths to exclude. Excluded paths are relative to the artifact path(s) and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

This is an example YAML paths spec file:

version: 1
artifacts:
  artifact_name_a:
    path: dir1
    exclude: [subdir1, **/log]

```shell
kosli snapshot paths ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -x, --exclude strings  |  [optional] The comma-separated list of literal paths or glob patterns to exclude when fingerprinting the artifact. Can only be used together with --path .  |
|    -h, --help  |  help for paths  |
|        --name string  |  [conditional] The reported name of the artifact. Only required when --path is used.  |
|        --path string  |  [conditional] The base path for the artifact to snapshot. Cannot be used together with --paths-file .  |
|        --paths-file string  |  [conditional] The path to a paths file in YAML/JSON/TOML format. Cannot be used together with --path .  |


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

**report one or more artifacts running in a filesystem using a path spec file**

```shell
kosli snapshot paths yourEnvironmentName \
	--paths-file path/to/your/paths/file \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report one artifact running in a specific path in a filesystem**

```shell
kosli snapshot paths yourEnvironmentName \
	--path path/to/your/artifact/dir/or/file \
	--name yourArtifactDisplayName \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report one artifact running in a specific path in a filesystem AND exclude certain path patterns**

```shell
kosli snapshot paths yourEnvironmentName \
	--path path/to/your/artifact/dir \
	--name yourArtifactDisplayName \
	--exclude **/log,unwanted.txt,path/**/output.txt
	--api-token yourAPIToken \
	--org yourOrgName
```

