package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"

	"github.com/kosli-dev/cli/internal/docgen"
	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/spf13/cobra"
)

const createRegoPolicyShortDesc = `Publish a new version of a rego policy in a Kosli org.`

const createRegoPolicyLongDesc = createRegoPolicyShortDesc + `
The policy is addressed by IDENTIFIER, which cannot be changed after the first
publish. PATH is a Rego file or a directory, sent as one bundle in the same way
as ^kosli evaluate policy --policy^ sends it.

Kosli validates the new version after it is stored. The command waits for that
result, and exits 0 only when the version is valid.`

const createRegoPolicyExample = `
# publish the first version of a rego policy from a directory:
kosli create rego-policy pr-approved ./policies/pr-approved/ \
	--input-schema trail \
	--name "Pull request approved" \
	--description "Every change reached main through an approved PR" \
	--comment "Initial version" \
	--api-token yourAPIToken \
	--org yourOrgName
`

// regoPolicyWaitOptions is how long validation is waited for, and the seam a
// test shortens it through.
var regoPolicyWaitOptions = regopolicies.WaitOptions{}

var regoPolicyInputSchemas = []string{"trail", "trails"}

type createRegoPolicyOptions struct {
	name        string
	description string
	inputSchema string
	comment     string
	output      string
}

func newCreateRegoPolicyCmd(out io.Writer) *cobra.Command {
	o := new(createRegoPolicyOptions)
	cmd := &cobra.Command{
		Use:     "rego-policy IDENTIFIER PATH",
		Short:   createRegoPolicyShortDesc,
		Long:    createRegoPolicyLongDesc,
		Example: createRegoPolicyExample,
		Args:    cobra.ExactArgs(2),
		// Hidden while the server endpoint is behind a feature flag.
		Hidden:      true,
		Annotations: map[string]string{docgen.DocHiddenAnnotation: "", betaCLIAnnotation: ""},
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

func (o *createRegoPolicyOptions) run(out io.Writer, identifier, path string) error {
	// Refused before publishing: a format refused where it is printed would
	// leave a version published by a command that reported an error.
	printVersion, known := regoPolicyPrinters[o.output]
	if !known {
		return fmt.Errorf("unsupported output format: %s. Valid formats are: [table, json]", o.output)
	}
	if !slices.Contains(regoPolicyInputSchemas, o.inputSchema) {
		return fmt.Errorf("--input-schema must be one of %v, got %q", regoPolicyInputSchemas, o.inputSchema)
	}
	// A published version is cited by decisions, so its source must be what
	// the author has on disk, not whatever a URL serves at the moment.
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
		// A dry run sent nothing, so there is no version to wait for.
		return nil
	}

	version, err := client.WaitForTerminal(context.Background(), global.Org, identifier,
		published.Version, regoPolicyWaitOptions)
	if err != nil {
		return err
	}
	if version.Status != regopolicies.StatusValid {
		return notValidError(identifier, version)
	}

	return printVersion(out, version)
}

var regoPolicyPrinters = map[string]func(io.Writer, *regopolicies.Version) error{
	"table": func(out io.Writer, version *regopolicies.Version) error {
		_, err := fmt.Fprintf(out, "rego policy '%s' version %d is valid (%s)\n",
			version.Identifier, version.Version, version.Digest)
		return err
	},
	"json": func(out io.Writer, version *regopolicies.Version) error {
		raw, err := json.Marshal(version)
		if err != nil {
			return err
		}
		return output.PrintJson(string(raw), out, 0)
	},
}

// notValidError reports a version that will never be evaluated. A rejected
// version is the policy's fault and a failed one is Kosli's, so the two are
// worded apart: the author of a failed version has nothing to fix.
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
