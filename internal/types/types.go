package types

type PREvidence struct {
	MergeCommit string   `json:"merge_commit"`
	URL         string   `json:"url"`
	State       string   `json:"state"`
	Approvers   []string `json:"approvers"`
	// LastCommit             string `json:"lastCommit"`
	// LastCommitter          string `json:"lastCommitter"`
	// SelfApproved           bool   `json:"selfApproved"`
}

type PRRetriever interface {
	PREvidenceForCommit(string) ([]*PREvidence, error)
}
