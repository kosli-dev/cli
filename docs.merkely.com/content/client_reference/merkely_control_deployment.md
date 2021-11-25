---
title: "merkely control deployment"
---

## merkely control deployment

Check if an artifact in Merkely has been approved for deployment.

### Synopsis

Check if an artifact in Merkely has been approved for deployment.
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly.
   

```
merkely control deployment ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string   The type of the artifact to be approved. Options are [dir, file, docker]. Only required if you don't specify --sha256.
  -h, --help                   help for deployment
  -p, --pipeline string        The Merkely pipeline name.
  -s, --sha256 string          The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely organization.
  -v, --verbose               Print verbose logs to stdout.
```

### SEE ALSO

* [merkely control](/client_reference/merkely_control/)	 - Check if artifact is allowed to be deployed.

