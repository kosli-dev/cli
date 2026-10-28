package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const deleteApiKeyShortDesc = `Delete one or more API keys for a service account, or your own personal API keys.`

const deleteApiKeyLongDesc = deleteApiKeyShortDesc + `

This permanently deletes the API key(s) identified by KEY-ID. Deletion is immediate and
cannot be undone. You are asked to confirm before the key is deleted; use
^--assume-yes^/^--yes^ to skip the confirmation prompt.

When stdin is not interactive (e.g. in CI) the prompt cannot be answered and the
command fails without deleting anything, so pass ^--assume-yes^ there.

Deleting a key also deletes every key that was created with it. With ^--personal^ and an
API key as ^--api-token^, you can delete only that key and keys created with it.`

const deleteApiKeyExample = `
# delete an API key for a service account (asks for confirmation):
kosli delete api-key yourApiKeyID \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# delete multiple API keys at once:
kosli delete api-key keyID1 keyID2 \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# delete an API key without confirmation:
kosli delete api-key yourApiKeyID \
	--service-account yourServiceAccountName \
	--assume-yes \
	--api-token yourAPIToken \
	--org yourOrgName

# delete one of your own personal API keys:
kosli delete api-key yourApiKeyID \
	--personal \
	--api-token yourAPIToken
`

type deleteApiKeyOptions struct {
	owner     apiKeyOwner
	assumeYes bool
}

func newDeleteApiKeyCmd(out io.Writer) *cobra.Command {
	o := new(deleteApiKeyOptions)
	cmd := &cobra.Command{
		Use:     "api-key KEY-ID [KEY-ID...]",
		Aliases: []string{"ak"},
		Short:   deleteApiKeyShortDesc,
		Long:    deleteApiKeyLongDesc,
		Example: deleteApiKeyExample,
		Args:    cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, o.owner.requiredGlobalFlags()); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd.InOrStdin(), args)
		},
	}

	o.owner.addFlags(cmd)
	cmd.Flags().BoolVarP(&o.assumeYes, "assume-yes", "y", false, apiKeyAssumeYesFlag)
	// keep --yes as a hidden alias for --assume-yes (bound to the same option)
	cmd.Flags().BoolVar(&o.assumeYes, "yes", false, apiKeyAssumeYesFlag)
	if f := cmd.Flags().Lookup("yes"); f != nil {
		f.Hidden = true
	}
	addDryRunFlag(cmd)

	return cmd
}

func (o *deleteApiKeyOptions) run(in io.Reader, args []string) error {
	if !o.assumeYes && !global.DryRun {
		confirmed, err := confirmApiKeyDeletion(args, o.owner, in)
		if err != nil {
			return err
		}
		if !confirmed {
			logger.Info("Deletion of API key(s) %s was cancelled.", strings.Join(styleApiKeyIDs(args), ", "))
			return nil
		}
	}

	// deletion is destructive and one-way: on any failure mid-batch, make clear
	// which keys were already deleted before it.
	reportAlreadyDeleted := func(i int) {
		if i > 0 {
			logger.Info("Keys already deleted before this failure: %s", strings.Join(styleApiKeyIDs(args[:i]), ", "))
		}
	}

	for i, keyID := range args {
		url, err := o.owner.url(keyID)
		if err != nil {
			reportAlreadyDeleted(i)
			return err
		}

		reqParams := &requests.RequestParams{
			Method: http.MethodDelete,
			URL:    url,
			DryRun: global.DryRun,
			Token:  global.ApiToken,
		}
		if _, err := kosliClient.Do(reqParams); err != nil {
			reportAlreadyDeleted(i)
			return fmt.Errorf("failed to delete API key: %w", err)
		}
		if !global.DryRun {
			styledKey := style(logger.Out, keyID, ansiBold, ansiCyan)
			if o.owner.personal {
				logger.Info("Personal API key %s was deleted!", styledKey)
			} else {
				logger.Info("API key %s for service account %s was deleted!", styledKey, o.owner.serviceAccount)
			}
		}
	}
	return nil
}

// styleApiKeyIDs styles key IDs for user-facing messages printed via logger
// (bold cyan when styling is enabled for the logger's info stream).
func styleApiKeyIDs(keyIDs []string) []string {
	styledKeys := make([]string, len(keyIDs))
	for i, keyID := range keyIDs {
		styledKeys[i] = style(logger.Out, keyID, ansiBold, ansiCyan)
	}
	return styledKeys
}

// confirmApiKeyDeletion prompts the user to confirm deletion and returns true
// only when the answer is an affirmative "y"/"yes" (case-insensitive). It errors
// when the prompt cannot be answered at all, see confirmDeletion.
func confirmApiKeyDeletion(keyIDs []string, owner apiKeyOwner, in io.Reader) (bool, error) {
	keys := strings.Join(styleApiKeyIDs(keyIDs), ", ")
	prompt := fmt.Sprintf("Are you sure you want to delete your personal API key(s) %s? [y/N] ", keys)
	if !owner.personal {
		prompt = fmt.Sprintf("Are you sure you want to delete API key(s) %s for service account %s? [y/N] ",
			keys, style(logger.Out, owner.serviceAccount, ansiBold, ansiGreen))
	}
	return confirmDeletion(prompt, in)
}
