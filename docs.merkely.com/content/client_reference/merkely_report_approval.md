---
title: "merkely report approval"
---

## merkely report approval

Approve deploying an artifact in Merkely. 

### Synopsis


   Approve a deployment of an artifact in Merkely.
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly.
   

```
merkely report approval ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string   The type of the artifact to be approved. Options are [dir, file, docker]. Only required if you don't specify --sha256.
  -d, --description string     [optional] The approval description.
  -h, --help                   help for approval
      --newest-commit string   The source commit sha for the newest change in the deployment approval. (default "HEAD")
      --oldest-commit string   The source commit sha for the oldest change in the deployment approval.
  -p, --pipeline string        The Merkely pipeline name.
      --repo-root string       The directory where the source git repository is volume-mounted. (default ".")
  -s, --sha256 string          The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.
  -u, --user-data string       [optional] The path to a JSON file containing additional data you would like to attach to this approval.
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

* [merkely report](/client_reference/merkely_report/)	 - Report compliance events to Merkely.

