package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
	"github.com/xeonx/timeago"
)

type Annotation struct {
	Type string `json:"type"`
	Was  int
	Now  int
}

type Owner struct {
	ApiVersion         string
	Kind               string
	Name               string
	Uid                string
	Controller         bool
	BlockOwnerDeletion bool
}

type PodContent struct {
	Namespace         string
	CreationTimestamp int64
	Owners            []Owner
}

type Artifact struct {
	Name              string
	FlowName          string `json:"flow_name"`
	Compliant         bool
	Deployments       []int
	Fingerprint       string
	GitCommit         string `json:"git_commit"`
	CommitUrl         string `json:"commit_url"`
	CreationTimestamp []int64
	Pods              map[string]PodContent
	Annotation        Annotation
}

type Snapshot struct {
	Index     int
	Timestamp float32
	Type      string `json:"type"`
	UserId    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Artifacts []Artifact
	Compliant bool
}

type ArtifactJsonOut struct {
	GitCommit    string `json:"git_commit"`
	CommitUrl    string `json:"commit_url"`
	Image        string `json:"artifact"`
	Fingerprint  string `json:"fingerprint"`
	Flow         string `json:"flow"`
	Replicas     int    `json:"replicas"`
	RunningSince string `json:"running_since"`
}
type environmentGetOptions struct {
	output string
}

const getSnapshotDescShort = `Get a specified environment snapshot.`

const getSnapshotDesc = getSnapshotDescShort + `
ENVIRONMENT-NAME-OR-EXPRESSION can be specified as follows:
- environmentName
    - the latest snapshot for environmentName, at the time of the request
    - e.g., **prod**
- environmentName#N
    - the Nth snapshot, counting from 1
    - e.g., **prod#42**
- environmentName~N
    - the Nth snapshot behind the latest, at the time of the request
    - e.g., **prod~5**
- environmentName@{YYYY-MM-DDTHH:MM:SS}
    - the snapshot at specific moment in time in UTC
    - e.g., **prod@{2023-10-02T12:00:00}**
- environmentName@{N.<hours|days|weeks|months>.ago}
    - the snapshot at a time relative to the time of the request
    - e.g., **prod@{2.hours.ago}**
`

const getSnapshotExample = `
# get the latest snapshot of an environment:
kosli get snapshot yourEnvironmentName
	--api-token yourAPIToken \
	--org yourOrgName 

# get the SECOND latest snapshot of an environment:
kosli get snapshot yourEnvironmentName~1
	--api-token yourAPIToken \
	--org yourOrgName 

# get the snapshot number 23 of an environment:
kosli get snapshot yourEnvironmentName#23
	--api-token yourAPIToken \
	--org yourOrgName 
	
# get the environment snapshot at midday (UTC), on valentine's day of 2023:
kosli get snapshot yourEnvironmentName@{2023-02-14T12:00:00}
	--api-token yourAPIToken \
	--org yourOrgName

# get the environment snapshot based on a relative time:
kosli get snapshot yourEnvironmentName@{3.weeks.ago}
--api-token yourAPIToken \
--org yourOrgName`

func newGetSnapshotCmd(out io.Writer) *cobra.Command {
	o := new(environmentGetOptions)
	cmd := &cobra.Command{
		Use:     "snapshot ENVIRONMENT-NAME-OR-EXPRESSION",
		Short:   getSnapshotDescShort,
		Long:    getSnapshotDesc,
		Example: getSnapshotExample,
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

func (o *environmentGetOptions) run(out io.Writer, args []string) error {
	envName, id, err := handleSnapshotExpressions(args[0])
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/api/v2/snapshots/%s/%s/%s", global.Host, global.Org, envName, id)

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
			"table": printSnapshotAsTable,
			"json":  output.PrintJson,
		})
}

func printSnapshotAsTable(raw string, out io.Writer, page int) error {
	var snapshot Snapshot
	err := json.Unmarshal([]byte(raw), &snapshot)
	if err != nil {
		return err
	}

	// check if the snapshot is empty by checking one of its elements
	if snapshot.Type == "" {
		logger.Info("No running artifacts were reported")
		return nil
	}

	header := []string{"COMMIT", "ARTIFACT", "FLOW", "RUNNING_SINCE", "REPLICAS"}
	rows := []string{}
	for _, artifact := range snapshot.Artifacts {
		if artifact.Annotation.Now == 0 {
			continue
		}
		timestamp := time.Unix(artifact.CreationTimestamp[0], 0)
		timeago.English.Max = 36 * timeago.Month
		since := timeago.English.Format(timestamp)

		gitCommit := "N/A"
		if artifact.GitCommit != "" {
			gitCommit = artifact.GitCommit[:7]
		}

		flowName := "N/A"
		if artifact.FlowName != "" {
			flowName = artifact.FlowName
		}

		row := fmt.Sprintf("%s\tName: %s\t%s\t%s\t%d", gitCommit, artifact.Name, flowName, since, len(artifact.CreationTimestamp))
		rows = append(rows, row)
		row = fmt.Sprintf("\tFingerprint: %s\t\t\t", artifact.Fingerprint)
		rows = append(rows, row)
		rows = append(rows, "\t\t\t\t")
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
