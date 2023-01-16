package gitview

import (
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kosli-dev/cli/internal/logger"
)

type ArtifactCommit struct {
	Sha1      string   `json:"sha1"`
	Message   string   `json:"message"`
	Author    string   `json:"author"`
	Timestamp int64    `json:"timestamp"`
	Branch    string   `json:"branch"`
	Parents   []string `json:"parents"`
}

// GitView
// A read-only view of a git repository.
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
func (gv *GitView) CommitsBetween(oldest, newest string, logger *logger.Logger) ([]*ArtifactCommit, error) {
	// Using 'var commits []*ArtifactCommit' will make '[]' convert to 'null' when converting to json
	// which will fail on the server side.
	// Using 'commits := make([]*ArtifactCommit, 0)' will make '[]' convert to '[]' when converting to json
	// See issue #522

	//TODO: check why this behaves weird, comment in server#556
	commits := make([]*ArtifactCommit, 0)

	branchName, err := gv.BranchName()
	if err != nil {
		return commits, err
	}

	newestHash, err := gv.repository.ResolveRevision(plumbing.Revision(newest))
	hint := "The commit does not exist in the git repository.\nThis may be caused by insufficient git clone depth."
	if err != nil {
		return commits, fmt.Errorf("failed to resolve git reference %s\n%s", newest, hint)
	}
	oldestHash, err := gv.repository.ResolveRevision(plumbing.Revision(oldest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve git reference %s\n%s", oldest, hint)
	}

	logger.Debug("newest commit hash %s", newestHash.String())
	logger.Debug("oldest commit hash %s", oldestHash.String())

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
	logger.Debug("parsed %d commits between newest and oldest commits", len(commits))
	return commits, nil
}

// RepoUrl returns HTTPS URL for the `origin` remote of a repo
func (gv *GitView) RepoUrl() (string, error) {
	repoRemote, err := gv.repository.Remote("origin") // TODO: We hard code this for now. Should we have a flag to set it from the cmdline? 2022-12-06
	if err != nil {
		fmt.Printf("Warning: Repo URL will not be reported since there is no remote('origin') in git repository (%s)\n", gv.repositoryRoot)
		return "", nil
	}
	remoteUrl := repoRemote.Config().URLs[0]
	if strings.HasPrefix(remoteUrl, "git@") {
		remoteUrl = strings.Replace(remoteUrl, ":", "/", 1)
		remoteUrl = strings.Replace(remoteUrl, "git@", "https://", 1)
	}
	remoteUrl = strings.TrimSuffix(remoteUrl, ".git")
	return remoteUrl, nil
}

// ChangeLog attempts to collect the changelog list of commits for an artifact,
// the changelog is all commits between current commit and the commit from which the previous artifact in Kosli
// was created.
// If collecting the changelog fails (e.g. if git history has been rewritten), the changelog only
// contains the single commit info which is the current commit

func (gv *GitView) ChangeLog(currentCommit, previousCommit string, logger *logger.Logger) ([]*ArtifactCommit, error) {
	if previousCommit != "" {
		commitsList, err := gv.CommitsBetween(previousCommit, currentCommit, logger)
		if err != nil {
			fmt.Printf("Warning: %s\n", err)
		} else {
			return commitsList, nil
		}
	}

	currentArtifactCommit, err := gv.newArtifactCommitFromGitCommit(currentCommit)
	if err != nil {
		return []*ArtifactCommit{}, fmt.Errorf("could not retrieve current git commit for %s: %v", currentCommit, err)
	}
	return []*ArtifactCommit{currentArtifactCommit}, nil
}

// BranchName returns the current branch name on a repository,
// or an error if the repo head is not on a branch
func (gv *GitView) BranchName() (string, error) {
	head, err := gv.repository.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get the current HEAD of the git repository: %v", err)
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "", nil
}

// newArtifactCommitFromGitCommit returns an ArtifactCommit object from a git commit
// the gitCommit can be a revision: e.g. HEAD or HEAD~2 etc
func (gv *GitView) newArtifactCommitFromGitCommit(gitCommit string) (*ArtifactCommit, error) {
	branchName, err := gv.BranchName()
	if err != nil {
		return &ArtifactCommit{}, err
	}

	currentHash, err := gv.repository.ResolveRevision(plumbing.Revision(gitCommit))
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("failed to resolve %s: %v", gitCommit, err)
	}
	currentCommit, err := gv.repository.CommitObject(*currentHash)
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("could not retrieve commit for %s: %v", *currentHash, err)
	}

	return asArtifactCommit(currentCommit, branchName), nil
}

// asArtifactCommit returns an ArtifactCommit from a git Commit object
func asArtifactCommit(commit *object.Commit, branchName string) *ArtifactCommit {
	commitParents := []string{}
	for _, hash := range commit.ParentHashes {
		commitParents = append(commitParents, hash.String())
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
