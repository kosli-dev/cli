package types

type PREvidence struct {
	MergeCommit string   `json:"merge_commit"`
	URL         string   `json:"url"`
	State       string   `json:"state"`
	Approvers   []any    `json:"approvers"`
	Author      string   `json:"author,omitempty"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	MergedAt    int64    `json:"merged_at,omitempty"`
	Title       string   `json:"title,omitempty"`
	HeadRef     string   `json:"head_ref,omitempty"`
	Commits     []Commit `json:"commits,omitempty"`
}

type PRApprovals struct {
	Username  string `json:"username"`
	State     string `json:"state,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type Commit struct {
	SHA               string `json:"sha1"`
	Message           string `json:"message"`
	Committer         string `json:"author"`
	CommitterUsername string `json:"author_username,omitempty"`
	Timestamp         int64  `json:"timestamp"`
	Branch            string `json:"branch"`
	URL               string `json:"url,omitempty"`
}

type PRRetriever interface {
	PREvidenceForCommitV2(string) ([]*PREvidence, error)
	PREvidenceForCommitV1(string) ([]*PREvidence, error)
	// PREvidenceForCommitHybrid tries V2 (GraphQL by commit SHA) first. If it
	// returns no results it falls back to V1 REST discovery + per-PR GraphQL so
	// that GitHub's eventual consistency on associatedPullRequests never causes
	// a false "no PR found".
	PREvidenceForCommitHybrid(string) ([]*PREvidence, error)
	// ProviderAndLabel returns the provider name (e.g. "github") and the label
	// for a pull request (e.g. "pull request", or "merge request" for GitLab).
	ProviderAndLabel() (string, string)
}
