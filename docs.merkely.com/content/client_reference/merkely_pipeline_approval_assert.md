---
title: "merkely pipeline approval assert"
---

## merkely pipeline approval assert

Assert if an artifact in Merkely has been approved for deployment.

### Synopsis

Assert if an artifact in Merkely has been approved for deployment.
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   

```shell
merkely pipeline approval assert ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string       The type of the artifact to calculate its SHA256 fingerprint.
  -h, --help                       help for assert
  -p, --pipeline string            The Merkely pipeline name.
      --registry-password string   The docker registry password or access token.
      --registry-provider string   The docker registry provider or url.
      --registry-username string   The docker registry username.
  -s, --sha256 string              The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.
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

