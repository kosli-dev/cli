---
title: "kosli commit report evidence snyk"
---

## kosli commit report evidence snyk

Report Snyk evidence for a commit in a Kosli pipeline.

### Synopsis

Report Snyk evidence for a commit in a Kosli pipeline.

```shell
kosli commit report evidence snyk [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  The git commit SHA1 for which the evidence belongs. (defaulted in some CIs: https://docs.kosli.com/ci-defaults).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the evidence.  |
|    -p, --pipelines strings  |  The comma separated list of pipelines for which a commit evidence belongs.  |
|    -R, --scan-results string  |  The path to Snyk scan results Json file.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this evidence.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


### Examples

```shell

# report Snyk evidence for a commit related to one Kosli pipeline:
kosli commit report evidence snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--pipelines yourPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk evidence for a commit related to multiple Kosli pipelines:
kosli commit report evidence snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--pipelines yourFirstPipelineName,yourSecondPipelineName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--owner yourOrgName	\
	--scan-results yourSnykJSONScanResults

```

