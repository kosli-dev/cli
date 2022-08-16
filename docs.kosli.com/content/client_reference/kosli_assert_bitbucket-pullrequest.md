---
title: "kosli assert bitbucket-pullrequest"
---

## kosli assert bitbucket-pullrequest

Assert if a Bitbucket pull request for the commit which produces an artifact exists.

### Synopsis


   Check if a pull request exists in Bitbucket for an artifact (based on the git commit that produced it) and fail if it does not. 

```shell
kosli assert bitbucket-pullrequest [flags]
```

### Flags
| Flag | Description |
| :--- | :--- |
|        --bitbucket-password string  |  Bitbucket password.  |
|        --bitbucket-username string  |  Bitbucket user name.  |
|        --bitbucket-workspace string  |  Bitbucket workspace.  |
|        --commit string  |  Git commit for which to find pull request evidence. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |
|    -h, --help  |  help for bitbucket-pullrequest  |
|        --repository string  |  Git repository. (defaulted in some CIs: https://docs.kosli.com/ci-defaults ).  |


### Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|    -D, --dry-run  |  [optional] Whether to run in dry-run mode. When enabled, data is not sent to Kosli and the CLI exits with 0 exit code regardless of errors.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --owner string  |  The Kosli user or organization.  |
|    -v, --verbose  |  [optional] Print verbose logs to stdout.  |


