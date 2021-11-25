---
title: "merkely fingerprint"
---

## merkely fingerprint

Print the SHA256 fingerprint of an artifact.

### Synopsis


Print the SHA256 fingerprint of an artifact. Requires artifact type flag to be set.
Artifact type can be one of: "file" for files, "dir" for directories, "docker" for docker images.


```
merkely fingerprint [flags]
```

### Options

```
  -h, --help          help for fingerprint
  -t, --type string   The type of the artifact to calculate its SHA256 fingerprint.
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

* [merkely](/client_reference/merkely/)	 - The Merkely evidence reporting CLI.

