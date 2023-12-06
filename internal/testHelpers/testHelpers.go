package testHelpers

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/require"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
)

func SkipIfEnvVarUnset(T *testing.T, requiredEnvVars []string) {
	for _, envVar := range requiredEnvVars {
		_, ok := os.LookupEnv(envVar)
		if !ok {
			T.Logf("skipping %s as %s is unset in environment", T.Name(), envVar)
			T.Skipf("requires %s", envVar)
		}
	}
}

// Originally we had commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6
// This worked for a while, but now the PR is no longer found by the github api
// for reasons we cannot fathom
// We are now using an even older commit, which currently works.
func GithubCommitWithPR() string {
	return "e21a8afff429e0c87ee523d683f2438113f0a105"
}

func CloneGitRepo(url, cloneTo string) (*git.Repository, error) {
	// the repo worktree filesystem. It has to be osfs so that we can give it a path
	fs := osfs.New(cloneTo)
	// the filesystem for git database
	storerFS := osfs.New(filepath.Join(cloneTo, ".git"))
	storer := filesystem.NewStorage(storerFS, cache.NewObjectLRUDefault())
	return git.Clone(storer, fs, &git.CloneOptions{URL: url})
}

func InitializeGitRepo(repoPath string) (*git.Repository, *git.Worktree, billy.Filesystem, error) {
	// the repo worktree filesystem. It has to be osfs so that we can give it a path
	fs := osfs.New(repoPath)
	// the filesystem for git database
	storerFS := osfs.New(filepath.Join(repoPath, ".git"))
	storer := filesystem.NewStorage(storerFS, cache.NewObjectLRUDefault())
	// initialize the git repo at the filesystem "fs" and using "storer" as the git database
	repo, err := git.Init(storer, fs)
	if err != nil {
		return repo, nil, fs, err
	}

	w, err := repo.Worktree()
	if err != nil {
		return repo, nil, fs, err
	}

	_, err = CommitToRepo(w, fs, "Initial Commit")
	if err != nil {
		return repo, w, fs, err
	}

	return repo, w, fs, nil
}

func CommitToRepo(w *git.Worktree, fs billy.Filesystem, commitMessage string) (string, error) {
	filePath := fmt.Sprintf("file-%d.txt", time.Now().UnixNano())
	newFile, err := fs.Create(filePath)
	if err != nil {
		return "", err
	}
	_, err = newFile.Write([]byte("this is a dummy line"))
	if err != nil {
		return "", err
	}
	err = newFile.Close()
	if err != nil {
		return "", err
	}
	_, err = w.Add(filePath)
	if err != nil {
		return "", err
	}
	hash, err := w.Commit(commitMessage, &git.CommitOptions{})
	if err != nil {
		return "", err
	}

	return hash.String(), nil
}

func CheckoutNewBranch(w *git.Worktree, branchName string) error {
	return w.Checkout(&git.CheckoutOptions{
		Create: true,
		Branch: plumbing.NewBranchReferenceName(branchName),
	})
}

func CheckoutMaster(workTree *git.Worktree, t *testing.T) {
	err := workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.Master,
	})
	require.NoError(t, err)
}
