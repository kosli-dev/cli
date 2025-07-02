package types

type PREvidence struct {
	MergeCommit string        `json:"merge_commit"`
	URL         string        `json:"url"`
	State       string        `json:"state"`
	Approvers   []string      `json:"approvers"`
	Approvers2  []PRApprovals `json:"approvers2"`
	Author      string        `json:"author"`
	CreatedAt   int64         `json:"created_at"`
	MergedAt    int64         `json:"merged_at"`
	Title       string        `json:"title"`
	HeadRef     string        `json:"head_ref"`
	Commits     []Commit      `json:"commits"`
	// LastCommit             string `json:"lastCommit"`
	// LastCommitter          string `json:"lastCommitter"`
	// SelfApproved           bool   `json:"selfApproved"`
}

type PRApprovals struct {
	Author    string `json:"author"`
	State     string `json:"state"`
	Timestamp int64  `json:"timestamp"`
}

type Commit struct {
	SHA       string `json:"sha1"`
	Message   string `json:"message"`
	Committer string `json:"author"`
	Timestamp int64  `json:"timestamp"`
}

type PRRetriever interface {
	PREvidenceForCommit(string) ([]*PREvidence, error)
}
