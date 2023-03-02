package main

type TypedEvidencePayload struct {
	ArtifactFingerprint string      `json:"artifact_fingerprint,omitempty"`
	CommitSHA           string      `json:"commit_sha,omitempty"`
	EvidenceName        string      `json:"name"`
	EvidenceURL         string      `json:"evidence_url,omitempty"`
	EvidenceFingerprint string      `json:"evidence_fingerprint,omitempty"`
	BuildUrl            string      `json:"build_url"`
	UserData            interface{} `json:"user_data,omitempty"`
	Flows               []string    `json:"pipelines,omitempty"`
}
