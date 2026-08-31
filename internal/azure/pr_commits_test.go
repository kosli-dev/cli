package azure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/stretchr/testify/require"
)

// fakePRCommitsClient returns canned pages and records the args of every call,
// so tests can assert Top and the continuation token were sent.
type fakePRCommitsClient struct {
	pages []*git.GetPullRequestCommitsResponseValue
	err   error
	calls []git.GetPullRequestCommitsArgs
}

func (f *fakePRCommitsClient) GetPullRequestCommits(_ context.Context,
	args git.GetPullRequestCommitsArgs) (*git.GetPullRequestCommitsResponseValue, error) {
	f.calls = append(f.calls, args)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.calls) > len(f.pages) {
		return f.pages[len(f.pages)-1], nil
	}
	return f.pages[len(f.calls)-1], nil
}

func commitRefs(shas ...string) []git.GitCommitRef {
	refs := []git.GitCommitRef{}
	for _, sha := range shas {
		sha, comment, url := sha, "msg", "https://dev.azure.com/o/_git/r/commit/"+sha
		name, email := "Ada", "ada@example.com"
		date := azuredevops.Time{Time: t0}
		refs = append(refs, git.GitCommitRef{
			CommitId: &sha,
			Comment:  &comment,
			Url:      &url,
			Author:   &git.GitUserDate{Name: &name, Email: &email, Date: &date},
		})
	}
	return refs
}

var t0 = time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC)

func page(token string, shas ...string) *git.GetPullRequestCommitsResponseValue {
	return &git.GetPullRequestCommitsResponseValue{Value: commitRefs(shas...), ContinuationToken: token}
}

func shasOf(refs []git.GitCommitRef) []string {
	shas := []string{}
	for _, r := range refs {
		shas = append(shas, *r.CommitId)
	}
	return shas
}

func TestFetchAzurePRCommits_FollowsContinuationToken(t *testing.T) {
	client := &fakePRCommitsClient{pages: []*git.GetPullRequestCommitsResponseValue{
		page("tok1", "sha1", "sha2"),
		page("", "sha3"),
	}}

	refs, err := fetchAzurePRCommits(context.Background(), client, git.GetPullRequestCommitsArgs{}, defaultMaxPages)
	require.NoError(t, err)
	require.Equal(t, []string{"sha1", "sha2", "sha3"}, shasOf(refs))
	require.Len(t, client.calls, 2)
	require.Equal(t, azurePageSize, *client.calls[0].Top)
	require.Nil(t, client.calls[0].ContinuationToken)
	require.Equal(t, "tok1", *client.calls[1].ContinuationToken)
}

func TestFetchAzurePRCommits_SingleCallWhenTokenEmpty(t *testing.T) {
	client := &fakePRCommitsClient{pages: []*git.GetPullRequestCommitsResponseValue{page("", "sha1")}}

	refs, err := fetchAzurePRCommits(context.Background(), client, git.GetPullRequestCommitsArgs{}, defaultMaxPages)
	require.NoError(t, err)
	require.Len(t, refs, 1)
	require.Len(t, client.calls, 1)
}

func TestFetchAzurePRCommits_ErrorsWhenTokenRepeats(t *testing.T) {
	client := &fakePRCommitsClient{pages: []*git.GetPullRequestCommitsResponseValue{
		page("stuck", "sha1"),
		page("stuck", "sha2"),
	}}

	_, err := fetchAzurePRCommits(context.Background(), client, git.GetPullRequestCommitsArgs{}, defaultMaxPages)
	require.ErrorContains(t, err, "did not advance")
	require.Len(t, client.calls, 2)
}

func TestFetchAzurePRCommits_ErrorsPastMaxPages(t *testing.T) {
	client := &fakePRCommitsClient{pages: []*git.GetPullRequestCommitsResponseValue{
		page("a", "sha1"), page("b", "sha2"), page("c", "sha3"),
	}}

	_, err := fetchAzurePRCommits(context.Background(), client, git.GetPullRequestCommitsArgs{}, 2)
	require.ErrorContains(t, err, "2 pages")
	require.Len(t, client.calls, 2)
}

func TestFetchAzurePRCommits_PropagatesError(t *testing.T) {
	wantErr := errors.New("boom")
	client := &fakePRCommitsClient{err: wantErr}

	_, err := fetchAzurePRCommits(context.Background(), client, git.GetPullRequestCommitsArgs{}, defaultMaxPages)
	require.ErrorIs(t, err, wantErr)
}

func TestGetPullRequestCommits_MapsEveryPage(t *testing.T) {
	client := &fakePRCommitsClient{pages: []*git.GetPullRequestCommitsResponseValue{
		page("tok1", "sha1"),
		page("", "sha2"),
	}}
	original := newPRCommitsClient
	newPRCommitsClient = func(context.Context, string, string) (prCommitsClient, error) { return client, nil }
	t.Cleanup(func() { newPRCommitsClient = original })

	prID, sourceRef := 5, "refs/heads/feature"
	config := &AzureConfig{Token: "t", OrgURL: "https://dev.azure.com/o", Project: "p", Repository: "r"}
	commits, err := config.GetPullRequestCommits(git.GitPullRequest{
		PullRequestId: &prID, SourceRefName: &sourceRef,
	})

	require.NoError(t, err)
	require.Len(t, commits, 2)
	require.Equal(t, "sha1", commits[0].SHA)
	require.Equal(t, "sha2", commits[1].SHA)
	require.Equal(t, sourceRef, commits[0].Branch)
	require.Equal(t, "Ada <ada@example.com>", commits[0].Author)
}

// A nil PullRequestId is the one input the SDK rejects outright
// (git/client.go:3434), so it is also the only input that reaches the error
// wrap without an id — which must therefore not dereference it.
func TestGetPullRequestCommits_NilPullRequestIDReturnsErrorNotPanic(t *testing.T) {
	client := &fakePRCommitsClient{err: errors.New("args.PullRequestId")}
	original := newPRCommitsClient
	newPRCommitsClient = func(context.Context, string, string) (prCommitsClient, error) { return client, nil }
	t.Cleanup(func() { newPRCommitsClient = original })

	sourceRef := "refs/heads/feature"
	config := &AzureConfig{Token: "t", OrgURL: "https://dev.azure.com/o", Project: "p", Repository: "r"}

	_, err := config.GetPullRequestCommits(git.GitPullRequest{SourceRefName: &sourceRef})

	require.ErrorContains(t, err, "args.PullRequestId", "the SDK's reason must survive")
	require.ErrorContains(t, err, "pull request", "the wrap must still say what failed")
}
