---
title: "merkely report deployment"
---

## merkely report deployment

Report/Log a deployment to Merkely. 

### Synopsis


   Report a deployment of an artifact to an environment in Merkely.
   The artifact SHA256 fingerprint is calculated and reported
   or,alternatively, can be provided directly.
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
merkely report deployment ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -b, --build-url string     The url of CI pipeline that built the artifact.
  -d, --description string   [optional] The artifact description.
  -e, --environment string   The environment name.
  -h, --help                 help for deployment
  -p, --pipeline string      The Merkely pipeline name.
  -s, --sha256 string        The SHA256 fingerprint for the artifact. Only required if you don't specify --type.
  -t, --type string          The type of the artifact. Options are [dir, file, docker].
  -u, --user-data string     [optional] The path to a JSON file containing additional data you would like to attach to this deployment.
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

