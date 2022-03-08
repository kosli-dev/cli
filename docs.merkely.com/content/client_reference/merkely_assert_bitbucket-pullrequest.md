---
title: "merkely assert bitbucket-pullrequest"
---

## merkely assert bitbucket-pullrequest

Assert if a Bitbucket pull request for the commit which produces an artifact exists.

### Synopsis


   Check if a pull request exists in Bitbucket for an artifact (based on the git commit that produced it) and fail if it does not. 

```shell
merkely assert bitbucket-pullrequest [flags]
```

### Options

```
      --bitbucket-password string    Bitbucket password.
      --bitbucket-username string    Bitbucket user name.
      --bitbucket-workspace string   Bitbucket workspace.
      --commit string                Git commit for which to find pull request evidence.
  -h, --help                         help for bitbucket-pullrequest
      --repository string            Git repository.
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to run in dry-run mode. When enabled, data is not sent to Merkely and the CLI exits with 0 exit code regardless of errors.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely user or organization.
  -v, --verbose               Print verbose logs to stdout.
```

