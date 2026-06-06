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

const revokeApiKeyShortDesc = `Revoke an API key for a service account.`

const revokeApiKeyLongDesc = revokeApiKeyShortDesc + `

This permanently revokes the API key identified by KEY-ID. You are asked to confirm
before the key is revoked; use ^--yes^ or ^--assume-yes^ to skip the confirmation prompt.`

const revokeApiKeyExample = `
# revoke an API key for a service account (asks for confirmation):
kosli service-account api-keys revoke yourApiKeyID \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# revoke multiple API keys at once:
kosli service-account api-keys revoke keyID1 keyID2 \
	--service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# revoke an API key without confirmation:
kosli service-account api-keys revoke yourApiKeyID \
	--service-account yourServiceAccountName \
	--assume-yes \
	--api-token yourAPIToken \
	--org yourOrgName
`

type revokeApiKeyOptions struct {
	serviceAccount string
	assumeYes      bool
}

func newRevokeApiKeyCmd(out io.Writer) *cobra.Command {
	o := new(revokeApiKeyOptions)
	cmd := &cobra.Command{
		Use:     "revoke KEY-ID [KEY-ID...]",
		Aliases: []string{"re", "del"},
		Short:   revokeApiKeyShortDesc,
		Long:    revokeApiKeyLongDesc,
		Example: revokeApiKeyExample,
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
	cmd.Flags().BoolVarP(&o.assumeYes, "assume-yes", "y", false, revokeApiKeyYesFlag)
	// keep --yes as a hidden alias for --assume-yes (bound to the same option)
	cmd.Flags().BoolVar(&o.assumeYes, "yes", false, revokeApiKeyYesFlag)
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

func (o *revokeApiKeyOptions) run(out io.Writer, in io.Reader, args []string) error {
	if !o.assumeYes && !global.DryRun {
		confirmed, err := confirmApiKeyRevoke(args, o.serviceAccount, out, in)
		if err != nil {
			return err
		}
		if !confirmed {
			logger.Info("revocation of API key(s) %s was cancelled", strings.Join(args, ", "))
			return nil
		}
	}

	for _, keyID := range args {
		url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, o.serviceAccount, "api-keys", keyID)
		if err != nil {
			return err
		}

		reqParams := &requests.RequestParams{
			Method: http.MethodDelete,
			URL:    url,
			DryRun: global.DryRun,
			Token:  global.ApiToken,
		}
		if _, err := kosliClient.Do(reqParams); err != nil {
			// Keep the returned error plain (no ANSI): it may be logged or
			// wrapped by callers that don't expect escape codes. Keys already
			// revoked were each logged in bold green above (user-facing output).
			return fmt.Errorf("failed to revoke API key %s: %w", keyID, err)
		}
		if !global.DryRun {
			logger.Info("API key %s for service account %s was revoked", style(out, keyID, ansiBold, ansiGreen), o.serviceAccount)
		}
	}
	return nil
}

// confirmApiKeyRevoke prompts the user to confirm revocation and returns true
// only when the answer is an affirmative "y"/"yes" (case-insensitive).
func confirmApiKeyRevoke(keyIDs []string, serviceAccount string, out io.Writer, in io.Reader) (bool, error) {
	styledKeys := make([]string, len(keyIDs))
	for i, keyID := range keyIDs {
		styledKeys[i] = style(out, keyID, ansiBold, ansiMagenta)
	}

	logger.Info("Are you sure you want to revoke API key(s) %s for service account %s? [y/N]",
		strings.Join(styledKeys, ", "), style(out, serviceAccount, ansiBold, ansiGreen))

	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}
