---
title: "kosli get snapshot"
beta: false
---

# kosli get snapshot

## Synopsis

Get a specific environment snapshot.
ENVIRONMENT-NAME-OR-EXPRESSION can specify the specific snapshot in five ways:
- environmentName
    - the latest snapshot, at the time of the request
    - e.g., **prod**
- environmentName#N
    - #N means snapshot number N, counting from 1
    - e.g., **prod#42** is prod's 42nd snapshot
- environmentName~N
    - ~N means N behind the latest snapshot, at the time of the request
    - e.g., **prod~5** is 5 snapshots behind the latest snapshot of prod
- environmentName@{YYYY-MM-DDTHH:MM:SS}
    - the snapshot at specific moment in time in UTC
    - e.g., **prod@{2023-10-02T12:00:00}** is prod's snapshot at midday, 2nd October 2023
- environmentName@{N.<hours|days|weeks|months>.ago}
    - the snapshot at a time relative to the time of the request
    - e.g., **prod@{2.hours.ago}** is prod's snapshot from 2 hours ago


```shell
kosli get snapshot ENVIRONMENT-NAME-OR-EXPRESSION [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshot  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


## Examples

```shell

# get the latest snapshot of an environment:
kosli get snapshot yourEnvironmentName
	--api-token yourAPIToken \
	--org yourOrgName 

# get the SECOND latest snapshot of an environment:
kosli get snapshot yourEnvironmentName~1
	--api-token yourAPIToken \
	--org yourOrgName 

# get the snapshot number 23 of an environment:
kosli get snapshot yourEnvironmentName#23
	--api-token yourAPIToken \
	--org yourOrgName 
	
# get the environment snapshot at midday (UTC), on valentine's day of 2023:
kosli get snapshot yourEnvironmentName@{2023-02-14T12:00:00}
	--api-token yourAPIToken \
	--org yourOrgName

# get the environment snapshot based on a relative time:
kosli get snapshot yourEnvironmentName@{3.weeks.ago}
--api-token yourAPIToken \
--org yourOrgName
```

