---
title: "kosli completion"
beta: false
deprecated: false
---

# kosli completion

## Synopsis

To load completions:

  ### Bash

```
  $ source <(kosli completion bash)
```
  To load completions for each session, execute once:  

  On Linux:
  ```
  $ kosli completion bash > /etc/bash_completion.d/kosli
  ``` 
  On macOS:
  ```
  $ kosli completion bash > $(brew --prefix)/etc/bash_completion.d/kosli
  ```
  ### Zsh

  If shell completion is not already enabled in your environment,  
you will need to enable it.  You can execute the following once:
  ```
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  ```
  To load completions for each session, execute once:
  ```
  $ kosli completion zsh > "${fpath[1]}/_kosli"
  ```
  You will need to start a new shell for this setup to take effect.

  ### fish
  ```
  $ kosli completion fish | source
  ```
  To load completions for each session, execute once:
  ``` 
  $ kosli completion fish > ~/.config/fish/completions/kosli.fish
  ```
  ### PowerShell
  ```
  PS> kosli completion powershell | Out-String | Invoke-Expression
  ```
 To load completions for every new session, run:
 ```
 PS> kosli completion powershell > kosli.ps1
 ``` 
 and source this file from your PowerShell profile.


```shell
kosli completion [bash|zsh|fish|powershell]
```

## Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for completion  |


## Options inherited from parent commands
| Flag | Description |
| :--- | :--- |
|    -a, --api-token string  |  The Kosli API token.  |
|    -c, --config-file string  |  [optional] The Kosli config file path. (default "kosli")  |
|        --debug  |  [optional] Print debug logs to stdout. A boolean flag https://docs.kosli.com/faq/#boolean-flags (default false)  |
|    -H, --host string  |  [defaulted] The Kosli endpoint. (default "https://app.kosli.com")  |
|    -r, --max-api-retries int  |  [defaulted] How many times should API calls be retried when the API host is not reachable. (default 3)  |
|        --org string  |  The Kosli organization.  |


