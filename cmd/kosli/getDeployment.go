package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getDeploymentShortDesc = `Get a deployment from a specified flow.`

const getDeploymentLongDesc = getDeploymentShortDesc + `
EXPRESSION can be specified as follows:
- flowName
    - the latest deployment to flowName, at the time of the request
    - e.g., **dashboard**
- flowName#N
    - the Nth deployment, counting from 1
    - e.g., **dashboard#453**
- flowName~N
    - the Nth deployment behind the latest, at the time of the request
    - e.g., **dashboard~56**
`

const getDeploymentExample = `
# get previous deployment in a flow
kosli get deployment flowName~1 \
	--api-token yourAPIToken \
	--org orgName

# get the 10th deployment in a flow
kosli get deployment flowName#10 \
	--api-token yourAPIToken \
	--org orgName

# get the latest deployment in a flow
kosli get deployment flowName \
	--api-token yourAPIToken \
	--org orgName`

type getDeploymentOptions struct {
	output string
}

func newGetDeploymentCmd(out io.Writer) *cobra.Command {
	o := new(getDeploymentOptions)
	cmd := &cobra.Command{
		Use:     "deployment EXPRESSION",
		Short:   getDeploymentShortDesc,
		Long:    getDeploymentLongDesc,
		Example: getDeploymentExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
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

func (o *getDeploymentOptions) run(out io.Writer, args []string) error {
	flowName, id, err := handleExpressions(args[0])
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/v2/deployments/%s/%s/%d", global.Host, global.Org, flowName, id)

	reqParams := &requests.RequestParams{
		Method:   http.MethodGet,
		URL:      url,
		Password: global.ApiToken,
	}
	response, err := kosliClient.Do(reqParams)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printDeploymentAsTable,
			"json":  output.PrintJson,
		})
}

func printDeploymentAsTable(raw string, out io.Writer, page int) error {
	var deployment map[string]interface{}
	err := json.Unmarshal([]byte(raw), &deployment)
	if err != nil {
		return err
	}

	rows := []string{}
	rows = append(rows, fmt.Sprintf("ID:\t%d", int64(deployment["deployment_id"].(float64))))
	rows = append(rows, fmt.Sprintf("Artifact fingerprint:\t%s", deployment["artifact_fingerprint"].(string)))
	rows = append(rows, fmt.Sprintf("Artifact name:\t%s", deployment["artifact_name"].(string)))
	buildURL := "N/A"
	if deployment["build_url"] != nil {
		buildURL = deployment["build_url"].(string)
	}
	rows = append(rows, fmt.Sprintf("Build URL:\t%s", buildURL))
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

	stateString := "Unknown"
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

	tabFormattedPrint(out, []string{}, rows)
	return nil
}
