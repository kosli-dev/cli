package gitview

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/kosli-dev/cli/internal/logger"
)

type BasicCommitInfo struct {
	Sha1      string `json:"sha1"`
	Message   string `json:"message"`
	Author    string `json:"author"`
	Timestamp int64  `json:"timestamp"`
	Branch    string `json:"branch"`
}

type CommitInfo struct {
	BasicCommitInfo
	Parents []string `json:"parents"`
}

// GitView
// A read-only view of a git repository.
type GitView struct {
	repositoryRoot string
	repository     *git.Repository
}

// New opens a git repository from the given path. It detects if the
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
func (gv *GitView) CommitsBetween(oldest, newest string, logger *logger.Logger) ([]*CommitInfo, error) {
	// Using 'var commits []*ArtifactCommit' will make '[]' convert to 'null' when converting to json
	// which will fail on the server side.
	// Using 'commits := make([]*ArtifactCommit, 0)' will make '[]' convert to '[]' when converting to json
	// See issue #522
	commits := make([]*CommitInfo, 0)

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

	if oldestHash.String() == newestHash.String() {
		commitObject, err := gv.repository.CommitObject(*newestHash)
		if err != nil {
			return commits, err
		}
		commit := asCommitInfo(commitObject, branchName)
		commits = append(commits, commit)

	} else {
		commitsIter, err := gv.repository.Log(&git.LogOptions{From: *newestHash, Order: git.LogOrderCommitterTime})
		if err != nil {
			return commits, fmt.Errorf("failed to git log: %v", err)
		}

		for {
			commit, err := commitsIter.Next()
			if err != nil {
				return commits, fmt.Errorf("failed to get next git commit: %v\n%s", err, hint)
			}
			if commit.Hash != *oldestHash {
				nextCommit := asCommitInfo(commit, branchName)
				commits = append(commits, nextCommit)
			} else {
				break
			}
		}
	}

	logger.Debug("parsed %d commits between newest and oldest git commits", len(commits))
	return commits, nil
}

// RepoUrl returns HTTPS URL for the `origin` remote of a repo
func (gv *GitView) RepoUrl() (string, error) {
	repoRemote, err := gv.repository.Remote("origin") // TODO: We hard code this for now. Should we have a flag to set it from the cmdline? 2022-12-06
	if err != nil {
		return "", fmt.Errorf("remote('origin') is not found in git repository: %s", gv.repositoryRoot)
	}
	remoteUrl := ExtractRepoURLFromRemote(repoRemote.Config().URLs[0])
	return remoteUrl, nil
}

// ExtractRepoURLFromRemote converts an SSH or http remote into a URL
func ExtractRepoURLFromRemote(remoteUrl string) string {
	if strings.HasPrefix(remoteUrl, "git@") {
		remoteUrl = strings.Replace(remoteUrl, ":", "/", 1)
		remoteUrl = strings.Replace(remoteUrl, "git@", "https://", 1)
	}
	remoteUrl = strings.Replace(remoteUrl, ".git", "", 1)
	return remoteUrl
}

// ChangeLog attempts to collect the changelog list of commits for an artifact,
// the changelog is all commits between current commit and the commit from which the previous artifact in Kosli
// was created.
// If collecting the changelog fails (e.g. if git history has been rewritten, or the clone depth is too shallow),
// the changelog only contains the single commit info which is the current commit
func (gv *GitView) ChangeLog(currentCommit, previousCommit string, logger *logger.Logger) ([]*CommitInfo, error) {
	if previousCommit != "" {
		commitsList, err := gv.CommitsBetween(previousCommit, currentCommit, logger)
		if err != nil {
			logger.Warning(err.Error())
		} else {
			return commitsList, nil
		}
	}

	currentArtifactCommit, err := gv.GetCommitInfoFromCommitSHA(currentCommit)
	if err != nil {
		return []*CommitInfo{}, fmt.Errorf("could not retrieve current git commit for %s: %v", currentCommit, err)
	}
	return []*CommitInfo{currentArtifactCommit}, nil
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

// GetCommitInfoFromCommitSHA returns a CommitInfo object from a git commit
// the gitCommit can be SHA1 or a revision: e.g. HEAD or HEAD~2 etc
func (gv *GitView) GetCommitInfoFromCommitSHA(gitCommit string) (*CommitInfo, error) {
	branchName, err := gv.BranchName()
	if err != nil {
		return &CommitInfo{}, err
	}

	hash, err := gv.repository.ResolveRevision(plumbing.Revision(gitCommit))
	if err != nil {
		return &CommitInfo{}, fmt.Errorf("failed to resolve git reference %s: %v", gitCommit, err)
	}
	commit, err := gv.repository.CommitObject(*hash)
	if err != nil {
		return &CommitInfo{}, fmt.Errorf("could not retrieve commit for %s: %v", *hash, err)
	}

	return asCommitInfo(commit, branchName), nil
}

// asCommitInfo returns a CommitInfo from a git Commit object
func asCommitInfo(commit *object.Commit, branchName string) *CommitInfo {
	commitParents := []string{}
	for _, hash := range commit.ParentHashes {
		commitParents = append(commitParents, hash.String())
	}
	return &CommitInfo{
		BasicCommitInfo: BasicCommitInfo{
			Sha1:      commit.Hash.String(),
			Message:   strings.TrimSpace(commit.Message),
			Author:    commit.Author.String(),
			Timestamp: commit.Author.When.UTC().Unix(),
			Branch:    branchName,
		},
		Parents: commitParents,
	}
}

// MatchPatternInCommitMessageORBranchName returns a slice of strings matching a pattern in a commit message or branch name
// matches lookup happens in the commit message first, and if none is found, matching against the branch name is done
// if no matches are found in both the commit message and the branch name, an empty slice is returned
func (gv *GitView) MatchPatternInCommitMessageORBranchName(pattern, commitSHA string) ([]string, *CommitInfo, error) {
	commitInfo, err := gv.GetCommitInfoFromCommitSHA(commitSHA)
	if err != nil {
		return []string{}, nil, err
	}

	re := regexp.MustCompile(pattern)
	matches := re.FindAllString(commitInfo.Message, -1)
	if matches != nil {
		return matches, commitInfo, nil
	} else {
		matches := re.FindAllString(commitInfo.Branch, -1)
		if matches != nil {
			return matches, commitInfo, nil
		}
	}
	return []string{}, commitInfo, nil
}

// ResolveRevision returns an explicit commit SHA1 from commit SHA or ref (e.g. HEAD~2)
func (gv *GitView) ResolveRevision(commitSHAOrRef string) (string, error) {
	hash, err := gv.repository.ResolveRevision(plumbing.Revision(commitSHAOrRef))
	if err != nil {
		return "", err
	}
	return hash.String(), nil
}
