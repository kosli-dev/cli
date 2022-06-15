package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/sirupsen/logrus"
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
	Pipeline_name     string
	Compliant         bool
	Deployments       []int
	Sha256            string
	Git_commit        string
	Commit_url        string
	CreationTimestamp []int64
	Pods              map[string]PodContent
	Annotation        Annotation
}

type Snapshot struct {
	Index     int
	Timestamp float32
	Type      string `json:"type"`
	User_id   string
	User_name string
	Artifacts []Artifact
	Compliant bool
}

type ArtifactJsonOut struct {
	Commit       string `json:"commit"`
	CommitUrl    string `json:"commit-url"`
	Image        string `json:"image"`
	Sha256       string `json:"sha256"`
	Replicas     int    `json:"replicas"`
	RunningSince string `json:"running-since"`
}

func snapshotLs(out io.Writer, o *environmentLsOptions, args []string) error {
	url := fmt.Sprintf("%s/api/v1/environments/%s/%s/data", global.Host, global.Owner, args[0])
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())

	if err != nil {
		return fmt.Errorf("kosli server %s is unresponsive", global.Host)
	}

	if o.json {
		return showJson(response, o)
	}

	return showList(response, o)
}

func showJson(response *requests.HTTPResponse, o *environmentLsOptions) error {
	if o.long {
		pj, err := prettyJson(response.Body)
		if err != nil {
			return err
		}
		fmt.Println(pj)
		return nil
	}

	var snapshot Snapshot
	err := json.Unmarshal([]byte(response.Body), &snapshot)
	if err != nil {
		return err
	}
	var result []ArtifactJsonOut
	for _, artifact := range snapshot.Artifacts {
		if artifact.Annotation.Now == 0 {
			continue
		}
		var artifactJsonOut ArtifactJsonOut
		artifactJsonOut.Commit = artifact.Git_commit
		artifactJsonOut.CommitUrl = artifact.Commit_url
		artifactJsonOut.Image = artifact.Name
		artifactJsonOut.Sha256 = artifact.Sha256
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

func showList(response *requests.HTTPResponse, o *environmentLsOptions) error {
	var snapshot Snapshot
	err := json.Unmarshal([]byte(response.Body), &snapshot)
	if err != nil {
		return err
	}

	hasType, formatStringLine := getFormatStrings(&snapshot)
	for _, artifact := range snapshot.Artifacts {
		if artifact.Annotation.Now == 0 {
			continue
		}
		since := time.Unix(artifact.CreationTimestamp[0], 0).Format(time.RFC3339)
		artifactNameSplit := strings.Split(artifact.Name, ":")
		artifactName := artifactNameSplit[0]
		if len(artifactName) > 40 && !o.long {
			artifactName = artifactName[:18] + "..." + artifactName[len(artifactName)-19:]
		}
		artifactTag := ""
		if hasType {
			if len(artifactNameSplit) > 1 {
				artifactTag = artifactNameSplit[1]
				if len(artifactTag) > 10 && !o.long {
					artifactTag = artifactTag[:10]
				}
			}
		}
		shortSha := ""
		if len(artifact.Sha256) == 64 {
			if o.long {
				shortSha = artifact.Sha256
			} else {
				shortSha = artifact.Sha256[:7] + "..." + artifact.Sha256[64-7:]
			}
		}
		gitCommit := "N/A"
		if artifact.Git_commit != "" {
			if o.long {
				gitCommit = artifact.Git_commit
			} else {
				gitCommit = artifact.Git_commit[:7]
			}
		}

		fmt.Printf(formatStringLine, gitCommit, artifactName, artifactTag, shortSha, since, len(artifact.CreationTimestamp))
	}

	return nil
}

func getFormatStrings(snapshot *Snapshot) (bool, string) {
	var hasType bool
	var formatStringHead string
	var formatStringLine string
	if snapshot.Type == "K8S" || snapshot.Type == "ECS" {
		hasType = true
		formatStringHead = "%-7s  %-40s  %-10s  %-17s  %-25s  %-10s\n"
		formatStringLine = "%-7s  %-40s  %-10s  %-17s  %-25s  %-10d\n"
		fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "TAG", "SHA256", "SINCE", "REPLICAS")
	} else if snapshot.Type == "server" {
		hasType = false
		formatStringHead = "%-7s  %-40s  %-17s  %-25s  %-10s\n"
		formatStringLine = "%-7s  %-40s %s  %-17s  %-25s  %-10d\n"
		fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "SHA256", "SINCE", "REPLICAS")
	}
	// TODO: add default handling of unknown snapshot type
	return hasType, formatStringLine
}
