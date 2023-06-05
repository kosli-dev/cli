---
title: "kosli artifact get"
---

# kosli artifact get

## Synopsis

Get artifact from a specified pipeline
You can get an artifact by its fingerprint or by its git commit sha.
In case of using the git commit, it is possible to get multiple artifacts matching the git commit.

The expected argument is an expression to specify the artifact to get.
It has the format <PIPELINE_NAME><SEPARATOR><COMMIT_SHA1|ARTIFACT_SHA256> 

Specify SNAPPISH by:
	pipelineName@<fingerprint>  artifact with a given fingerprint. The fingerprint can be short or complete.
	pipelineName:<commit_sha>   artifact with a given commit SHA. The commit sha can be short or complete.

Examples of valid expressions are: pipe@184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0, pipe:110d048bf1fce72ba546cbafc4427fb21b958dee


```shell
kosli artifact get SNAPPISH [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for get  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |


## Examples

```shell
# get an artifact with a given fingerprint from a pipeline
kosli artifact get pipelineName@fingerprint \
	--api-token yourAPIToken \
	--owner orgName
# get an artifact with a given commit SHA from a pipeline
kosli artifact get pipelineName:commitSHA \
	--api-token yourAPIToken \
	--owner orgName
```

