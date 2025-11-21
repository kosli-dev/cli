---
title: "kosli tag"
beta: false
deprecated: false
summary: "Tag a resource in Kosli with key-value pairs.  "
---

# kosli tag

## Synopsis

```shell
kosli tag RESOURCE-TYPE RESOURCE-ID [flags]
```

Tag a resource in Kosli with key-value pairs.  
use --set to add or update tags, and --unset to remove tags.


## Flags
| Flag | Description |
| :--- | :--- |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|    -h, --help  |  help for tag  |
|        --set stringToString  |  [optional] The key-value pairs to tag the resource with. The format is: key=value  |
|        --unset strings  |  [optional] The list of tag keys to remove from the resource.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli tag` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli+tag){{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli tag` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli+tag){{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

**add/update tags to a flow**

```shell
kosli tag flow yourFlowName 
	--set key1=value1 
	--set key2=value2 

```

**tag an environment**

```shell
kosli tag env yourEnvironmentName 
	--set key1=value1 
	--set key2=value2 

```

**add/update tags to an environment**

```shell
kosli tag env yourEnvironmentName 
	--set key1=value1 
	--set key2=value2 

```

**remove tags from an environment**

```shell
kosli tag env yourEnvironmentName 
	--unset key1=value1 
```

