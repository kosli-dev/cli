package types

type PREvidence struct {
	MergeCommit string        `json:"merge_commit"`
	URL         string        `json:"url"`
	State       string        `json:"state"`
	Approvers   []interface{} `json:"approvers"`
	Author      string        `json:"author,omitempty"`
	CreatedAt   int64         `json:"created_at,omitempty"`
	MergedAt    int64         `json:"merged_at,omitempty"`
	Title       string        `json:"title,omitempty"`
	HeadRef     string        `json:"head_ref,omitempty"`
	Commits     []Commit      `json:"commits,omitempty"`
}

type PRApprovals struct {
	Username  string `json:"username"`
	State     string `json:"state"`
	Timestamp int64  `json:"timestamp"`
}

type Commit struct {
	SHA               string `json:"sha1"`
	Message           string `json:"message"`
	Committer         string `json:"author"`
	CommitterUsername string `json:"author_username"`
	Timestamp         int64  `json:"timestamp"`
	Branch            string `json:"branch"`
	URL               string `json:"url,omitempty"`
}

type PRRetriever interface {
	PREvidenceForCommitV2(string) ([]*PREvidence, error)
	PREvidenceForCommitV1(string) ([]*PREvidence, error)
}
