package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const artifactGetDesc = `Get artifact from specified pipeline`

type artifactGetOptions struct {
	json         bool
	pipelineName string
}

func newArtifactGetCmd(out io.Writer) *cobra.Command {
	o := new(artifactGetOptions)
	cmd := &cobra.Command{
		Use:   "get ARTIFACT-DIGEST",
		Short: artifactGetDesc,
		Long:  artifactGetDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorAfterPrintingHelp(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorAfterPrintingHelp(cmd, "pipeline name argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().BoolVarP(&o.json, "json", "j", false, jsonOutputFlag)

	err := RequireFlags(cmd, []string{"pipeline"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *artifactGetOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s", global.Host, global.Owner, o.pipelineName, args[0])
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return err
	}

	if o.json {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
		return nil
	}

	approvalUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/approvals/", global.Host, global.Owner, o.pipelineName, args[0])
	approvalResponse, err := requests.DoBasicAuthRequest([]byte{}, approvalUrl, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return err
	}

	var artifact map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &artifact)
	if err != nil {
		return err
	}

	var approvals []map[string]interface{}
	err = json.Unmarshal([]byte(approvalResponse.Body), &approvals)
	if err != nil {
		return err
	}

	artifactData := artifact["evidence"].(map[string]interface{})["artifact"].(map[string]interface{})

	rows := []string{}
	rows = append(rows, fmt.Sprintf("Name:\t%s", artifactData["filename"].(string)))
	rows = append(rows, fmt.Sprintf("State:\t%s", artifact["state"].(string)))
	rows = append(rows, fmt.Sprintf("Git commit:\t%s", artifactData["git_commit"].(string)))
	rows = append(rows, fmt.Sprintf("Build URL:\t%s", artifactData["build_url"].(string)))
	rows = append(rows, fmt.Sprintf("Commit URL:\t%s", artifactData["commit_url"].(string)))
	createdAt, err := formattedTimestamp(artifactData["logged_at"])
	if err != nil {
		return err
	}
	rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAt))
	rows = append(rows, "Approvals:")
	for _, approval := range approvals {
		timestamp, err := formattedTimestamp(approval["last_modified_at"])
		if err != nil {
			return err
		}
		approvalRow := fmt.Sprintf("\t#%d  %s  Last modified: %s", int64(approval["release_number"].(float64)), approval["state"].(string), timestamp)
		rows = append(rows, approvalRow)
	}

	printTable(out, []string{}, rows)
	return nil
}
