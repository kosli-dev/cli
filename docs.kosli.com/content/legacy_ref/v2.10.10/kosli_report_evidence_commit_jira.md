---
title: "kosli report evidence commit jira"
beta: false
deprecated: true
---

# kosli report evidence commit jira

{{< hint danger >}}**kosli report evidence commit jira** is deprecated. See **kosli attest** commands.  Deprecated commands will be removed in a future release.{{< /hint >}}
## Synopsis

Report Jira evidence for a commit in Kosli flows.  
Parses the given commit's message or current branch name for Jira issue references of the 
form:  
'at least 2 characters long, starting with an uppercase letter project key followed by
dash and one or more digits'. 

The found issue references will be checked against Jira to confirm their existence.
The evidence is reported in all cases, and its compliance status depends on referencing
existing Jira issues.  
If you have wrong Jira credentials or wrong Jira-base-url it will be reported as non existing Jira issue.
This is because Jira returns same 404 error code in all cases.


```shell
kosli report evidence commit jira [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if no jira issue reference found, or jira issue does not exist, for the given commit or branch.  |
|    -b, --build-url string  |  The url of CI pipeline that generated the evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --commit string  |  Git commit for which to verify a given evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --evidence-fingerprint string  |  [optional] The SHA256 fingerprint of the evidence file or dir.  |
|    -e, --evidence-paths strings  |  [optional] The comma-separated list of paths containing supporting proof for the reported evidence. Paths can be for files or directories. All provided proofs will be uploaded to Kosli's evidence vault.  |
|        --evidence-url string  |  [optional] The external URL where the evidence file or dir is stored.  |
|    -f, --flows strings  |  [defaulted] The comma separated list of Kosli flows. Defaults to all flows of the org.  |
|    -h, --help  |  help for jira  |
|        --jira-api-token string  |  Jira API token (for Jira Cloud)  |
|        --jira-base-url string  |  The base url for the jira project, e.g. 'https://kosli.atlassian.net/browse/'  |
|        --jira-pat string  |  Jira personal access token (for self-hosted Jira)  |
|        --jira-username string  |  Jira username (for Jira Cloud)  |
|    -n, --name string  |  The name of the evidence.  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. (default ".")  |
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

**report Jira evidence for a commit related to one Kosli flow (with Jira Cloud)**

```shell
kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report Jira evidence for a commit related to one Kosli flow (with self-hosted Jira)**

```shell
kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://jira.example.com \
	--jira-pat yourJiraPATToken \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName

```

**report Jira  evidence for a commit related to multiple Kosli flows with user-data (with Jira Cloud)**

```shell
kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--flows yourFlowName1,yourFlowName2 \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName \
	--user-data /path/to/json/file.json


```

**fail if no issue reference is found, or the issue is not found in your jira instance**

```shell
kosli report evidence commit jira \
	--commit yourGitCommitSha1 \
	--name yourEvidenceName \
	--jira-base-url https://kosli.atlassian.net \
	--jira-username user@domain.com \
	--jira-api-token yourJiraAPIToken \
	--flows yourFlowName \
	--build-url https://exampleci.com \
	--api-token yourAPIToken \
	--org yourOrgName \
	--assert
```

