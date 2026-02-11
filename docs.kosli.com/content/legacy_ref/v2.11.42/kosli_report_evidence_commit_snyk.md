---
title: "kosli report evidence commit snyk"
beta: false
deprecated: true
summary: "Report Snyk vulnerability scan evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit snyk

{{% hint danger %}}
**kosli report evidence commit snyk** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

```shell
kosli report evidence commit snyk [flags]
```

Report Snyk vulnerability scan evidence for a commit in Kosli flows.    
The --scan-results .json file is parsed and uploaded to Kosli's evidence vault.

In CLI <v2.8.2, Snyk results could only be in the Snyk JSON output format. "snyk code test" results were not supported by 
this command and could be reported as generic evidence.

Starting from v2.8.2, the Snyk results can be in Snyk JSON or SARIF output format for "snyk container test". 
"snyk code test" is now supported but only in the SARIF format.

If no vulnerabilities are detected the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.


## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for snyk  |
|    -n, --name string  |  The name of the evidence.  |
|    -R, --scan-results string  |  The path to Snyk SARIF or JSON scan results file from 'snyk test' and 'snyk container test'. By default, the Snyk results will be uploaded to Kosli's evidence vault.  |
|        --upload-results  |  [defaulted] Whether to upload the provided Snyk results file as an attachment to Kosli or not. (default true)  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the evidence.  |


## Flags inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|        --http-proxy string  |  [optional] The HTTP proxy URL including protocol and port number. e.g. 'http://proxy-server-ip:proxy-port'  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### report Snyk evidence for a commit related to one Kosli flow

```shell
kosli report evidence commit snyk 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName1 
	--build-url https://exampleci.com 
	--scan-results yourSnykJSONScanResults

```

##### report Snyk evidence for a commit related to multiple Kosli flows

```shell
kosli report evidence commit snyk 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--scan-results yourSnykJSONScanResults
```

