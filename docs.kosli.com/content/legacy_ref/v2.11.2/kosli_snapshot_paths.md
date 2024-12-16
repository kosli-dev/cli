---
title: "kosli snapshot paths"
beta: false
deprecated: false
---

# kosli snapshot paths

## Synopsis

Report a snapshot of artifacts running in specific filesystem paths to Kosli.  
You can report directory or file artifacts in one or more filesystem paths. 
Artifacts names and the paths to include and exclude when fingerprinting them can be 
defined in a paths file which can be provided using `--paths-file`.

Paths files can be in YAML, JSON or TOML formats.
They specify a list of artifacts to fingerprint. For each artifact, the file specifies a base path to look for the artifact in 
and (optionally) a list of paths to exclude. Excluded paths are relative to the artifact path(s) and can be literal paths or
glob patterns.  
The supported glob pattern syntax is what is documented here: https://pkg.go.dev/path/filepath#Match , 
plus the ability to use recursive globs "**"

To specify paths in a directory artifact that should always be excluded from the SHA256 calculation, you can add a `.kosli_ignore` file to the root of the artifact.
Each line should specify a relative path or path glob to be ignored. You can include comments in this file, using `#`.
The `.kosli_ignore` will be treated as part of the artifact like any other file,unless it is explicitly ignored itself.

This is an example YAML paths spec file:
```yaml
version: 1
artifacts:
  artifact_name_a:
    path: dir1
    exclude: [subdir1, **/log]
```

```shell
kosli snapshot paths ENVIRONMENT-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for paths  |
|        --paths-file string  |  The path to a paths file in YAML/JSON/TOML format. Cannot be used together with --path .  |


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

