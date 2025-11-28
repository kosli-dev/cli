---
title: "kosli get snapshot"
beta: false
deprecated: false
summary: "Get a specified environment snapshot.  "
---

# kosli get snapshot

## Synopsis

```shell
kosli get snapshot ENVIRONMENT-NAME-OR-EXPRESSION [flags]
```

Get a specified environment snapshot.  
ENVIRONMENT-NAME-OR-EXPRESSION can be specified as follows:
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


## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for snapshot  |
|    -o, --output string  |  [defaulted] The format of the output. Valid formats are: [table, json]. (default "table")  |


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

{{< raw-html >}}To view a live example of 'kosli get snapshot' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+snapshot+aws-prod+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get snapshot aws-prod --output=json</pre>{{< / raw-html >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### get the latest snapshot of an environment

```shell
kosli get snapshot yourEnvironmentName

```

##### get the SECOND latest snapshot of an environment

```shell
kosli get snapshot yourEnvironmentName~1

```

##### get the snapshot number 23 of an environment

```shell
kosli get snapshot yourEnvironmentName#23

```

##### get the environment snapshot at midday (UTC), on valentine's day of 2023

```shell
kosli get snapshot yourEnvironmentName@{2023-02-14T12:00:00}

```

##### get the environment snapshot based on a relative time

```shell
kosli get snapshot yourEnvironmentName@{3.weeks.ago}
```

