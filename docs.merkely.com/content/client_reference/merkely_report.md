---
title: "merkely report"
---

## merkely report

Report compliance events to Merkely.

### Synopsis


Report compliance events back to Merkely.


### Options

```
  -h, --help   help for report
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
* [merkely report approval](/client_reference/merkely_report_approval/)	 - Approve deploying an artifact in Merkely. 
* [merkely report artifact](/client_reference/merkely_report_artifact/)	 - Report/Log an artifact to Merkely. 
* [merkely report deployment](/client_reference/merkely_report_deployment/)	 - Report/Log a deployment to Merkely. 
* [merkely report env](/client_reference/merkely_report_env/)	 - Report running artifacts in an environment to Merkely.
* [merkely report evidence](/client_reference/merkely_report_evidence/)	 - Report/Log an evidence to an artifact in Merkely. 
* [merkely report request](/client_reference/merkely_report_request/)	 - Request an approval for deploying an artifact in Merkely. 
* [merkely report test](/client_reference/merkely_report_test/)	 - Report/Log a JUnit test evidence to an artifact in Merkely. 

