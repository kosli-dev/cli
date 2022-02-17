---
title: "merkely environment report s3"
---

## merkely environment report s3

Report artifact from AWS S3 bucket to Merkely.

### Synopsis


Report the artifact deployed in an AWS S3 bucket and their digests 
and reports it to Merkely. 


```shell
merkely environment report s3 env-name [flags]
```

### Examples

```shell

* report what's running in an AWS S3 bucket:
merkely environment report s3 prod --api-token 1234 --owner exampleOrg

```

### Options

```
      --access-key string   The AWS access key
  -C, --bucket string       The name of the S3 bucket.
  -h, --help                help for s3
      --region string       The AWS region
      --secret-key string   The AWS secret key
```

### Options inherited from parent commands

```
  -a, --api-token string      The merkely API token.
  -c, --config-file string    [optional] The merkely config file path. (default "merkely")
  -D, --dry-run               Whether to send the request to the endpoint or just log it in stdout.
  -H, --host string           The merkely endpoint. (default "https://app.merkely.com")
  -r, --max-api-retries int   How many times should API calls be retried when the API host is not reachable. (default 3)
  -o, --owner string          The merkely user or organization.
  -v, --verbose               Print verbose logs to stdout.
```

