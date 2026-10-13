package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getRegoPolicyShortDesc = `Get a rego policy, or one version of it.`

const getRegoPolicyLongDesc = getRegoPolicyShortDesc + `
Without ^--policy-version^, the policy is shown with its newest version and its current
version. The current version is the highest-numbered valid version, which is the
one an evaluation runs, so it can be older than the newest version.

With ^--policy-version^, that version is shown with its status, its digest, who
published it and when, and the reason when it is rejected or failed. The command
exits 0 whatever the status of the version.`

const getRegoPolicyExample = `
# get a rego policy and its current version:
kosli get rego-policy pr-approved \
	--api-token yourAPIToken \
	--org yourOrgName

# get version 3 of a rego policy:
kosli get rego-policy pr-approved \
	--policy-version 3 \
	--api-token yourAPIToken \
	--org yourOrgName

# get version 3 of a rego policy as JSON:
kosli get rego-policy pr-approved \
	--policy-version 3 \
	--output json \
	--api-token yourAPIToken \
	--org yourOrgName
`

type getRegoPolicyOptions struct {
	output  string
	version int
}

func newGetRegoPolicyCmd(out io.Writer) *cobra.Command {
	o := new(getRegoPolicyOptions)
	cmd := &cobra.Command{
		Use:     "rego-policy IDENTIFIER",
		Short:   getRegoPolicyShortDesc,
		Long:    getRegoPolicyLongDesc,
		Example: getRegoPolicyExample,
		Args:    cobra.ExactArgs(1),
		// Hidden while the server gates the endpoint behind is-server-side-evaluation-enabled.
		Hidden:      true,
		Annotations: map[string]string{betaCLIAnnotation: ""},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if cmd.Flags().Changed("policy-version") && o.version < 1 {
				return ErrorBeforePrintingUsage(cmd, "--policy-version must be a positive integer")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args[0])
		},
	}

	cmd.Flags().IntVar(&o.version, "policy-version", 0, regoPolicyVersionFlag)
	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *getRegoPolicyOptions) run(out io.Writer, identifier string) error {
	var endpoint string
	var err error
	var printTable output.FormatOutputFunc
	if o.version > 0 {
		endpoint, err = regopolicies.VersionURL(global.Host, global.Org, identifier, o.version)
		printTable = printRegoPolicyVersionAsTable
	} else {
		endpoint, err = regopolicies.PolicyURL(global.Host, global.Org, identifier)
		printTable = printRegoPolicyAsTable
	}
	if err != nil {
		return err
	}

	response, err := kosliClient.Do(&requests.RequestParams{
		Method: http.MethodGet,
		URL:    endpoint,
		Token:  global.ApiToken,
	})
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printTable,
			"json":  output.PrintJson,
		})
}

func printRegoPolicyAsTable(raw string, out io.Writer, page int) error {
	var policy map[string]any
	if err := json.Unmarshal([]byte(raw), &policy); err != nil {
		return err
	}

	rows := []string{
		fmt.Sprintf("Identifier:\t%s", policy["identifier"]),
		fmt.Sprintf("Name:\t%s", policy["name"]),
	}
	if description, ok := policy["description"]; ok && description != nil {
		rows = append(rows, fmt.Sprintf("Description:\t%s", description))
	}
	rows = append(rows, fmt.Sprintf("Created by:\t%s", policy["created_by"]))
	for _, field := range []struct{ label, key string }{
		{"Created at", "created_at"},
		{"Last updated", "updated_at"},
	} {
		at, err := formattedTimestamp(policy[field.key], true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("%s:\t%s", field.label, at))
	}
	if newest, ok := policy["newest_version"].(map[string]any); ok {
		rows = append(rows, fmt.Sprintf("Newest version:\t%.0f (%s)", newest["version"], newest["status"]))
	}

	current, ok := policy["current_version"].(map[string]any)
	if !ok {
		rows = append(rows, "Current version:\tnone, as no version is valid yet")
		tabFormattedPrint(out, []string{}, rows)
		return nil
	}
	rows = append(rows,
		fmt.Sprintf("Current version:\t%.0f", current["version"]),
		fmt.Sprintf("  Digest:\t%s", current["digest"]),
		fmt.Sprintf("  Input schema:\t%s", current["input_schema"]),
	)
	if schema, ok := current["params_schema"]; ok && schema != nil {
		compact, err := json.Marshal(schema)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("  Params schema:\t%s", compact))
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}

func printRegoPolicyVersionAsTable(raw string, out io.Writer, page int) error {
	var version regopolicies.Version
	if err := json.Unmarshal([]byte(raw), &version); err != nil {
		return err
	}

	rows := []string{
		fmt.Sprintf("Identifier:\t%s", version.Identifier),
		fmt.Sprintf("Version:\t%d", version.Version),
		fmt.Sprintf("Status:\t%s", version.Status),
	}
	if version.Error != nil {
		rows = append(rows, fmt.Sprintf("Error:\t%s: %s", version.Error.Kind, version.Error.Message))
	}
	rows = append(rows,
		fmt.Sprintf("Digest:\t%s", version.Digest),
		fmt.Sprintf("Input schema:\t%s", version.InputSchema),
	)
	if version.Comment != "" {
		rows = append(rows, fmt.Sprintf("Comment:\t%s", version.Comment))
	}
	rows = append(rows, fmt.Sprintf("Published by:\t%s", version.CreatedBy))
	publishedAt, err := formattedTimestamp(version.CreatedAt, true)
	if err != nil {
		return err
	}
	rows = append(rows, fmt.Sprintf("Published at:\t%s", publishedAt))
	if version.ValidatedAt != nil {
		validatedAt, err := formattedTimestamp(*version.ValidatedAt, true)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Validated at:\t%s", validatedAt))
	}
	if version.OPAVersion != "" {
		rows = append(rows, fmt.Sprintf("OPA version:\t%s", version.OPAVersion))
	}
	if version.Entrypoint != "" {
		rows = append(rows, fmt.Sprintf("Entrypoint:\t%s", version.Entrypoint))
	}
	if version.ParamsSchema != nil {
		compact, err := json.Marshal(version.ParamsSchema)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Params schema:\t%s", compact))
	}

	tabFormattedPrint(out, []string{}, rows)
	return nil
}
