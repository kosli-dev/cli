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

	hasTag, formatStringLine := getFormatStrings(&snapshot, o.long)
	for _, artifact := range snapshot.Artifacts {
		if artifact.Annotation.Now == 0 {
			continue
		}
		since := time.Unix(artifact.CreationTimestamp[0], 0).Format(time.RFC3339)
		artifactName, artifactTag := splitImageName(artifact.Name)
		if len(artifactName) > 40 && !o.long {
			artifactName = artifactName[:18] + "..." + artifactName[len(artifactName)-19:]
		}
		if hasTag {
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

func getFormatStrings(snapshot *Snapshot, longOption bool) (bool, string) {
	var hasTag bool
	var formatStringHead string
	var formatStringLine string
	maxCommitLength := 7
	maxImageLength := 40
	maxTagLength := 10
	maxSha256Length := 17

	if longOption {
		maxImageLength = 0
		maxTagLength = 0
		for _, artifact := range snapshot.Artifacts {
			artifactName, artifactTag := splitImageName(artifact.Name)
			if len(artifactName) > maxImageLength {
				maxImageLength = len(artifactName)
			}
			if len(artifactTag) > maxTagLength {
				maxTagLength = len(artifactTag)
			}
		}
		maxCommitLength = 40
		maxSha256Length = 64
	}

	if snapshot.Type == "K8S" || snapshot.Type == "ECS" {
		hasTag = true
		fmt.Println(maxImageLength)
		formatStringHead = fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-25s  %%-10s\n", maxCommitLength, maxImageLength, maxTagLength, maxSha256Length)
		formatStringLine = fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-%ds  %%-25s  %%-10d\n", maxCommitLength, maxImageLength, maxTagLength, maxSha256Length)
		fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "TAG", "SHA256", "SINCE", "REPLICAS")
	} else if snapshot.Type == "server" {
		hasTag = false
		formatStringHead = fmt.Sprintf("%%-%ds  %%-%ds  %%-%ds  %%-25s  %%-10s\n", maxCommitLength, maxImageLength, maxSha256Length)
		formatStringLine = fmt.Sprintf("%%-%ds  %%-%ds %%s %%-%ds  %%-25s  %%-10d\n", maxCommitLength, maxImageLength, maxSha256Length)
		fmt.Printf(formatStringHead, "COMMIT", "IMAGE", "SHA256", "SINCE", "REPLICAS")
	}
	// TODO: add default handling of unknown snapshot type
	return hasTag, formatStringLine
}

func splitImageName(imageName string) (string, string) {
	// TODO: properly parse the image name to get tag
	// https://github.com/cyber-dojo/runner/blob/e98bc280c5349cb2919acecb0dfbfefa1ac4e5c3/src/docker/image_name.rb
	artifactNameSplit := strings.Split(imageName, ":")
	artifactName := artifactNameSplit[0]
	artifactTag := ""
	if len(artifactNameSplit) > 1 {
		artifactTag = artifactNameSplit[1]
	}
	return artifactName, artifactTag
}
