---
title: "merkely pipeline artifact report creation"
---

## merkely pipeline artifact report creation

Report an artifact creation to a Merkely pipeline. 

### Synopsis


   Report an artifact creation to a pipeline in Merkely. 
   The artifact SHA256 fingerprint is calculated and reported 
   or, alternatively, can be provided directly. 
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

```shell
merkely pipeline artifact report creation ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string       The type of the artifact to calculate its SHA256 fingerprint.
  -b, --build-url string           The url of CI pipeline that built the artifact. (default "https://github.com/merkely-development/cli/actions/runs/1858697748")
  -u, --commit-url string          The url for the git commit that created the artifact. (default "https://github.com/merkely-development/cli/commit/81e68bf9da24970c6a6e8e51ed4dc1afc62d9c09")
  -C, --compliant                  Whether the artifact is compliant or not. (default true)
  -d, --description string         [optional] The artifact description.
  -g, --git-commit string          The git commit from which the artifact was created. (default "81e68bf9da24970c6a6e8e51ed4dc1afc62d9c09")
  -h, --help                       help for creation
  -p, --pipeline string            The Merkely pipeline name.
      --registry-password string   The docker registry password or access token.
      --registry-provider string   The docker registry provider or url.
      --registry-username string   The docker registry username.
  -s, --sha256 string              The SHA256 fingerprint for the artifact. Only required if you don't specify --artifact-type.
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

