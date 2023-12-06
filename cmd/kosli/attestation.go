package main

import (
	"fmt"
	"strings"

	"github.com/kosli-dev/cli/internal/gitview"
	"github.com/kosli-dev/cli/internal/requests"
)

type CommonAttestationPayload struct {
	ArtifactFingerprint string                   `json:"artifact_fingerprint,omitempty"`
	Commit              *gitview.BasicCommitInfo `json:"git_commit_info,omitempty"`
	AttestationName     string                   `json:"step_name"`
	TargetArtifacts     []string                 `json:"target_artifacts,omitempty"`
	EvidenceURL         string                   `json:"evidence_url,omitempty"`
	EvidenceFingerprint string                   `json:"evidence_fingerprint,omitempty"`
	Url                 string                   `json:"url,omitempty"`
	UserData            interface{}              `json:"user_data,omitempty"`
}

type CommonAttestationOptions struct {
	fingerprintOptions      *fingerprintOptions
	attestationNameTemplate string
	flowName                string
	trailName               string
	userDataFilePath        string
	evidencePaths           []string
	commitSHA               string
	srcRepoRoot             string
}

func (o *CommonAttestationOptions) run(args []string, payload *CommonAttestationPayload) error {
	var err error

	p1, p2, err := parseAttestationNameTemplate(o.attestationNameTemplate)
	if err != nil {
		return err
	}
	if p1 != "" && p2 != "" {
		payload.TargetArtifacts = []string{p1}
		payload.AttestationName = p2
	} else {
		payload.AttestationName = p1
	}

	if o.fingerprintOptions.artifactType != "" {
		payload.ArtifactFingerprint, err = GetSha256Digest(args[0], o.fingerprintOptions, logger)
		if err != nil {
			return err
		}
	}

	if o.commitSHA != "" {
		gv, err := gitview.New(o.srcRepoRoot)
		if err != nil {
			return err
		}
		commitInfo, err := gv.GetCommitInfoFromCommitSHA(o.commitSHA)
		if err != nil {
			return err
		}
		payload.Commit = &commitInfo.BasicCommitInfo
	}

	payload.UserData, err = LoadJsonData(o.userDataFilePath)
	return err
}

func prepareAttestationForm(payload interface{}, evidencePaths []string) ([]requests.FormItem, bool, string, error) {
	form, cleanupNeeded, evidencePath, err := newAttestationForm(payload, evidencePaths)
	if err != nil {
		return []requests.FormItem{}, cleanupNeeded, evidencePath, err
	}
	return form, cleanupNeeded, evidencePath, nil
}

func parseAttestationNameTemplate(template string) (string, string, error) {
	parts := strings.Split(template, ".")
	if len(parts) == 1 {
		return parts[0], "", nil
	} else if len(parts) == 2 {
		return parts[0], parts[1], nil
	} else {
		return "", "", fmt.Errorf("invalid attestation name format")
	}
}
