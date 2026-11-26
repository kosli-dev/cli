package main

import (
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const archiveRegoPolicyShortDesc = `Archive a rego policy.`

const archiveRegoPolicyLongDesc = archiveRegoPolicyShortDesc + `
An archived rego policy is left out of ^kosli list rego-policies^ unless ^--archived^ is given.
Archiving creates no version: its versions stay readable with ^kosli get rego-policy^, an
evaluation of it still runs, and its identifier stays taken. Only an org admin can archive.
Restore it with ^kosli unarchive rego-policy^.
`

const archiveRegoPolicyExample = `
# archive a rego policy:
kosli archive rego-policy yourPolicyIdentifier \
	--api-token yourAPIToken \
	--org yourOrgName
`

func newArchiveRegoPolicyCmd(out io.Writer) *cobra.Command {
	return newRegoPolicyArchiveStateCmd("archive", archiveRegoPolicyShortDesc,
		archiveRegoPolicyLongDesc, archiveRegoPolicyExample)
}

// newRegoPolicyArchiveStateCmd makes the command that posts to the policy's
// archive or unarchive endpoint; verb is the endpoint's final path segment.
func newRegoPolicyArchiveStateCmd(verb, short, long, example string) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "rego-policy POLICY-IDENTIFIER",
		Short:       short,
		Long:        long,
		Example:     example,
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			policy, err := regopolicies.PolicyURL(global.Host, global.Org, args[0])
			if err != nil {
				return err
			}
			endpoint, err := url.JoinPath(policy, verb)
			if err != nil {
				return err
			}
			_, err = kosliClient.Do(&requests.RequestParams{
				Method: http.MethodPost,
				URL:    endpoint,
				DryRun: global.DryRun,
				Token:  global.ApiToken,
			})
			if err == nil && !global.DryRun {
				logger.Info("rego policy %s was %sd", args[0], verb)
			}
			return err
		},
	}
	addDryRunFlag(cmd)
	return cmd
}
