package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/sirupsen/logrus"
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
	PipelineName      string `json:"pipeline_name"`
	Compliant         bool
	Deployments       []int
	Sha256            string
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
	Image        string `json:"image"`
	Sha256       string `json:"sha256"`
	Pipeline     string `json:"pipeline"`
	Replicas     int    `json:"replicas"`
	RunningSince string `json:"running_since"`
}

func getSnapshot(out io.Writer, o *snapshotGetOptions, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/snapshots/%s", global.Host, global.Owner, url.QueryEscape(args[0]))
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, 0,
		map[string]output.FormatOutputFunc{
			"table": printSnapshotAsTable,
			"json":  printSnapshotAsJson,
		})
}

func printSnapshotAsJson(raw string, out io.Writer, page int) error {
	var snapshot Snapshot
	err := json.Unmarshal([]byte(raw), &snapshot)
	if err != nil {
		return err
	}
	// check if the snapshot is empty by checking one of its elements
	if snapshot.Type == "" {
		fmt.Println("{}")
		return nil
	}
	var result []ArtifactJsonOut
	for _, artifact := range snapshot.Artifacts {
		if artifact.Annotation.Now == 0 {
			continue
		}
		var artifactJsonOut ArtifactJsonOut
		artifactJsonOut.GitCommit = artifact.GitCommit
		artifactJsonOut.CommitUrl = artifact.CommitUrl
		artifactJsonOut.Image = artifact.Name
		artifactJsonOut.Sha256 = artifact.Sha256
		artifactJsonOut.Pipeline = artifact.PipelineName
		artifactJsonOut.Replicas = artifact.Annotation.Now
		sort.Slice(artifact.CreationTimestamp, func(i, j int) bool {
			return artifact.CreationTimestamp[i] < artifact.CreationTimestamp[j]
		})
		oldestTimestamp := artifact.CreationTimestamp[0]
		artifactJsonOut.RunningSince = time.Unix(oldestTimestamp, 0).Format(time.RFC3339)
		result = append(result, artifactJsonOut)
	}

	res, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(res))

	return nil
}

func printSnapshotAsTable(raw string, out io.Writer, page int) error {
	var snapshot Snapshot
	err := json.Unmarshal([]byte(raw), &snapshot)
	if err != nil {
		return err
	}

	// check if the snapshot is empty by checking one of its elements
	if snapshot.Type == "" {
		_, err := out.Write([]byte("No running artifacts were reported\n"))
		if err != nil {
			return err
		}
		return nil

	}

	header := []string{"COMMIT", "ARTIFACT", "PIPELINE", "RUNNING_SINCE", "REPLICAS"}
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

		pipelineName := "N/A"
		if artifact.PipelineName != "" {
			pipelineName = artifact.PipelineName
		}

		row := fmt.Sprintf("%s\tName: %s\t%s\t%s\t%d", gitCommit, artifact.Name, pipelineName, since, len(artifact.CreationTimestamp))
		rows = append(rows, row)
		row = fmt.Sprintf("\tSHA256: %s\t\t\t", artifact.Sha256)
		rows = append(rows, row)
		rows = append(rows, "\t\t\t\t")
	}
	tabFormattedPrint(out, header, rows)
	return nil
}
