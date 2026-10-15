package types

import "encoding/json"

type PREvidence struct {
	MergeCommit string   `json:"merge_commit"`
	URL         string   `json:"url"`
	State       string   `json:"state"`
	Approvers   []any    `json:"approvers"`
	Author      string   `json:"author"`
	CreatedAt   int64    `json:"created_at,omitempty"`
	MergedAt    int64    `json:"merged_at,omitempty"`
	Title       string   `json:"title,omitempty"`
	HeadRef     string   `json:"head_ref,omitempty"`
	HeadSHA     string   `json:"head_sha,omitempty"`
	BaseRef     string   `json:"base_ref,omitempty"`
	Commits     []Commit `json:"commits"`
	// Reviews holds every review, not only approvals, so a policy can see a
	// change request given after an approval.
	Reviews []PRApprovals `json:"reviews,omitempty"`
}

// MarshalJSON keeps "commits" in the payload even when a provider returns no
// commits. The API requires the field on FoundPullRequestV2, so omitting it —
// or sending null for a nil slice — produces a payload matching neither
// attestation schema (#1081).
//
// It also sends an empty reviews list, which omitempty would drop: a missing
// field means reviews were not collected, an empty one that there were none.
func (e PREvidence) MarshalJSON() ([]byte, error) {
	type prEvidence PREvidence // sheds the method set, so this does not recurse
	out := struct {
		prEvidence
		Reviews *[]PRApprovals `json:"reviews,omitempty"` // shadows the embedded field
	}{prEvidence: prEvidence(e)}
	if out.Commits == nil {
		out.Commits = []Commit{}
	}
	if e.Reviews != nil {
		out.Reviews = &e.Reviews
	}
	return json.Marshal(out)
}

type PRApprovals struct {
	Username  string `json:"username"`
	State     string `json:"state,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
	CommitSHA string `json:"commit_sha,omitempty"`
}

// Review states, whatever the git provider calls them.
const (
	ReviewApproved         = "APPROVED"
	ReviewChangesRequested = "CHANGES_REQUESTED"
	ReviewCommented        = "COMMENTED"
	// ReviewDismissed is a review GitHub withdrew; it no longer says what it decided.
	ReviewDismissed = "DISMISSED"
)

// ApprovalsFrom lists the approvals among reviews, one per reviewer per commit
// with the latest kept: approving the same commit twice adds nothing, while
// approvals of two commits are two facts a policy may need.
func ApprovalsFrom(reviews []PRApprovals) []any {
	approvals := []any{}
	index := map[[2]string]int{}
	for _, r := range reviews {
		if r.State != ReviewApproved {
			continue
		}
		key := [2]string{r.Username, r.CommitSHA}
		if i, seen := index[key]; seen {
			approvals[i] = r
			continue
		}
		index[key] = len(approvals)
		approvals = append(approvals, r)
	}
	return approvals
}

type Commit struct {
	SHA            string  `json:"sha1"`
	Message        string  `json:"message"`
	Author         string  `json:"author"`
	AuthorUsername string  `json:"author_username,omitempty"`
	Timestamp      int64   `json:"timestamp"`
	Branch         string  `json:"branch"`
	URL            string  `json:"url,omitempty"`
	Verified       *bool   `json:"verified,omitempty"`
	SignatureState *string `json:"signature_state,omitempty"`
	// SignerUsername is the account whose key made the signature, which the
	// author fields cannot show: git lets anyone set those.
	SignerUsername string `json:"signer_username,omitempty"`
	// SignedByPlatform is true when the platform signed the commit, not a
	// person's key; nil when the commit is unsigned or the provider does not say.
	SignedByPlatform *bool `json:"signed_by_platform,omitempty"`
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
