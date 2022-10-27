package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd(out io.Writer) *cobra.Command {
	rootCommandName := "kosli"
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: fmt.Sprintf(`To load completions:

  ### Bash

`+"```"+`
  $ source <(%[1]s completion bash)
`+"```"+`
  To load completions for each session, execute once:  

  On Linux:
  `+"```"+`
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  `+"```"+` 
  On macOS:
  `+"```"+`
  $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s
  `+"```"+`
  ### Zsh

  If shell completion is not already enabled in your environment,  
you will need to enable it.  You can execute the following once:
  `+"```"+`
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  `+"```"+`
  To load completions for each session, execute once:
  `+"```"+`
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"
  `+"```"+`
  You will need to start a new shell for this setup to take effect.

  ### fish
  `+"```"+`
  $ %[1]s completion fish | source
  `+"```"+`
  To load completions for each session, execute once:
  `+"```"+` 
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish
  `+"```"+`
  ### PowerShell
  `+"```"+`
  PS> %[1]s completion powershell | Out-String | Invoke-Expression
  `+"```"+`
 To load completions for every new session, run:
 `+"```"+`
 PS> %[1]s completion powershell > %[1]s.ps1
 `+"```"+` 
 and source this file from your PowerShell profile.
`, rootCommandName),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				err := cmd.Root().GenBashCompletion(os.Stdout)
				if err != nil {
					return err
				}
			case "zsh":
				err := cmd.Root().GenZshCompletion(os.Stdout)
				if err != nil {
					return err
				}
			case "fish":
				err := cmd.Root().GenFishCompletion(os.Stdout, true)
				if err != nil {
					return err
				}
			case "powershell":
				err := cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	return cmd
}
