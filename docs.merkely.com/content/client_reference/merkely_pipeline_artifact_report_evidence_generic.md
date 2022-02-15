---
title: "merkely pipeline artifact report evidence generic"
---

## merkely pipeline artifact report evidence generic

Report a generic evidence to an artifact in a Merkely pipeline. 

### Synopsis


   Report a generic evidence to an artifact in a Merkely pipeline. 
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

```shell
merkely pipeline artifact report evidence generic ARTIFACT-NAME-OR-PATH [flags]
```

### Options

```
  -t, --artifact-type string       The type of the artifact to calculate its SHA256 fingerprint.
  -b, --build-url string           The url of CI pipeline that generated the evidence. (default "https://github.com/merkely-development/cli/actions/runs/1847906195")
  -C, --compliant                  Whether the evidence is compliant or not. (default true)
  -d, --description string         [optional] The evidence description.
  -e, --evidence-type string       The type of evidence being reported.
  -h, --help                       help for generic
  -p, --pipeline string            The Merkely pipeline name.
      --registry-password string   The docker registry password or access token.
      --registry-provider string   The docker registry provider or url.
      --registry-username string   The docker registry username.
  -s, --sha256 string              The SHA256 fingerprint for the artifact. Only required if you don't specify --type.
  -u, --user-data string           [optional] The path to a JSON file containing additional data you would like to attach to this evidence.
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

