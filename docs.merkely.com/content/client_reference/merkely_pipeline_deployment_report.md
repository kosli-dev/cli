---
title: "merkely pipeline deployment report"
---

## merkely pipeline deployment report

Report a deployment to Merkely. 

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

```shell
merkely pipeline deployment report [ARTIFACT-NAME-OR-PATH] [flags]
```

### Options

```
  -t, --artifact-type string       The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]
  -b, --build-url string           The url of CI pipeline that built the artifact. (default "https://github.com/merkely-development/cli/actions/runs/1915357107")
  -d, --description string         [optional] The artifact description.
  -e, --environment string         The environment name.
  -h, --help                       help for report
  -p, --pipeline string            The Merkely pipeline name.
      --registry-password string   The docker registry password or access token.
      --registry-provider string   The docker registry provider or url.
      --registry-username string   The docker registry username.
  -s, --sha256 string              The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.
  -u, --user-data string           [optional] The path to a JSON file containing additional data you would like to attach to this deployment.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely user or organization.
  -v, --verbose               Print verbose logs to stdout.
```

