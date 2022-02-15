---
title: "merkely environment allowedartifacts add"
---

## merkely environment allowedartifacts add

Add an artifact to an environment's allowlist. 

### Synopsis


   Add an artifact to an environment's allowlist. 
   The artifact SHA256 fingerprint is calculated and reported 
   or, alternatively, can be provided directly. 
   

```shell
merkely environment allowedartifacts add ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string       The type of the artifact to calculate its SHA256 fingerprint.
  -e, --environment string         The environment name for which the artifact is allowlisted.
  -h, --help                       help for add
      --reason string              The reason why this artifact is allowlisted.
      --registry-password string   The docker registry password or access token.
      --registry-provider string   The docker registry provider or url.
      --registry-username string   The docker registry username.
  -s, --sha256 string              The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely user or organization.
  -v, --verbose               Print verbose logs to stdout.
```

