package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const environmentDiffDesc = `Diff snapshots.`

type environmentDiffOptions struct {
	// long bool
	// json bool
}

type EnvironmentDiffPayload struct {
	Snappish1 string `json:"snappish1"`
	Snappish2 string `json:"snappish2"`
}

func newEnvironmentDiffCmd(out io.Writer) *cobra.Command {
	o := new(environmentDiffOptions)
	cmd := &cobra.Command{
		Use:   "diff [ENVIRONMENT-NAME]",
		Short: environmentDiffDesc,
		Long:  environmentDiffDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	// cmd.Flags().BoolVarP(&o.long, "long", "l", false, environmentLongFlag)
	// cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)

	return cmd
}

func (o *environmentDiffOptions) run(out io.Writer, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("At least one snappish required")
	}

	payload := new(EnvironmentDiffPayload)
	payload.Snappish1 = args[0] + "^0"
	payload.Snappish2 = args[0] + "^1"

	url := fmt.Sprintf("%s/api/v1/env-diff/%s/", global.Host, global.Owner)
	// response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
	// 	global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	response, err := requests.SendPayload(payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodGet, log)

	if err != nil {
		return fmt.Errorf("kosli server %s is unresponsive", global.Host)
	}

	pj, err := prettyJson(response.Body)
	if err != nil {
		return err
	}
	fmt.Println(pj)
	return nil
}
