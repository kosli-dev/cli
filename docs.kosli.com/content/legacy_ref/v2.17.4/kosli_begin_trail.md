---
title: "kosli begin trail"
beta: false
deprecated: false
summary: "Begin or update a Kosli flow trail."
---

# kosli begin trail

## Synopsis

```shell
kosli begin trail TRAIL-NAME [flags]
```

Begin or update a Kosli flow trail.

You can optionally associate the trail to a git commit using `--commit` (requires access to a git repo). And you
can optionally redact some of the git commit data sent to Kosli using `--redact-commit-info`.
To record repository information, all three of `--repo-id`, `--repo-url`, and `--repository` must be set together.
These are automatically set in GitHub Actions, GitLab CI, Bitbucket Pipelines, and Azure DevOps.
In other CI systems, set them explicitly to capture repository metadata.

`TRAIL-NAME`s must start with a letter or number, and only contain letters, numbers, `.`, `-`, `_`, and `~`.


## Flags
| Flag | Description |
| :--- | :--- |
|    -g, --commit string  |  [defaulted] The git commit from which the trail is begun. (defaulted in some CIs: https://docs.kosli.com/ci-defaults, otherwise defaults to HEAD ).  |
|        --description string  |  [optional] The Kosli trail description.  |
|    -D, --dry-run  |  [optional] Run in dry-run mode. When enabled, no data is sent to Kosli and the CLI exits with 0 exit code regardless of any errors.  |
|        --external-fingerprint stringToString  |  [optional] A SHA256 fingerprint of an external attachment represented by --external-url. The format is label=fingerprint (labels cannot contain '.' or '='). This flag can be set multiple times. There must be an external url with a matching label for each external fingerprint.  |
|        --external-url stringToString  |  [optional] Add labeled reference URL for an external resource. The format is label=url (labels cannot contain '.' or '='). This flag can be set multiple times. If the resource is a file or dir, you can optionally add its fingerprint via --external-fingerprint  |
|        --flow string  |  The Kosli flow name.  |
|    -h, --help  |  help for trail  |
|    -o, --origin-url string  |  [optional] The url pointing to where the attestation came from or is related. (defaulted to the CI url in some CIs: https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables ).  |
|        --redact-commit-info strings  |  [optional] The list of commit info to be redacted before sending to Kosli. Allowed values are one or more of [author, message, branch].  |
|        --repo-id string  |  [conditional] The stable, unique identifier for the repository in your VCS provider (e.g. a numeric ID). Do not use the repository name as it can change if the repo is renamed. All three of --repo-id, --repo-url and --repository must be set to record repository information (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --repo-provider string  |  [optional] The source code hosting provider. One of: github, gitlab, bitbucket, azure-devops (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --repo-root string  |  [defaulted] The directory where the source git repository is available. Only used if --commit is used or defaulted in CI, see https://docs.kosli.com/integrations/ci_cd/#defaulted-kosli-command-flags-from-ci-variables . (default ".")  |
|        --repo-url string  |  [conditional] The URL of the repository. Must be a valid URL. All three of --repo-id, --repo-url and --repository must be set to record repository information (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|        --repository string  |  [conditional] The name of the repository (e.g. owner/repo-name). All three of --repo-id, --repo-url and --repository must be set to record repository information (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -f, --template-file string  |  [optional] The path to a yaml template file.  |
|    -u, --user-data string  |  [optional] The path to a JSON file containing additional data you would like to attach to the flow trail.  |


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

{{< tabs "live-examples" "col-no-wrap" >}}{{< tab "GitHub" >}}View an example of the `kosli begin trail` command in GitHub.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=kosli%2Bbegin%2Btrail), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=kosli%2Bbegin%2Btrail).{{< /tab >}}{{< tab "GitLab" >}}View an example of the `kosli begin trail` command in GitLab.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=gitlab&command=kosli%2Bbegin%2Btrail), which created [this Kosli Event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=gitlab&command=kosli%2Bbegin%2Btrail).{{< /tab >}}{{< /tabs >}}

## Examples Use Cases

These examples all assume that the flags  `--api-token`, `--org`, `--host`, (and `--flow`, `--trail` when required), are [set/provided](https://docs.kosli.com/getting_started/install/#assigning-flags-via-environment-variables). 

##### begin/update a Kosli flow trail

```shell
kosli begin trail yourTrailName 
	--description yourTrailDescription 
	--template-file /path/to/your/template/file.yml 
	--user-data /path/to/your/user-data/file.json 
```

