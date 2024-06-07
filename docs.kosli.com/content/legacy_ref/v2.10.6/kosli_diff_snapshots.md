---
title: "kosli diff snapshots"
beta: false
deprecated: false
---

# kosli diff snapshots

## Synopsis

Diff environment snapshots.  
Specify SNAPPISH_1 and SNAPPISH_2 by:
- environmentName
    - the latest snapshot for environmentName, at the time of the request
    - e.g., **prod**
- environmentName#N
    - the Nth snapshot, counting from 1
    - e.g., **prod#42**
- environmentName~N
    - the Nth snapshot behind the latest, at the time of the request
    - e.g., **prod~5**
- environmentName@{YYYY-MM-DDTHH:MM:SS}
    - the snapshot at specific moment in time in UTC
    - e.g., **prod@{2023-10-02T12:00:00}**
- environmentName@{N.<hours|days|weeks|months>.ago}
    - the snapshot at a time relative to the time of the request
    - e.g., **prod@{2.hours.ago}**


```shell
kosli diff snapshots SNAPPISH_1 SNAPPISH_2 [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshots  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -u, --show-unchanged  |  [defaulted] Show the unchanged artifacts present in both snapshots within the diff output.  |


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


## Live Example

{{< raw-html >}}<pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
<a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+diff+snapshots+aws-beta+aws-prod+--output=json">kosli diff snapshots aws-beta aws-prod --output=json</a></pre>{{< / raw-html >}}

## Examples Use Cases

**compare the third latest snapshot in an environment to the latest**

```shell
kosli diff snapshots envName~3 envName \
	--api-token yourAPIToken \
	--org orgName
	
```

**compare snapshots of two different environments of the same type**

```shell
kosli diff snapshots envName1 envName2 \
	--api-token yourAPIToken \
	--org orgName

```

**show the not-changed artifacts in both snapshots**

```shell
kosli diff snapshots envName1 envName2 \
	--show-unchanged \
	--api-token yourAPIToken \
	--org orgName

```

**compare the snapshot from 2 weeks ago in an environment to the latest**

```shell
kosli diff snapshots envName@{2.weeks.ago} envName \
--api-token yourAPIToken \
--org orgName
```

