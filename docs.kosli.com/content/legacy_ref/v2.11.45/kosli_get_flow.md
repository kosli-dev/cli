---
title: "kosli get flow"
beta: false
deprecated: false
summary: "Get the metadata of a specific flow."
---

# kosli get flow

## Synopsis

```shell
kosli get flow FLOW-NAME [flags]
```

Get the metadata of a specific flow.

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for flow  |
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

{{< raw-html >}}To view a live example of 'kosli get flow' you can run the commands below (for the <a href="https://app.kosli.com/cyber-dojo/environments/aws-prod/snapshots/">cyber-dojo</a> demo organization).<br/><a href="https://app.kosli.com/api/v2/livedocs/cyber-dojo/cli?command=kosli+get+flow+dashboard-ci+--output=json">Run the commands below and view the output.</a><pre>export KOSLI_ORG=cyber-dojo
export KOSLI_API_TOKEN=Pj_XT2deaVA6V1qrTlthuaWsmjVt4eaHQwqnwqjRO3A  # read-only
kosli get flow dashboard-ci --output=json</pre>{{< / raw-html >}}

