package gitlab

import (
	"fmt"
	"net/http"
	"time"

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

// NewGitlabRetrieverFunc creates a types.PRRetriever from GitLab config
// parameters. It can be replaced in tests to inject a FakeGitlabClient.
var NewGitlabRetrieverFunc = defaultNewGitlabRetriever

func defaultNewGitlabRetriever(token, baseURL, org, repository string) types.PRRetriever {
	return &GitlabConfig{
		Token:      token,
		BaseURL:    baseURL,
		Org:        org,
		Repository: repository,
	}
}

// ResetGitlabRetrieverFunc restores NewGitlabRetrieverFunc to its default.
func ResetGitlabRetrieverFunc() {
	NewGitlabRetrieverFunc = defaultNewGitlabRetriever
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
		Author:      gitlabUserName(mr.Author),
		CreatedAt:   unixOrZero(mr.CreatedAt),
		MergedAt:    unixOrZero(mr.MergedAt),
		Title:       mr.Title,
	}
	evidence.HeadRef, evidence.BaseRef = gitlabMRRefs(mr)
	approvers, reviews, err := c.GetMergeRequestReviews(mr.IID)
	if err != nil {
		return evidence, err
	}
	evidence.Approvers = approvers
	evidence.Reviews = reviews
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

	mrs, err = c.listMergeRequestsForCommit(client, commit, defaultMaxPages)
	if err != nil {
		return mrs, fmt.Errorf("failed to list merge requests for commit %s: %w", commit, err)
	}
	return mrs, noNullItems("merge requests", "commit "+commit, mrs)
}

// GetMergeRequestApprovers returns a list of users (name and username) who approved an MR
func (c *GitlabConfig) GetMergeRequestApprovers(mrIID, version int64) ([]any, error) {
	var approvers []any
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return approvers, err
	}
	approvedBy, err := c.approvedBy(client, mrIID)
	if err != nil {
		return approvers, err
	}
	return gitlabApprovers(approvedBy, version), nil
}

func (c *GitlabConfig) approvedBy(client *gitlab.Client, mrIID int64) ([]*gitlab.MergeRequestApproverUser, error) {
	approvals, _, err := client.MergeRequestApprovals.GetConfiguration(c.ProjectID(), mrIID)
	if err != nil || approvals == nil {
		return nil, err
	}
	return approvals.ApprovedBy, noNullItems("approved_by", fmt.Sprintf("merge request !%d", mrIID), approvals.ApprovedBy)
}

// GetMergeRequestReviews returns an MR's approvers and its reviews: each
// reviewer's current decision, since GitLab keeps no history of them.
func (c *GitlabConfig) GetMergeRequestReviews(mrIID int64) ([]any, []types.PRApprovals, error) {
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return nil, nil, err
	}
	approvedBy, err := c.approvedBy(client, mrIID)
	if err != nil {
		return nil, nil, err
	}
	approvers := gitlabApprovers(approvedBy, 2)
	reviewers, resp, err := client.MergeRequests.GetMergeRequestReviewers(c.ProjectID(), mrIID)
	if err != nil {
		// Older GitLab has no reviewers endpoint. Reviews without its change
		// requests would read as none given, so they are not reported at all.
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return approvers, nil, nil
		}
		return nil, nil, err
	}
	if err := noNullItems("reviewers", fmt.Sprintf("merge request !%d", mrIID), reviewers); err != nil {
		return nil, nil, err
	}
	return approvers, gitlabReviews(approvedBy, reviewers), nil
}

// gitlabUserName is empty when GitLab names no account (a deleted user): such
// a review is never an approver, and is in Reviews only if it requests changes.
func gitlabUserName(u *gitlab.BasicUser) string {
	if u == nil || u.Username == "" {
		return ""
	}
	return fmt.Sprintf("%s (@%s)", u.Name, u.Username)
}

// gitlabUsername matches people across approvals and reviews; empty when
// GitLab names no account.
func gitlabUsername(u *gitlab.BasicUser) string {
	if u == nil {
		return ""
	}
	return u.Username
}

