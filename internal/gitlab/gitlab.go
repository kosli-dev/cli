package gitlab

import (
	"fmt"
	"net/http"

	"github.com/kosli-dev/cli/internal/types"
	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabConfig struct {
	Token      string
	BaseURL    string
	Org        string
	Repository string
}

// GetClientOptFns creates a list of ClientOptionFunc
func (c *GitlabConfig) GetClientOptFns() []gitlab.ClientOptionFunc {
	clientOptFns := []gitlab.ClientOptionFunc{}
	if c.BaseURL != "" {
		clientOptFns = append(clientOptFns, gitlab.WithBaseURL(c.BaseURL))
	}
	return clientOptFns
}

// NewGitlabClientFromToken returns an API client from GitlabConfig
func (c *GitlabConfig) NewGitlabClientFromToken() (*gitlab.Client, error) {
	client, err := gitlab.NewClient(c.Token, c.GetClientOptFns()...)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// projectID returns a project ID that can be
// used when making calls to Gitlab API
func (c *GitlabConfig) ProjectID() string {
	return fmt.Sprintf("%s/%s", c.Org, c.Repository)
}

func (c *GitlabConfig) ProviderAndLabel() (string, string) {
	return "gitlab", "merge request"
}

// This is the old implementation, it will be removed after the PR payload is enhanced for all VCS providers
func (c *GitlabConfig) PREvidenceForCommitV1(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	mrs, err := c.MergeRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := c.newPRGitlabEvidenceV1(mr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

func (c *GitlabConfig) PREvidenceForCommitHybrid(commit string) ([]*types.PREvidence, error) {
	return c.PREvidenceForCommitV2(commit)
}

// This is the new implementation, it will be used for all VCS providers
func (c *GitlabConfig) PREvidenceForCommitV2(commit string) ([]*types.PREvidence, error) {
	pullRequestsEvidence := []*types.PREvidence{}
	mrs, err := c.MergeRequestsForCommit(commit)
	if err != nil {
		return pullRequestsEvidence, err
	}
	for _, mr := range mrs {
		evidence, err := c.newPRGitlabEvidenceV2(mr)
		if err != nil {
			return pullRequestsEvidence, err
		}
		pullRequestsEvidence = append(pullRequestsEvidence, evidence)
	}
	return pullRequestsEvidence, nil
}

func (c *GitlabConfig) newPRGitlabEvidenceV1(mr *gitlab.BasicMergeRequest) (*types.PREvidence, error) {
	evidence := &types.PREvidence{
		URL:         mr.WebURL,
		MergeCommit: mr.MergeCommitSHA,
		State:       mr.State,
	}
	approvers, err := c.GetMergeRequestApprovers(mr.IID, 1)
	if err != nil {
		return evidence, err
	}
	if approvers == nil {
		approvers = []any{}
	}
	evidence.Approvers = approvers
	return evidence, nil
}

func (c *GitlabConfig) newPRGitlabEvidenceV2(mr *gitlab.BasicMergeRequest) (*types.PREvidence, error) {
	evidence := &types.PREvidence{
		URL:         mr.WebURL,
		MergeCommit: mr.MergeCommitSHA,
		State:       mr.State,
		Author:      fmt.Sprintf("%s (@%s)", mr.Author.Name, mr.Author.Username),
		CreatedAt:   mr.CreatedAt.Unix(),
		MergedAt:    mr.MergedAt.Unix(),
		Title:       mr.Title,
	}
	evidence.HeadRef, evidence.BaseRef = gitlabMRRefs(mr)
	approvers, err := c.GetMergeRequestApprovers(mr.IID, 2)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	commits, err := c.GetMergeRequestCommits(mr)
	if err != nil {
		return evidence, err
	}
	evidence.Commits = commits
	return evidence, nil
}

// MergeRequestsForCommit returns a list of MRs for a given commit
func (c *GitlabConfig) MergeRequestsForCommit(commit string) ([]*gitlab.BasicMergeRequest, error) {
	mrs := []*gitlab.BasicMergeRequest{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return mrs, err
	}

	mrs, _, err = client.Commits.ListMergeRequestsByCommit(c.ProjectID(), commit)
	if err != nil {
		return mrs, fmt.Errorf("failed to list merge requests for commit %s: %v", commit, err)
	}
	return mrs, nil
}

// GetMergeRequestApprovers returns a list of users (name and username) who approved an MR
func (c *GitlabConfig) GetMergeRequestApprovers(mrIID, version int64) ([]any, error) {
	var approvers []any
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return approvers, err
	}
	approvals, _, err := client.MergeRequestApprovals.GetConfiguration(c.ProjectID(), mrIID)
	if err != nil {
		return approvers, err
	}

	for _, approver := range approvals.ApprovedBy {
		approverName := fmt.Sprintf("%s (@%s)", approver.User.Name, approver.User.Username)
		if version == 1 {
			approvers = append(approvers, approverName)
		} else {
			approvers = append(approvers, types.PRApprovals{
				Username: approverName,
			})
		}
	}
	return approvers, nil
}

// GetMergeRequestCommits returns a list of commits for a given MR
func (c *GitlabConfig) GetMergeRequestCommits(mr *gitlab.BasicMergeRequest) ([]types.Commit, error) {
	commits := []types.Commit{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return commits, err
	}
	glCommits, _, err := client.MergeRequests.GetMergeRequestCommits(c.ProjectID(), mr.IID,
		&gitlab.GetMergeRequestCommitsOptions{})
	if err != nil {
		return commits, err
	}
	for _, commit := range glCommits {
		mappedCommit := commitFromGitlabCommit(commit, mr.SourceBranch)
		verified, signatureState, err := resolveGitlabSignature(
			client.Commits.GetGPGSignature(c.ProjectID(), commit.ID),
		)
		if err != nil {
			return commits, err
		}
		mappedCommit.Verified = verified
		mappedCommit.SignatureState = signatureState
		commits = append(commits, mappedCommit)
	}
	return commits, nil
}

// commitFromGitlabCommit maps a GitLab API commit to a types.Commit.
func commitFromGitlabCommit(commit *gitlab.Commit, branch string) types.Commit {
	// Use the authored date to match the recorded author identity; fall back to
	// created_at if the API omits it (server#5479).
	timestamp := commit.CreatedAt.Unix()
	if commit.AuthoredDate != nil {
		timestamp = commit.AuthoredDate.Unix()
	}
	return types.Commit{
		SHA:       commit.ID,
		Message:   commit.Message,
		Author:    fmt.Sprintf("%s <%s>", commit.AuthorName, commit.AuthorEmail),
		Timestamp: timestamp,
		Branch:    branch,
		URL:       commit.WebURL,
	}
}

// gitlabCommitVerification maps a GitLab signature verification_status to the
// neutral verified/signature_state fields (server#5892). The status is
// "verified" only when the signature is cryptographically valid; an empty
// status leaves both nil (unsigned commits 404 and are handled by the caller).
func gitlabCommitVerification(status string) (*bool, *string) {
	if status == "" {
		return nil, nil
	}
	verified := status == "verified"
	return &verified, &status
}

// resolveGitlabSignature maps the result of a GetGPGSignature call to the
// neutral verified/signature_state fields. GitLab returns 404 for unsigned
// commits, which is treated as unsigned (nil fields) rather than a fatal error
// (server#5892); other errors propagate so incomplete signature data is never
// silently recorded.
func resolveGitlabSignature(sig *gitlab.GPGSignature, resp *gitlab.Response, err error) (*bool, *string, error) {
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	verified, signatureState := gitlabCommitVerification(sig.VerificationStatus)
	return verified, signatureState, nil
}

// gitlabMRRefs returns the head (source) and base (target) branch names of a
// merge request.
func gitlabMRRefs(mr *gitlab.BasicMergeRequest) (head, base string) {
	return mr.SourceBranch, mr.TargetBranch
}
