package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/spf13/cobra"
)

const publishRegoPolicyShortDesc = `Publish a new version of a rego policy in a Kosli org.`

const publishRegoPolicyLongDesc = publishRegoPolicyShortDesc + `
The policy is addressed by IDENTIFIER, which decisions cite and which cannot be
changed after the first publish. ^--name^ is required on the first publish only.
Only an org admin can publish.

PATH is a Rego file or a directory on this machine. A directory is sent as one
bundle, by the same rules as ^kosli evaluate policy --policy^: files in
dot-directories are left out, and the bundle can hold at most 100 files and
1 MiB.

Each publish stores the next version, with ^--comment^ kept on it. A change to
the files or to ^--input-schema^ makes a new version. When the files and the
input schema are the same as the newest version, and that version is pending or
valid, nothing new is stored and that version is reported. A change to only
^--name^ or ^--description^ makes no new version.

Kosli validates each version after it is stored, and the command waits up to 30
seconds for the result:
  - valid: the version can be evaluated. The command exits 0.
  - rejected: the policy is at fault, for example it does not compile. The
    command prints the reason and exits 1.
  - failed: Kosli could not validate the version. Publish again to retry. The
    command exits 1.
If the wait expires, the command exits 1 and says how to read the version later.
With ^--output json^, the version is always printed: with the reason for a
rejected or failed one, and as pending when the wait expires.

An evaluation runs the highest-numbered valid version, so a rejected or failed
version leaves the policy in service unchanged.`

const publishRegoPolicyExample = `
# publish the first version of a rego policy from a directory:
kosli publish rego-policy pr-approved ./policies/pr-approved/ \
	--input-schema trail \
	--name "Pull request approved" \
	--description "Every change reached main through an approved PR" \
	--comment "Initial version" \
	--api-token yourAPIToken \
	--org yourOrgName

# publish a new version of an existing rego policy from one file:
kosli publish rego-policy pr-approved ./policies/pr-approved.rego \
	--input-schema trail \
	--comment "Also require two approvers" \
	--api-token yourAPIToken \
	--org yourOrgName

# publish a policy written against a set of trails, and print the version as JSON:
kosli publish rego-policy release-ready ./policies/release-ready/ \
	--input-schema trails \
	--name "Release ready" \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName

# show what would be published, without sending it:
kosli publish rego-policy pr-approved ./policies/pr-approved/ \
	--input-schema trail \
	--dry-run \
	--api-token yourAPIToken \
	--org yourOrgName
`

// A variable so that tests can shorten the wait.
var regoPolicyWaitOptions = regopolicies.WaitOptions{}

var regoPolicyInputSchemas = []string{"trail", "trails"}

type publishRegoPolicyOptions struct {
	name        string
	description string
	inputSchema string
	comment     string
	output      string
}

