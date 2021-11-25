---
title: "merkely report artifact"
---

## merkely report artifact

Report/Log an artifact to Merkely. 

### Synopsis


   Report an artifact to a pipeline in Merkely.
   The artifact SHA256 fingerprint is calculated and reported
   or,alternatively, can be provided directly.
   The following flags are defaulted as follows in the CI list below:

   
	| Bitbucket 
	|---------------------------------------------------------------------------
	| git-commit : ${BITBUCKET_COMMIT}
	| build-url : https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER}
	| commit-url : https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT}
	|---------------------------------------------------------------------------
	| Github 
	|---------------------------------------------------------------------------
	| git-commit : ${GITHUB_SHA}
	| build-url : ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}
	| commit-url : ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA}
	|---------------------------------------------------------------------------
	| Teamcity 
	|---------------------------------------------------------------------------
	| git-commit : ${BUILD_VCS_NUMBER}
	|---------------------------------------------------------------------------

```
merkely report artifact ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -b, --build-url string     The url of CI pipeline that built the artifact.
  -u, --commit-url string    The url for the git commit that created the artifact.
  -C, --compliant            Whether the artifact is compliant or not. (default true)
  -d, --description string   [optional] The artifact description.
  -g, --git-commit string    The git commit from which the artifact was created.
  -h, --help                 help for artifact
  -p, --pipeline string      The Merkely pipeline name.
  -s, --sha256 string        The SHA256 fingerprint for the artifact. Only required if you don't specify --type.
  -t, --type string          The type of the artifact. Options are [dir, file, docker].
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

