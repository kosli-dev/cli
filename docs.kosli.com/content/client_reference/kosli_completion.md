---
title: "kosli completion"
---

## kosli completion

Generate completion script

### Synopsis

To load completions:

Bash:

  $ source <(kosli completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ kosli completion bash > /etc/bash_completion.d/kosli
  # macOS:
  $ kosli completion bash > $(brew --prefix)/etc/bash_completion.d/kosli

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ kosli completion zsh > "${fpath[1]}/_kosli"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ kosli completion fish | source

  # To load completions for each session, execute once:
  $ kosli completion fish > ~/.config/fish/completions/kosli.fish

PowerShell:

  PS> kosli completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> kosli completion powershell > kosli.ps1
  # and source this file from your PowerShell profile.


```shell
kosli completion [bash|zsh|fish|powershell]
```

### Flags
| Flag | Description |
| :--- | :--- |
|    -h, --help  |  help for completion  |


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


