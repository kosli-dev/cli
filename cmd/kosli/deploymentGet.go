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

const deploymentGetDesc = `Get a deployment from a specified pipeline`

type deploymentGetOptions struct {
	json         bool
	pipelineName string
}

func newDeploymentGetCmd(out io.Writer) *cobra.Command {
	o := new(deploymentGetOptions)
	cmd := &cobra.Command{
		Use:   "get DEPLOYMENT-ID",
		Short: deploymentGetDesc,
		Long:  deploymentGetDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "deployment ID argument is required")
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

func (o *deploymentGetOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/deployments/%s", global.Host, global.Owner, o.pipelineName, args[0])
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

	var deployment map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &deployment)
	if err != nil {
		return err
	}

	rows := []string{}
	rows = append(rows, fmt.Sprintf("ID:\t%d", int64(deployment["deployment_id"].(float64))))
	rows = append(rows, fmt.Sprintf("Artifact SHA256:\t%s", deployment["artifact_sha256"].(string)))
	rows = append(rows, fmt.Sprintf("Artifact name:\t%s", deployment["artifact_name"].(string)))
	rows = append(rows, fmt.Sprintf("Build URL:\t%s", deployment["build_url"].(string)))
	createdAt, err := formattedTimestamp(deployment["created_at"], false)
	if err != nil {
		return err
	}
	rows = append(rows, fmt.Sprintf("Created at:\t%s", createdAt))
	rows = append(rows, fmt.Sprintf("Environment:\t%s", deployment["environment"].(string)))

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
		stateString = fmt.Sprintf("The artifact running since %s", stateTimestamp)
	} else if state == "exited" {
		stateString = fmt.Sprintf("The artifact exited on %s", stateTimestamp)
	}

	deploymentRow := fmt.Sprintf("Runtime state:\t%s",
		stateString)
	rows = append(rows, deploymentRow)

	printTable(out, []string{}, rows)
	return nil
}
