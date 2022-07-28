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

const artifactGetDesc = `Get artifact from a specified pipeline`

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
				return ErrorAfterPrintingHelp(cmd, "artifact digest argument is required")
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

	approvalsUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/approvals/", global.Host, global.Owner, o.pipelineName, args[0])
	approvalsResponse, err := requests.DoBasicAuthRequest([]byte{}, approvalsUrl, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return err
	}

	deploymentsUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/deployments/", global.Host, global.Owner, o.pipelineName, args[0])
	deploymentsResponse, err := requests.DoBasicAuthRequest([]byte{}, deploymentsUrl, "", global.ApiToken,
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
	err = json.Unmarshal([]byte(approvalsResponse.Body), &approvals)
	if err != nil {
		return err
	}

	var deployments []map[string]interface{}
	err = json.Unmarshal([]byte(deploymentsResponse.Body), &deployments)
	if err != nil {
		return err
	}

	evidenceMap := artifact["evidence"].(map[string]interface{})
	artifactData := evidenceMap["artifact"].(map[string]interface{})

	rows := []string{}
	rows = append(rows, fmt.Sprintf("Name:\t%s", artifactData["filename"].(string)))
	rows = append(rows, fmt.Sprintf("State:\t%s", artifact["state"].(string)))
	rows = append(rows, fmt.Sprintf("Git commit:\t%s", artifactData["git_commit"].(string)))
	rows = append(rows, fmt.Sprintf("Build URL:\t%s", artifactData["build_url"].(string)))
	rows = append(rows, fmt.Sprintf("Commit URL:\t%s", artifactData["commit_url"].(string)))
	createdAt, err := formattedTimestamp(artifactData["logged_at"], false)
	if err != nil {
		return err
	}
	rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAt))

	if len(approvals) > 0 {
		rows = append(rows, "Approvals:")
		for _, approval := range approvals {
			timestamp, err := formattedTimestamp(approval["last_modified_at"], true)
			if err != nil {
				return err
			}
			approvalRow := fmt.Sprintf("\t#%d  %s  Last modified: %s", int64(approval["release_number"].(float64)), approval["state"].(string), timestamp)
			rows = append(rows, approvalRow)
		}
	} else {
		rows = append(rows, "Approvals:\tNone")
	}

	if len(deployments) > 0 {
		rows = append(rows, "Deployments:")
		for _, deployment := range deployments {
			deploymentState := deployment["running_state"].(map[string]interface{})
			state := deploymentState["state"].(string)
			stateTimestamp, err := formattedTimestamp(deploymentState["timestamp"], true)
			if err != nil {
				return err
			}

			stateString := "Runtime state unknown"
			if state == "deploying" {
				stateString = "Deploying"
			} else if state == "running" {
				stateString = fmt.Sprintf("Running since %s", stateTimestamp)
			} else if state == "exited" {
				stateString = fmt.Sprintf("Exited on %s", stateTimestamp)
			}

			createdAtTimestamp, err := formattedTimestamp(deployment["created_at"], true)
			if err != nil {
				return err
			}

			deploymentRow := fmt.Sprintf("\t#%d Reported deployment to %s at %s (%s)",
				int64(deployment["deployment_id"].(float64)),
				deployment["environment"].(string),
				createdAtTimestamp,
				stateString)
			rows = append(rows, deploymentRow)
		}
	} else {
		rows = append(rows, "Deployments:\tNone")
	}

	rows = append(rows, "Evidence:")
	for _, evidenceName := range artifact["template"].([]interface{}) {
		if evidenceName != "artifact" {
			if v, ok := evidenceMap[evidenceName.(string)]; !ok {
				rows = append(rows, fmt.Sprintf("\t%s:\tMISSING", evidenceName))
			} else {
				evidenceData := v.(map[string]interface{})
				isCompliant := "COMPLIANT"
				if !evidenceData["is_compliant"].(bool) {
					isCompliant = "INCOMPLIANT"
				}
				rows = append(rows, fmt.Sprintf("\t%s:\t%s", evidenceName, isCompliant))
			}
		}
	}

	printTable(out, []string{}, rows)
	return nil
}
