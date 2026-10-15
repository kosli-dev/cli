package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/spf13/cobra"
)

const getEnvironmentDesc = `Get an environment's metadata.`

type getEnvironmentOptions struct {
	output string
}

func newGetEnvironmentCmd(out io.Writer) *cobra.Command {
	o := new(getEnvironmentOptions)
	cmd := &cobra.Command{
		Use:     "environment ENVIRONMENT-NAME",
		Aliases: []string{"env"},
		Short:   getEnvironmentDesc,
		Long:    getEnvironmentDesc,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(cmd.Context(), out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *getEnvironmentOptions) run(ctx context.Context, out io.Writer, args []string) error {
	environment, err := kosliAPI().GetEnvironment(ctx, args[0])
	if err != nil {
		return err
	}

	return output.FormattedPrint(string(environment.Raw), o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printEnvironmentAsTable,
			"json":  output.PrintJson,
		})
}

func printEnvironmentAsTable(raw string, out io.Writer, page int) error {
	var env map[string]any
	err := json.Unmarshal([]byte(raw), &env)
	if err != nil {
		return err
	}

	lastReportedAt, err := formattedTimestamp(env["last_reported_at"], false)
	if err != nil {
		return err
	}

	state := "N/A"
	if env["state"] != nil && env["state"].(bool) {
		state = "COMPLIANT"
	} else if env["state"] != nil {
		state = "NON-COMPLIANT"
	}

	tagsOutput := formatTags(env["tags"])
	if tagsOutput == "" {
		tagsOutput = "None"
	}

	policies := []any{}

	if env["policies"] != nil {
		policies = env["policies"].([]any)
	}

	header := []string{}
	rows := []string{}
	rows = append(rows, fmt.Sprintf("Name:\t%s", env["name"]))
	rows = append(rows, fmt.Sprintf("Type:\t%s", env["type"]))
	rows = append(rows, fmt.Sprintf("Description:\t%s", env["description"]))
	rows = append(rows, fmt.Sprintf("State:\t%s", state))
	rows = append(rows, fmt.Sprintf("Last Reported At:\t%s", lastReportedAt))
	rows = append(rows, fmt.Sprintf("Tags:\t%s", tagsOutput))
	rows = append(rows, fmt.Sprintf("Policies:\t%s", policies))

	tabFormattedPrint(out, header, rows)

	return nil
}
