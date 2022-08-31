package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const artifactGetDesc = `Get artifact from a specified pipeline`

const artifactGetExample = `
# get an artifact with a given SHA256 in a pipeline
kosli artifact get yourPipelineName@yourSHA256 \
	--api-token yourAPIToken \
	--owner yourOrgName
# get an artifact with a given commit SHA256 in a pipeline
kosli artifact get yourPipelineName:yourCommitSHA256 \
	--api-token yourAPIToken \
	--owner yourOrgName
`

type artifactGetOptions struct {
	output string
}

func newArtifactGetCmd(out io.Writer) *cobra.Command {
	o := new(artifactGetOptions)
	cmd := &cobra.Command{
		Use:     "get SNAPPISH",
		Short:   artifactGetDesc,
		Long:    artifactGetDesc,
		Example: artifactGetExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "SNAPPISH argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)

	return cmd
}

func (o *artifactGetOptions) run(out io.Writer, args []string) error {
	kurl := fmt.Sprintf("%s/api/v1/projects/%s/artifact/?snappish=%s", global.Host, global.Owner, url.QueryEscape(args[0]))
	response, err := requests.SendPayload([]byte{}, kurl, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printArtifactAsTableWrapper,
			"json":  output.PrintJson,
		})
}

func printArtifactAsTableWrapper(artifactRaw string, out io.Writer, pageNumber int) error {
	// TODO: we have this function for backward compatibility with API.
	// API returns array when querrying with commmit and returns single map for sha256.
	// In the future, the return json should always be an array
	if artifactRaw[0] != '[' {
		artifactRaw = "[" + artifactRaw + "]"
	}
	return printArtifactsAsTable(artifactRaw, out, pageNumber)
}

func printArtifactsAsTable(artifactRaw string, out io.Writer, pageNumber int) error {
	var artifacts []map[string]interface{}
	err := json.Unmarshal([]byte(artifactRaw), &artifacts)
	if err != nil {
		return err
	}
	for _, artifact := range artifacts {
		approvals := artifact["approvals"].([]interface{})
		deployments := artifact["deployments"].([]interface{})

		evidenceMap := artifact["evidence"].(map[string]interface{})
		artifactData := evidenceMap["artifact"].(map[string]interface{})

		rows := []string{}
		rows = append(rows, fmt.Sprintf("Name:\t%s", artifactData["filename"].(string)))
		rows = append(rows, fmt.Sprintf("SHA256:\t%s", artifactData["sha256"].(string)))
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
			for _, rawApproval := range approvals {
				approval := rawApproval.(map[string]interface{})
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
			for _, rawDeployment := range deployments {
				deployment := rawDeployment.(map[string]interface{})
				deploymentState := deployment["running_state"].(map[string]interface{})
				state := deploymentState["state"].(string)
				stateTimestamp, err := formattedTimestamp(deploymentState["timestamp"], true)
				if err != nil {
					return err
				}

				stateString := "Unknown"
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

		tabFormattedPrint(out, []string{}, rows)
	}
	return nil

}
