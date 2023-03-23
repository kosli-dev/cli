---
title: "kosli status"
---

# kosli status

## Synopsis

Check the status of Kosli server. 
The status is logged and the command always exits with 0 exit code.
If you like to assert the Kosli server status, you can use the --assert flag or the "kosli assert status" command.

```shell
kosli status [flags]
```

## Flags
| Flag | Description |
| :--- | :--- |
|        --assert  |  [optional] Exit with non-zero code if Kosli server is not responding.  |
|    -h, --help  |  help for status  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout.  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


