package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const assertEnvironmentDesc = `Assert the compliance status of an environment in Kosli. Exits with non-zero code if the environment has a non-compliant status.`

func newAssertEnvironmentCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environment ENVIRONMENT-NAME-OR-EXPRESSION",
		Short: assertEnvironmentDesc,
		Long:  assertEnvironmentDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "environment name/expression argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(out, args)
		},
	}
	addDryRunFlag(cmd)

	return cmd
}

func run(out io.Writer, args []string) error {
	var err error

	url := fmt.Sprintf("%s/api/v1/environments/%s/snapshots/%s", global.Host, global.Owner, url.QueryEscape(args[0]))
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken, global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
	if err != nil {
		return err
	}

	var environmentData map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &environmentData)
	if err != nil {
		return err
	}

	if environmentData["compliant"].(bool) {
		fmt.Fprintln(out, "COMPLIANT")
	} else {
		return fmt.Errorf("INCOMPLIANT")
	}

	return nil
}
