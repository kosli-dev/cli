package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const getArtifactShortDesc = `Get artifact from a specified flow`

const getArtifactLongDesc = getArtifactShortDesc + `
You can get an artifact by its fingerprint or by its git commit sha.
In case of using the git commit, it is possible to get multiple artifacts matching the git commit.

The expected argument is an expression to specify the artifact to get.
It has the format <FLOW_NAME><SEPARATOR><COMMIT_SHA1|ARTIFACT_FINGERPRINT> 

Specify SNAPPISH by:
	flowName@<fingerprint>  artifact with a given fingerprint. The fingerprint can be short or complete.
	flowName:<commit_sha>   artifact with a given commit SHA. The commit sha can be short or complete.

Examples of valid expressions are: flow@184c799cd551dd1d8d5c5f9a5d593b2e931f5e36122ee5c793c1d08a19839cc0, flow:110d048bf1fce72ba546cbafc4427fb21b958dee
`

const getArtifactExample = `
# get an artifact with a given fingerprint from a flow
kosli get artifact flowName@fingerprint \
	--api-token yourAPIToken \
	--org orgName

# get an artifact with a given commit SHA from a flow
kosli get artifact flowName:commitSHA \
	--api-token yourAPIToken \
	--org orgName`

type getArtifactOptions struct {
	output string
}

func newGetArtifactCmd(out io.Writer) *cobra.Command {
	o := new(getArtifactOptions)
	cmd := &cobra.Command{
		Use:     "artifact SNAPPISH",
		Short:   getArtifactShortDesc,
		Long:    getArtifactLongDesc,
		Example: getArtifactExample,
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

func (o *getArtifactOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/artifact/?snappish=%s", global.Host, global.Org, url.QueryEscape(args[0]))
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
			"table": printArtifactAsTableWrapper,
			"json":  output.PrintJson,
		})
}

func printArtifactAsTableWrapper(artifactRaw string, out io.Writer, pageNumber int) error {
	// TODO: we have this function for backward compatibility with API.
	// API returns array when querying with commit and returns single map for sha256.
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
	return printArtifactsJsonAsTable(artifacts, out, pageNumber)
}

func printArtifactsJsonAsTable(artifacts []map[string]interface{}, out io.Writer, pageNumber int) error {
	separator := ""
	for _, artifact := range artifacts {
		evidenceMap := artifact["evidence"].(map[string]interface{})
		artifactData := evidenceMap["artifact"].(map[string]interface{})

		rows := []string{}
		rows = append(rows, fmt.Sprintf("Name:\t%s", artifactData["filename"].(string)))
		rows = append(rows, fmt.Sprintf("Flow:\t%s", artifact["pipeline_name"].(string)))
		rows = append(rows, fmt.Sprintf("Fingerprint:\t%s", artifactData["sha256"].(string)))
		createdAt, err := formattedTimestamp(artifactData["logged_at"], false)
		if err != nil {
			return err
		}
		rows = append(rows, fmt.Sprintf("Created on:\t%s", createdAt))
		rows = append(rows, fmt.Sprintf("Git commit:\t%s", artifactData["git_commit"].(string)))
		rows = append(rows, fmt.Sprintf("Commit URL:\t%s", artifactData["commit_url"].(string)))
		rows = append(rows, fmt.Sprintf("Build URL:\t%s", artifactData["build_url"].(string)))

		rows = append(rows, fmt.Sprintf("State:\t%s", artifact["state"].(string)))

		runningInEnvs := artifact["running"].([]interface{})
		if len(runningInEnvs) > 0 {
			runningInEnvNames := []string{}
			for _, envDataInterface := range runningInEnvs {
				envData := envDataInterface.(map[string]interface{})
				runningInEnvNames = append(runningInEnvNames,
					fmt.Sprintf("%s#%.0f", envData["environment_name"].(string), envData["snapshot_index"].(float64)))
			}
			sort.Strings(runningInEnvNames)
			rows = append(rows, fmt.Sprintf("Running in environments:\t%s", strings.Join(runningInEnvNames, ", ")))
		}

		exitedInEnvs := artifact["exited"].([]interface{})
		if len(exitedInEnvs) > 0 {
			exitedInEnvNames := []string{}
			for _, envDataInterface := range exitedInEnvs {
				envData := envDataInterface.(map[string]interface{})
				exitedInEnvNames = append(exitedInEnvNames,
					fmt.Sprintf("%s#%.0f", envData["environment_name"].(string), envData["snapshot_index"].(float64)))
			}
			rows = append(rows, fmt.Sprintf("Exited from environments:\t%s", strings.Join(exitedInEnvNames, ", ")))
		}

		history := artifact["history"].([]interface{})
		if len(history) > 0 {
			rows = append(rows, "History:")
			for _, rawHistory := range history {
				event := rawHistory.(map[string]interface{})
				eventString := event["event"]
				eventTimestamp, err := formattedTimestamp(event["timestamp"], true)
				if err != nil {
					return err
				}
				historyRow := fmt.Sprintf("    %s\t%s", eventString, eventTimestamp)
				rows = append(rows, historyRow)
			}
		}

		fmt.Print(separator)
		separator = "\n"
		tabFormattedPrint(out, []string{}, rows)
	}
	return nil
}
