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

const deleteServiceAccountShortDesc = `Delete one or more service accounts.`

const deleteServiceAccountLongDesc = deleteServiceAccountShortDesc + `

This permanently removes the service account(s) identified by SERVICE-ACCOUNT-NAME
from the organization, along with their API keys. Deletion is immediate and
cannot be undone. You are asked to confirm before deletion; use
^--assume-yes^/^--yes^ to skip the confirmation prompt.`

const deleteServiceAccountExample = `
# delete a service account (asks for confirmation):
kosli delete service-account yourServiceAccountName \
	--api-token yourAPIToken \
	--org yourOrgName

# delete multiple service accounts at once:
kosli delete service-account sa1 sa2 \
	--api-token yourAPIToken \
	--org yourOrgName

# delete a service account without confirmation:
kosli delete service-account yourServiceAccountName \
	--assume-yes \
	--api-token yourAPIToken \
	--org yourOrgName
`

type deleteServiceAccountOptions struct {
	assumeYes bool
}

func newDeleteServiceAccountCmd(out io.Writer) *cobra.Command {
	o := new(deleteServiceAccountOptions)
	cmd := &cobra.Command{
		Use:     "service-account SERVICE-ACCOUNT-NAME [SERVICE-ACCOUNT-NAME...]",
		Aliases: []string{"sa"},
		Short:   deleteServiceAccountShortDesc,
		Long:    deleteServiceAccountLongDesc,
		Example: deleteServiceAccountExample,
		Args:    cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireGlobalFlags(global, []string{"Org", "ApiToken"}); err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd.InOrStdin(), args)
		},
	}

	cmd.Flags().BoolVarP(&o.assumeYes, "assume-yes", "y", false, serviceAccountAssumeYesFlag)
	// keep --yes as a hidden alias for --assume-yes (bound to the same option)
	cmd.Flags().BoolVar(&o.assumeYes, "yes", false, serviceAccountAssumeYesFlag)
	if f := cmd.Flags().Lookup("yes"); f != nil {
		f.Hidden = true
	}
	addDryRunFlag(cmd)

	return cmd
}

func (o *deleteServiceAccountOptions) run(in io.Reader, args []string) error {
	if !o.assumeYes && !global.DryRun {
		confirmed, err := confirmServiceAccountDeletion(args, in)
		if err != nil {
			return err
		}
		if !confirmed {
			logger.Info("Deletion of service account(s) %s was cancelled.", strings.Join(styleServiceAccountNames(args), ", "))
			return nil
		}
	}

	// deletion is destructive and one-way: on any failure mid-batch, make clear
	// which service accounts were already deleted before it.
	reportAlreadyDeleted := func(i int) {
		if i > 0 {
			logger.Info("Service accounts already deleted before this failure: %s", strings.Join(styleServiceAccountNames(args[:i]), ", "))
		}
	}

	for i, name := range args {
		url, err := url.JoinPath(global.Host, "api/v2/service-accounts", global.Org, name)
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
			return fmt.Errorf("failed to delete service account: %w", err)
		}
		if !global.DryRun {
			logger.Info("service account %s was deleted!", style(logger.Out, name, ansiBold, ansiCyan))
		}
	}
	return nil
}

// styleServiceAccountNames styles service account names for user-facing
// messages printed via logger (bold cyan when styling is enabled).
func styleServiceAccountNames(names []string) []string {
	styled := make([]string, len(names))
	for i, name := range names {
		styled[i] = style(logger.Out, name, ansiBold, ansiCyan)
	}
	return styled
}

// confirmServiceAccountDeletion prompts the user to confirm deletion and
// returns true only when the answer is an affirmative "y"/"yes"
// (case-insensitive). The prompt has no trailing newline so the answer is
// typed on the same line.
func confirmServiceAccountDeletion(names []string, in io.Reader) (bool, error) {
	logger.Print("Are you sure you want to delete service account(s) %s? [y/N] ",
		strings.Join(styleServiceAccountNames(names), ", "))

	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && err != io.EOF {
		return false, err
	}

	answer = strings.ToLower(strings.TrimSpace(answer))
	return answer == "y" || answer == "yes", nil
}