// unixOrZero is t in Unix seconds, or 0 when GitLab sent no time.
func unixOrZero(t *time.Time) int64 {
	if t == nil {
		return 0
	}
	return t.Unix()
}

// noNullItems errors on a null entry in a list GitLab sent for owner: it is
// missing data, not an empty entry.
func noNullItems[T any](list, owner string, items []*T) error {
	for i, item := range items {
		if item == nil {
			return fmt.Errorf("GitLab sent %s[%d] for %s as null", list, i, owner)
		}
	}
	return nil
}

func gitlabApprovers(approvedBy []*gitlab.MergeRequestApproverUser, version int64) []any {
	var approvers []any
	for _, approver := range approvedBy {
		approverName := gitlabUserName(approver.User)
		if approverName == "" {
			continue
		}
		if version == 1 {
			approvers = append(approvers, approverName)
		} else {
			approvers = append(approvers, types.PRApprovals{
				Username: approverName,
			})
		}
	}
	return approvers
}

// gitlabReviews takes approvals from approved_by only, as people who are not
// reviewers can approve too. A change request outweighs an approval by the
// same person, as there are no review times to order them by.
func gitlabReviews(approvedBy []*gitlab.MergeRequestApproverUser, reviewers []*gitlab.MergeRequestReviewer) []types.PRApprovals {
	// Users with no account all share the key "". That's safe only because
	// add keeps no nameless entry other than a change request.
	reviewerStates := map[string]string{}
	for _, r := range reviewers {
		reviewerStates[gitlabUsername(r.User)] = r.State
	}
	reviews := []types.PRApprovals{}
	add := func(u *gitlab.BasicUser, state string) {
		name := gitlabUserName(u)
		// A change request is kept without a name: dropping it would read as
		// none given.
		if name == "" && state != types.ReviewChangesRequested {
			return
		}
		reviews = append(reviews, types.PRApprovals{Username: name, State: state})
	}
	approved := map[string]bool{}
	for _, a := range approvedBy {
		approved[gitlabUsername(a.User)] = true
		if reviewerStates[gitlabUsername(a.User)] != "requested_changes" {
			add(a.User, types.ReviewApproved)
		}
	}
	for _, r := range reviewers {
		switch {
		case r.State == "requested_changes":
			add(r.User, types.ReviewChangesRequested)
		case r.State == "reviewed" && !approved[gitlabUsername(r.User)]:
			add(r.User, types.ReviewCommented)
		}
	}
	return reviews
}

// GetMergeRequestCommits returns a list of commits for a given MR
func (c *GitlabConfig) GetMergeRequestCommits(mr *gitlab.BasicMergeRequest) ([]types.Commit, error) {
	commits := []types.Commit{}
	client, err := c.NewGitlabClientFromToken()
	if err != nil {
		return commits, err
	}
	glCommits, err := c.listMergeRequestCommits(client, mr.IID, defaultMaxPages)
	if err != nil {
		return commits, err
	}
	owner := fmt.Sprintf("merge request !%d", mr.IID)
	if err := noNullItems("commits", owner, glCommits); err != nil {
		return commits, err
	}
	for i, commit := range glCommits {
		if commit.ID == "" {
			return commits, fmt.Errorf("GitLab sent commits[%d] for %s without an id", i, owner)
		}
	}
	// One signature lookup per commit: pagination multiplies this, but GitLab
	// has no batch signature endpoint.
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
	timestamp := unixOrZero(commit.CreatedAt)
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
// commits; that and a null reply are unsigned (nil fields), not a fatal error
// (server#5892). Other errors propagate so incomplete signature data is never
// silently recorded.
func resolveGitlabSignature(sig *gitlab.GPGSignature, resp *gitlab.Response, err error) (*bool, *string, error) {
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, nil, nil
		}
		return nil, nil, err
	}
	if sig == nil {
		return nil, nil, nil
	}
	verified, signatureState := gitlabCommitVerification(sig.VerificationStatus)
	return verified, signatureState, nil
}

// gitlabMRRefs returns the head (source) and base (target) branch names of a
// merge request.
func gitlabMRRefs(mr *gitlab.BasicMergeRequest) (head, base string) {
	return mr.SourceBranch, mr.TargetBranch
}
