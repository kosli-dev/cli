---
title: "kosli environment allowedartifacts add"
---

## kosli environment allowedartifacts add

Add an artifact to an environment's allowlist. 

### Synopsis


   Add an artifact to an environment's allowlist. 
   The artifact SHA256 fingerprint is calculated and reported 
   or, alternatively, can be provided directly. 
   

```shell
kosli environment allowedartifacts add ARTIFACT-NAME-OR-PATH [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|    -e, --environment string  |  The environment name for which the artifact is allowlisted.  |
|    -h, --help  |  help for add  |
|        --reason string  |  The reason why this artifact is allowlisted.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'.  |


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


