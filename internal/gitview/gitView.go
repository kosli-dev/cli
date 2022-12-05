package gitview

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"strings"
)

type ArtifactCommit struct {
	Sha1      string   `json:"sha1"`
	Message   string   `json:"message"`
	Author    string   `json:"author"`
	Timestamp int64    `json:"timestamp"`
	Branch    string   `json:"branch"`
	Parents   []string `json:"parents"`
}

//
// This type should replace all direct access to the git library
// Currently, artifactCreation.go and pipelineBackfillCommits.go
//

type GitView struct {
	repositoryRoot string
	repository     *git.Repository
}

// Open opens a git repository from the given path. It detects if the
// repository is bare or a normal one. If the path doesn't contain a valid
// repository ErrRepositoryNotExists is returned

func New(repositoryRoot string) (*GitView, error) {
	repository, err := git.PlainOpen(repositoryRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open git repository at %s: %v", repositoryRoot, err)
	}
	return &GitView{
		repositoryRoot: repositoryRoot,
		repository:     repository,
	}, nil
}

// CommitsBetween list all commits that have happened between two commits in a git repo
func (gv *GitView) CommitsBetween(oldest, newest string) ([]*ArtifactCommit, error) {
	// Using 'var commits []*ArtifactCommit' will make '[]' convert to 'null' when converting to json
	// which will fail on the server side.
	// Using 'commits := make([]*ArtifactCommit, 0)' will make '[]' convert to '[]' when converting to json
	// See issue #522
	commits := make([]*ArtifactCommit, 0)

	branchName, err := gv.branchName()
	if err != nil {
		return commits, err
	}

	newestHash, err := gv.repository.ResolveRevision(plumbing.Revision(newest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve %s: %v", newest, err)
	}
	oldestHash, err := gv.repository.ResolveRevision(plumbing.Revision(oldest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve %s: %v", oldest, err)
	}

	//log.Debugf("This is the newest commit hash %s", newestHash.String())
	//log.Debugf("This is the oldest commit hash %s", oldestHash.String())

	commitsIter, err := gv.repository.Log(&git.LogOptions{From: *newestHash, Order: git.LogOrderCommitterTime})
	if err != nil {
		return commits, fmt.Errorf("failed to git log: %v", err)
	}

	for {
		commit, err := commitsIter.Next()
		if err != nil {
			return commits, fmt.Errorf("failed to get next commit: %v", err)
		}
		if commit.Hash != *oldestHash {
			currentCommit := asArtifactCommit(commit, branchName)
			commits = append(commits, currentCommit)
		} else {
			break
		}
	}

	return commits, nil
}

// branchName returns the current branch name on a repository,
// or an error if the repo head is not on a branch
func (gv *GitView) branchName() (string, error) {
	head, err := gv.repository.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get the current HEAD of the git repository: %v", err)
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "", nil
}

// asArtifactCommit returns an ArtifactCommit from a git Commit object
func asArtifactCommit(commit *object.Commit, branchName string) *ArtifactCommit {
	var commitParents []string
	for _, h := range commit.ParentHashes {
		commitParents = append(commitParents, h.String())
	}
	return &ArtifactCommit{
		Sha1:      commit.Hash.String(),
		Message:   strings.TrimSpace(commit.Message),
		Author:    commit.Author.String(),
		Timestamp: commit.Author.When.UTC().Unix(),
		Branch:    branchName,
		Parents:   commitParents,
	}
}
