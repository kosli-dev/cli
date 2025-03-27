---
title: "kosli report evidence commit generic"
beta: false
deprecated: true
summary: "Report Generic evidence for a commit in Kosli flows.  "
---

# kosli report evidence commit generic

{{% hint danger %}}
**kosli report evidence commit generic** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.
{{% /hint %}}
## Synopsis

Report Generic evidence for a commit in Kosli flows.  

```shell
kosli report evidence commit generic [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -C, --compliant  |  [defaulted] Whether the evidence is compliant or not. A boolean flag https://docs.kosli.com/faq/#boolean-flags  |
|    -d, --description string  |  [optional] The evidence description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for generic  |
|    -n, --name string  |  The name of the evidence.  |
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

**report Generic evidence for a commit related to one Kosli flow**

```shell
kosli report evidence commit generic \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--description "some description" \
	--compliant \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report Generic evidence for a commit related to multiple Kosli flows with user-data**

```shell
kosli report evidence commit generic \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--description "some description" \
	--compliant \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName \
	--user-data /path/to/json/file.json
```

