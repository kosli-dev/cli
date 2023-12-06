package main

import (
	"github.com/kosli-dev/cli/internal/jira"
	"github.com/kosli-dev/cli/internal/requests"
)

type TypedEvidencePayload struct {
	ArtifactFingerprint string      `json:"artifact_fingerprint,omitempty"`
	CommitSHA           string      `json:"commit_sha,omitempty"`
	EvidenceName        string      `json:"name"`
	EvidenceURL         string      `json:"evidence_url,omitempty"`
	EvidenceFingerprint string      `json:"evidence_fingerprint,omitempty"`
	BuildUrl            string      `json:"build_url"`
	UserData            interface{} `json:"user_data,omitempty"`
	Flows               []string    `json:"flows,omitempty"`
}

type GenericEvidencePayload struct {
	TypedEvidencePayload
	Description string `json:"description,omitempty"`
	Compliant   bool   `json:"is_compliant"`
}

type JiraEvidencePayload struct {
	TypedEvidencePayload
	JiraResults []*jira.JiraIssueInfo `json:"jira_results"`
}

type WorkflowEvidencePayload struct {
	ExternalId          string      `json:"external_id"`
	Step                string      `json:"step"`
	EvidenceURL         string      `json:"evidence_url,omitempty"`
	EvidenceFingerprint string      `json:"evidence_fingerprint,omitempty"`
	UserData            interface{} `json:"user_data,omitempty"`
}

// newEvidenceForm constructs a list of FormItems for an evidence
// form submission.
func newEvidenceForm(payload interface{}, evidencePaths []string) (
	[]requests.FormItem, bool, string, error,
) {
	form := []requests.FormItem{
		{Type: "field", FieldName: "evidence_json", Content: payload},
	}

	var evidencePath string
	var cleanupNeeded bool
	var err error

	if len(evidencePaths) > 0 {
		evidencePath, cleanupNeeded, err = getPathOfEvidenceFileToUpload(evidencePaths)
		if err != nil {
			return form, cleanupNeeded, evidencePath, err
		}
		form = append(form, requests.FormItem{Type: "file", FieldName: "evidence_file", Content: evidencePath})
		logger.Debug("evidence file %s will be uploaded", evidencePath)
	}

	return form, cleanupNeeded, evidencePath, nil
}

// newAttestationForm constructs a list of FormItems for an evidence
// form submission.
func newAttestationForm(payload interface{}, evidencePaths []string) (
	[]requests.FormItem, bool, string, error,
) {
	form := []requests.FormItem{
		{Type: "field", FieldName: "data_json", Content: payload},
	}

	var evidencePath string
	var cleanupNeeded bool
	var err error

	if len(evidencePaths) > 0 {
		evidencePath, cleanupNeeded, err = getPathOfEvidenceFileToUpload(evidencePaths)
		if err != nil {
			return form, cleanupNeeded, evidencePath, err
		}
		form = append(form, requests.FormItem{Type: "file", FieldName: "evidence_file", Content: evidencePath})
		logger.Debug("evidence file %s will be uploaded", evidencePath)
	}

	return form, cleanupNeeded, evidencePath, nil
}
