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

type SnapshotType struct {
	Type string `json:"type"`
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
		if o.long {
			pj, err := prettyJson(response.Body)
			if err != nil {
				return err
			}
			fmt.Println(pj)
			return nil
		}
		return showJson(response)
	}

	var snapshotType SnapshotType
	err = json.Unmarshal([]byte(response.Body), &snapshotType)
	if err != nil {
		return err
	}

	if snapshotType.Type == "K8S" || snapshotType.Type == "ECS" {
		return showK8sEcsList(response, o)
	} else if snapshotType.Type == "server" {
		return showServerList(response, o)
	}
	return nil
}

func showJson(response *requests.HTTPResponse) error {
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
		artifactJsonOut.Commit = "xxx"
		artifactJsonOut.CommitUrl = "zzz"
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

func showK8sEcsList(response *requests.HTTPResponse, o *environmentLsOptions) error {
	var snapshot Snapshot
	err := json.Unmarshal([]byte(response.Body), &snapshot)
	if err != nil {
		return err
	}

	formatStringHead := "%-7s  %-40s  %-10s  %-17s  %-25s  %-10s\n"
	formatStringLine := "%-7s  %-40s  %-10s  %-17s  %-25s  %-10d\n"
	fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "TAG", "SHA256", "SINCE", "REPLICAS")

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
		if len(artifactNameSplit) > 1 {
			artifactTag = artifactNameSplit[1]
			if len(artifactTag) > 10 && !o.long {
				artifactTag = artifactTag[:10]
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
		fmt.Printf(formatStringLine, "xxxx", artifactName, artifactTag, shortSha, since, len(artifact.CreationTimestamp))
	}

	return nil
}

func showServerList(response *requests.HTTPResponse, o *environmentLsOptions) error {
	var snapshot Snapshot
	err := json.Unmarshal([]byte(response.Body), &snapshot)
	if err != nil {
		return err
	}

	formatStringHead := "%-7s  %-40s  %-17s  %-25s  %-10s\n"
	formatStringLine := "%-7s  %-40s  %-17s  %-25s  %-10d\n"
	fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "SHA256", "SINCE", "REPLICAS")

	for _, artifact := range snapshot.Artifacts {
		if artifact.Annotation.Now == 0 {
			continue
		}
		since := time.Unix(artifact.CreationTimestamp[0], 0).Format(time.RFC3339)
		artifactName := artifact.Name
		if len(artifactName) > 40 && !o.long {
			artifactName = artifactName[:18] + "..." + artifactName[len(artifactName)-19:]
		}
		shortSha := ""
		if len(artifact.Sha256) == 64 {
			if o.long {
				shortSha = artifact.Sha256
			} else {
				shortSha = artifact.Sha256[:7] + "..." + artifact.Sha256[64-7:]
			}
		}
		fmt.Printf(formatStringLine, "xxxx", artifactName, shortSha, since, artifact.Annotation.Now)
	}

	return nil
}
