---
title: "kosli create flow"
beta: false
deprecated: false
summary: "Create or update a Kosli flow."
---

# kosli create flow

## Synopsis

Create or update a Kosli flow.
You can specify flow parameters in flags.

`FLOW-NAME`s must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.


```shell
kosli create flow FLOW-NAME [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --description string  |  [optional] The Kosli flow description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for flow  |
|    -t, --template strings  |  [defaulted] The comma-separated list of required compliance controls names.  |
|    -f, --template-file string  |  [optional] The path to a yaml template file. Cannot be used together with --use-empty-template  |
|        --use-empty-template  |  Use an empty template for the flow creation without specifying a file. Cannot be used together with --template or --template-file  |
|        --visibility string  |  [defaulted] The visibility of the Kosli flow. Valid visibilities are [public, private]. (default "private")  |


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


## Live Examples in different CI systems

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli create flow` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+create+flow), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli+create+flow).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli create flow` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+create+flow), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli+create+flow).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are set/provided. 

**create/update a Kosli flow (with empty template)**

```shell
kosli create flow yourFlowName 
	--description yourFlowDescription 
	--visibility private OR public 
	--use-empty-template 

```

**create/update a Kosli flow (with template file)**

```shell
kosli create flow yourFlowName 
	--description yourFlowDescription 
	--visibility private OR public 
	--template-file /path/to/your/template/file.yml 
```

