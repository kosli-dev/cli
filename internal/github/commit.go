package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/kosli-dev/cli/internal/types"
)

// CommitFinder looks up commits through GitHub's REST API. An empty BaseURL
// means github.com; set it for GitHub Enterprise Server.
type CommitFinder struct {
	Token   string
	BaseURL string
	Debug   bool
}

// CommitBySHA returns the commit with sha in repository, named owner/name.
func (f CommitFinder) CommitBySHA(repository, sha string) (*types.Commit, error) {
	owner, name, ok := strings.Cut(repository, "/")
	if !ok || owner == "" || name == "" {
		return nil, fmt.Errorf("GitHub repository %q is not in the form owner/name", repository)
	}
	ctx := context.Background()
	client, err := NewGithubClientFromToken(ctx, f.Token, f.BaseURL, f.Debug)
	if err != nil {
		return nil, err
	}
	found, _, err := client.Repositories.GetCommit(ctx, owner, name, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s from GitHub repository %s: %w", sha, repository, err)
	}

	gitAuthor := found.GetCommit().GetAuthor()
	// A commit with no author is recorded as "", not " <>".
	author := ""
	if gitAuthor.GetName() != "" || gitAuthor.GetEmail() != "" {
		author = fmt.Sprintf("%s <%s>", gitAuthor.GetName(), gitAuthor.GetEmail())
	}
	var timestamp int64
	if date := gitAuthor.GetDate(); !date.IsZero() {
		timestamp = date.Unix()
	}
	return &types.Commit{
		SHA:       found.GetSHA(),
		Message:   strings.TrimSpace(found.GetCommit().GetMessage()),
		Author:    author,
		Timestamp: timestamp,
		URL:       found.GetHTMLURL(),
	}, nil
}
