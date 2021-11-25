---
title: "merkely report evidence"
---

## merkely report evidence

Report/Log an evidence to an artifact in Merkely. 

### Synopsis


   Report an evidence to an artifact in Merkely.
   The artifact SHA256 fingerprint is calculated or alternatively it can be provided directly.
   The following flags are defaulted as follows in the CI list below:

   
	| Bitbucket 
	|---------------------------------------------------------------------------
	| build-url : https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER}
	|---------------------------------------------------------------------------
	| Github 
	|---------------------------------------------------------------------------
	| build-url : ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}
	|---------------------------------------------------------------------------
	| Teamcity 
	|---------------------------------------------------------------------------
	|---------------------------------------------------------------------------

```
merkely report evidence ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string   The type of the artifact related to the evidence. Options are [dir, file, docker].
  -b, --build-url string       The url of CI pipeline that generated the evidence.
  -C, --compliant              Whether the evidence is compliant or not. (default true)
  -d, --description string     [optional] The evidence description.
  -e, --evidence-type string   The type of evidence being reported.
  -h, --help                   help for evidence
  -p, --pipeline string        The Merkely pipeline name.
  -s, --sha256 string          The SHA256 fingerprint for the artifact. Only required if you don't specify --type.
  -u, --user-data string       [optional] The path to a JSON file containing additional data you would like to attach to this evidence.
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

