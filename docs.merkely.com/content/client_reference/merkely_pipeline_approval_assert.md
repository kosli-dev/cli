---
title: "merkely pipeline approval assert"
---

## merkely pipeline approval assert

Assert if an artifact in Merkely has been approved for deployment.

### Synopsis


Assert if an artifact in Merkely has been approved for deployment. Exits with non-zero code if artifact has not been approved.
The artifact SHA256 fingerprint is calculated (based on --artifact-type flag) or alternatively it can be provided directly (with --sha256 flag).

```shell
merkely pipeline approval assert [ARTIFACT-NAME-OR-PATH] [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -t, --artifact-type string  |  The type of the artifact to calculate its SHA256 fingerprint. One of: [docker, file, dir]. Only required if you don't specify 'sha256'  |
|    -h, --help  |  help for assert  |
|    -p, --pipeline string  |  The Merkely pipeline name.  |
|        --registry-password string  |  The docker registry password or access token.  |
|        --registry-provider string  |  The docker registry provider or url.  |
|        --registry-username string  |  The docker registry username.  |
|    -s, --sha256 string  |  The SHA256 fingerprint for the artifact. Only required if you don't specify 'artifact-type'.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The merkely API token.  |
|    -c, --config-file string  |  [optional] The merkely config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The merkely endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The merkely user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


### Examples

```shell

# Assert that a file tyoe artifact has been approved
merkely pipeline approval assert FILE.tgz \
	--api-token yourAPIToken \
	--artifact-type file \
	--owner yourOrgName \
	--pipeline yourPipelineName 


# Assert that an artifact with a provided fingerprint (sha256) has been approved
	merkely pipeline approval assert \
		--api-token yourAPIToken \
		--owner yourOrgName \
		--pipeline yourPipelineName \
		--sha256 yourSha256

```