func newPublishRegoPolicyCmd(out io.Writer) *cobra.Command {
	o := new(publishRegoPolicyOptions)
	cmd := &cobra.Command{
		Use:     "rego-policy IDENTIFIER PATH",
		Short:   publishRegoPolicyShortDesc,
		Long:    publishRegoPolicyLongDesc,
		Example: publishRegoPolicyExample,
		Args:    cobra.ExactArgs(2),
		// Hidden while the server gates the endpoint behind is-server-side-evaluation-enabled.
		Hidden:      true,
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args[0], args[1])
		},
	}

	cmd.Flags().StringVar(&o.inputSchema, "input-schema", "", regoPolicyInputSchemaFlag)
	cmd.Flags().StringVar(&o.name, "name", "", regoPolicyNameFlag)
	cmd.Flags().StringVar(&o.description, "description", "", regoPolicyDescriptionFlag)
	cmd.Flags().StringVar(&o.comment, "comment", "", regoPolicyCommentFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	addDryRunFlag(cmd)

	err := RequireFlags(cmd, []string{"input-schema"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *publishRegoPolicyOptions) run(out io.Writer, identifier, path string) error {
	// Checked before publishing, so a bad format cannot fail a command whose version is stored.
	printVersion, known := regoPolicyPrinters[o.output]
	if !known {
		return fmt.Errorf("unsupported output format: %s. Valid formats are: [table, json]", o.output)
	}
	if !slices.Contains(regoPolicyInputSchemas, o.inputSchema) {
		return fmt.Errorf("--input-schema must be one of %v, got %q", regoPolicyInputSchemas, o.inputSchema)
	}
	// Decisions cite a published version, so its source must be on disk, not behind a URL.
	if isRemotePolicyRef(path) {
		return fmt.Errorf("PATH takes a file or a directory on this machine, not a URL")
	}

	files, err := policyBundle([]string{path})
	if err != nil {
		return err
	}

	client := regopolicies.NewClient(kosliClient, global.Host, global.ApiToken, global.DryRun)
	published, err := client.Publish(global.Org, regopolicies.PublishRequest{
		Identifier:  identifier,
		Name:        o.name,
		Description: o.description,
		InputSchema: o.inputSchema,
		Comment:     o.comment,
		Files:       files,
	})
	if err != nil {
		return err
	}
	if published == nil {
		// Nil on a dry run.
		return nil
	}

	// Publishing the content of the newest version answers with that version, which can be terminal.
	version := published
	if !version.IsTerminal() {
		version, err = client.WaitForTerminal(context.Background(), global.Org, identifier,
			published.Version, regoPolicyWaitOptions)
		if err != nil {
			if printErr := o.printIfJSON(out, published); printErr != nil {
				return printErr
			}
			var stillPending *regopolicies.StillPendingError
			if errors.As(err, &stillPending) {
				return err
			}
			return fmt.Errorf("rego policy '%s' version %d was published, but could not be read back: %w",
				identifier, published.Version, err)
		}
	}
	if version.Status != regopolicies.StatusValid {
		if err := o.printIfJSON(out, version); err != nil {
			return err
		}
		return notValidError(identifier, version)
	}

	// A new version is stored pending and with the comment it was sent, so a
	// terminal answer, or one with another comment, is an existing version.
	commentDropped := o.comment != "" && published.Comment != o.comment
	note := ""
	if published.IsTerminal() || commentDropped {
		note = "nothing new was stored, as the files and input schema are the same as this version"
		if commentDropped {
			note += ", so --comment was not kept"
		}
	}
	return printVersion(out, version, note)
}

// printIfJSON gives a script the version even when the command fails, as JSON
// carries what the error text does not, such as the version number.
func (o *publishRegoPolicyOptions) printIfJSON(out io.Writer, version *regopolicies.Version) error {
	if o.output != "json" {
		return nil
	}
	return regoPolicyPrinters["json"](out, version, "")
}

// The JSON printer ignores note, so a script reads only the fields of Version.
var regoPolicyPrinters = map[string]func(out io.Writer, version *regopolicies.Version, note string) error{
	"table": func(out io.Writer, version *regopolicies.Version, note string) error {
		line := fmt.Sprintf("rego policy '%s' version %d is valid (%s)",
			version.Identifier, version.Version, version.Digest)
		if note != "" {
			line += "; " + note
		}
		_, err := fmt.Fprintln(out, line)
		return err
	},
	"json": func(out io.Writer, version *regopolicies.Version, _ string) error {
		raw, err := json.Marshal(version)
		if err != nil {
			return err
		}
		return output.PrintJson(string(raw), out, 0)
	},
}

// notValidError words a rejected version as the policy's fault and a failed one as Kosli's.
func notValidError(identifier string, version *regopolicies.Version) error {
	subject := fmt.Sprintf("rego policy '%s' version %d", identifier, version.Version)
	switch version.Status {
	case regopolicies.StatusRejected:
		if version.Error == nil {
			return fmt.Errorf("%s was rejected, and Kosli gave no reason", subject)
		}
		return fmt.Errorf("%s was rejected (%s): %s", subject, version.Error.Kind, version.Error.Message)
	case regopolicies.StatusFailed:
		reason := ", with no reason given"
		if version.Error != nil {
			reason = fmt.Sprintf(" (%s): %s", version.Error.Kind, version.Error.Message)
		}
		return fmt.Errorf("validation of %s failed on the Kosli side%s; "+
			"this is not a fault in the policy, publish again to retry", subject, reason)
	}
	return fmt.Errorf("%s is %s", subject, version.Status)
}
