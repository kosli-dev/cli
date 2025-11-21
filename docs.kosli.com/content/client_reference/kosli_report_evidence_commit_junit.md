---
title: "kosli report evidence commit junit"
beta: false
deprecated: true
summary: "Report JUnit test evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit junit

{{% hint danger %}}
**kosli report evidence commit junit** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

```shell
kosli report evidence commit junit [flags]
```

Report JUnit test evidence for a commit in Kosli flows.    
All .xml files from --results-dir are parsed and uploaded to Kosli's evidence vault.  
If there are no failing tests and no errors the evidence is reported as compliant. Otherwise the evidence is reported as non-compliant.


## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for junit  |
|    -n, --name string  |  The name of the evidence.  |
|    -R, --results-dir string  |  [defaulted] The path to a directory with JUnit test results. By default, the directory will be uploaded to Kosli's evidence vault. (default ".")  |
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

##### report JUnit test evidence for a commit related to one Kosli flow

```shell
kosli report evidence commit junit 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults

```

##### report JUnit test evidence for a commit related to multiple Kosli flows

```shell
kosli report evidence commit junit 
	--commit yourGitCommitSha1 
	--name yourEvidenceName 
	--flows yourFlowName1,yourFlowName2 
	--build-url https://exampleci.com 
	--results-dir yourFolderWithJUnitResults
```

