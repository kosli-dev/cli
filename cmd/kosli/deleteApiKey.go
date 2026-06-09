package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const deleteApiKeyShortDesc = `Delete one or more API keys for a service account.`

const deleteApiKeyLongDesc = deleteApiKeyShortDesc + `

This permanently deletes the API key(s) identified by KEY-ID. Deletion is immediate and
cannot be undone. You are asked to confirm before the key is deleted; use
^--assume-yes^/^--yes^ to skip the confirmation prompt.`

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
`

type deleteApiKeyOptions struct {
	serviceAccount string
	assumeYes      bool
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
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, cmd.InOrStdin(), args)
		},
	}

	cmd.Flags().StringVarP(&o.serviceAccount, "service-account", "s", "", serviceAccountNameFlag)
	cmd.Flags().BoolVarP(&o.assumeYes, "assume-yes", "y", false, apiKeyAssumeYesFlag)
	// keep --yes as a hidden alias for --assume-yes (bound to the same option)
	cmd.Flags().BoolVar(&o.assumeYes, "yes", false, apiKeyAssumeYesFlag)
	if f := cmd.Flags().Lookup("yes"); f != nil {
		f.Hidden = true
	}
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"service-account"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *deleteApiKeyOptions) run(out io.Writer, in io.Reader, args []string) error {
	if !o.assumeYes && !global.DryRun {
		confirmed, err := confirmApiKeyDeletion(args, o.serviceAccount, out, in)
		if err != nil {
			return err
		}
		if !confirmed {
			logger.Info("Deletion of API key(s) %s was cancelled!", strings.Join(styleApiKeyIDs(out, args), ", "))
			return nil
		}
	}

	// deletion is destructive and one-way: on any failure mid-batch, make clear
	// which keys were already deleted before it.
	reportAlreadyDeleted := func(i int) {
		if i > 0 {
			logger.Info("keys already deleted before this failure: %s", strings.Join(styleApiKeyIDs(out, args[:i]), ", "))
		}
	}

	for i, keyID := range args {
		url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, o.serviceAccount, "api-keys", keyID)
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
			return fmt.Errorf("failed to delete API key %s: %w", keyID, err)
		}
		if !global.DryRun {
			logger.Info("API key %s for service account %s was deleted!", style(out, keyID, ansiBold, ansiCyan), o.serviceAccount)
		}
	}
	return nil
}

// styleApiKeyIDs styles key IDs for user-facing messages (bold cyan when
// styling is enabled for out).
func styleApiKeyIDs(out io.Writer, keyIDs []string) []string {
	styledKeys := make([]string, len(keyIDs))
	for i, keyID := range keyIDs {
		styledKeys[i] = style(out, keyID, ansiBold, ansiCyan)
	}
	return styledKeys
}

// confirmApiKeyDeletion prompts the user to confirm deletion and returns true
// only when the answer is an affirmative "y"/"yes" (case-insensitive). The
// prompt has no trailing newline so the answer is typed on the same line.
func confirmApiKeyDeletion(keyIDs []string, serviceAccount string, out io.Writer, in io.Reader) (bool, error) {
	if _, err := fmt.Fprintf(out, "Are you sure you want to delete API key(s) %s for service account %s? [y/N] ",
		strings.Join(styleApiKeyIDs(out, keyIDs), ", "), style(out, serviceAccount, ansiBold, ansiGreen)); err != nil {
		return false, err
	}

	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}
