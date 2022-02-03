---
title: "merkely pipeline approval report"
---

## merkely pipeline approval report

Report approval of deploying an artifact in Merkely. 

### Synopsis


   Approve a deployment of an artifact in Merkely. 
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly. 
   

```shell
merkely pipeline approval report ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string       The type of the artifact to calculate its SHA256 fingerprint.
  -d, --description string         [optional] The approval description.
  -h, --help                       help for report
      --newest-commit string       The source commit sha for the newest change in the deployment approval. (default "HEAD")
      --oldest-commit string       The source commit sha for the oldest change in the deployment approval.
  -p, --pipeline string            The Merkely pipeline name.
      --registry-password string   The docker registry password or access token.
      --registry-provider string   The docker registry provider or url.
      --registry-username string   The docker registry username.
      --repo-root string           The directory where the source git repository is volume-mounted. (default ".")
  -s, --sha256 string              The SHA256 fingerprint for the artifact to be approved. Only required if you don't specify --type.
  -u, --user-data string           [optional] The path to a JSON file containing additional data you would like to attach to this approval.
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

