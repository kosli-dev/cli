---
title: "kosli assert github-pullrequest"
---

## kosli assert github-pullrequest

Assert if a Github pull request for the commit which produces an artifact exists.

### Synopsis


   Check if a pull request exists in Github for an artifact (based on the git commit that produced it) and fail if it does not. 

```shell
kosli assert github-pullrequest [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --commit string  |  Git commit for which to find pull request evidence.  |
|        --github-org string  |  Github organization.  |
|        --github-token string  |  Github token.  |
|    -h, --help  |  help for github-pullrequest  |
|        --repository string  |  Git repository.  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "merkely")  |
|    -D, --dry-run  |  Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  The Kosli endpoint. (default "https://app.merkely.com")  |
|    -r, --max-api-retries int  |  How many times should API calls be retried when the API host is not reachable. (default 3)  |
|    -o, --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  Print verbose logs to stdout.  |


