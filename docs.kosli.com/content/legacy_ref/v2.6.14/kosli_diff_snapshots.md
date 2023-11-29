---
title: "kosli diff snapshots"
beta: false
---

# kosli diff snapshots

## Synopsis

Diff environment snapshots.  
Specify SNAPPISH_1 and SNAPPISH_2 by:  
- environmentName~<N>  N'th behind the latest snapshot  
- environmentName#<N>  snapshot number N  
- environmentName@{YYYY-MM-DDTHH:MM:SS} snapshot at specific moment in time in UTC
- environmentName@{<N>.<hours|days|weeks|months>.ago} snapshot at a relative time
- environmentName      the latest snapshot

```shell
kosli diff snapshots SNAPPISH_1 SNAPPISH_2 [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshots  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |
|    -u, --show-unchanged  |  [defaulted] Show the unchanged artifacts present in both snapshots within the diff output.  |


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

# compare the third latest snapshot in an environment to the latest
kosli diff snapshots envName~3 envName \
	--api-token yourAPIToken \
	--org orgName
	
# compare snapshots of two different environments of the same type
kosli diff snapshots envName1 envName2 \
	--api-token yourAPIToken \
	--org orgName

# show the not-changed artifacts in both snapshots
kosli diff snapshots envName1 envName2 \
	--show-unchanged \
	--api-token yourAPIToken \
	--org orgName

# compare the snapshot from 2 weeks ago in an environment to the latest
kosli diff snapshots envName@{2.weeks.ago} envName \
--api-token yourAPIToken \
--org orgName
```

