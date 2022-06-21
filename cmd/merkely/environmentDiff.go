package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/spf13/cobra"
)

const environmentDiffDesc = `Diff snapshots.`

type environmentDiffOptions struct {
	// long bool
	json bool
}

type EnvironmentDiffPayload struct {
	Snappish1 string `json:"snappish1"`
	Snappish2 string `json:"snappish2"`
}

type EnvironmentDiffResponse struct {
	Sha256 string   `json:"sha256"`
	Name   string   `json:"name"`
	Pods   []string `json:"pods"`
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
	cmd.Flags().BoolVarP(&o.json, "json", "j", false, environmentJsonFlag)

	return cmd
}

func (o *environmentDiffOptions) run(out io.Writer, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("Two snappish required")
	}

	payload := new(EnvironmentDiffPayload)
	payload.Snappish1 = args[0]
	payload.Snappish2 = args[1]

	url := fmt.Sprintf("%s/api/v1/env-diff/%s/", global.Host, global.Owner)
	// response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
	// 	global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	response, err := requests.SendPayload(payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodGet, log)

	if err != nil {
		return fmt.Errorf("kosli server %s is unresponsive", global.Host)
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
		return nil
	}

	var diffs map[string][]EnvironmentDiffResponse
	err = json.Unmarshal([]byte(response.Body), &diffs)
	if err != nil {
		return err
	}

	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"

	fmt.Print(colorRed)
	for _, entry := range diffs["-"] {
		fmt.Printf("- %s\n", entry.Name)
		fmt.Printf("  %s\n", entry.Sha256)
	}
	fmt.Print(colorReset)
	fmt.Println()
	fmt.Print(colorGreen)
	for _, entry := range diffs["+"] {
		fmt.Printf("+ %s\n", entry.Name)
		fmt.Printf("  %s\n", entry.Sha256)
	}
	fmt.Print(colorReset)
	return nil
}
