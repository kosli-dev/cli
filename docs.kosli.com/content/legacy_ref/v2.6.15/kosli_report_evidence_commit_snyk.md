---
title: "kosli report evidence commit snyk"
beta: false
---

# kosli report evidence commit snyk

## Synopsis

Report Snyk vulnerability scan evidence for a commit in Kosli flows.    
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.  
If no vulnerabilities are detected the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.


```shell
kosli report evidence commit snyk [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify and given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the evidence.  |
|    -R, --scan-results string  |  The path to Snyk scan results JSON file from 'snyk test' and 'snyk container test'. The Snyk results will be uploaded to Kosli's evidence vault.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to this evidence.  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# report Snyk evidence for a commit related to one Kosli flow:
kosli report evidence commit snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults

# report Snyk evidence for a commit related to multiple Kosli flows:
kosli report evidence commit snyk \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName	\
	--scan-results yourSnykJSONScanResults

```

