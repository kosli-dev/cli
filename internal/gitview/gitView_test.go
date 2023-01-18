package gitview

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/cache"
	"github.com/go-git/go-git/v5/storage/filesystem"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type GitViewTestSuite struct {
	suite.Suite
	tmpDir string
	logger *logger.Logger
}

func (suite *GitViewTestSuite) SetupSuite() {
	suite.logger = logger.NewStandardLogger()
}

func (suite *GitViewTestSuite) SetupTest() {
	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testRepoDir")
	require.NoError(suite.T(), err, "error creating a temporary test directory")
}

// clean up tmpDir after each test
func (suite *GitViewTestSuite) AfterTest() {
	err := os.RemoveAll(suite.tmpDir)
	require.NoErrorf(suite.T(), err, "error cleaning up the temporary test directory %s", suite.tmpDir)
}

func (suite *GitViewTestSuite) TestCommitsBetween() {
	for i, t := range []struct {
		name                    string
		newestCommit            string
		oldestCommit            string
		commitsNumber           int
		expectedNumberOfCommits int
		expectError             bool
	}{
		{
			name:                    "can list commits when the repo has only one commit and newest == oldest",
			commitsNumber:           1,
			newestCommit:            "HEAD",
			oldestCommit:            "HEAD",
			expectedNumberOfCommits: 1,
		},
		{
			name:                    "can list commits when the repo has 3 commits and newest == oldest",
			commitsNumber:           3,
			newestCommit:            "HEAD",
			oldestCommit:            "HEAD",
			expectedNumberOfCommits: 1,
		},
		{
			name:                    "can list commits when the repo has 3 commits and newest != oldest",
			commitsNumber:           3,
			newestCommit:            "HEAD",
			oldestCommit:            "HEAD~2",
			expectedNumberOfCommits: 2,
		},
		{
			name:                    "can list commits when the repo has 3 commits and newest != oldest",
			commitsNumber:           4,
			newestCommit:            "HEAD",
			oldestCommit:            "HEAD~1",
			expectedNumberOfCommits: 1,
		},
		{
			name:          "fails when oldest commit cannot be resolved",
			commitsNumber: 1,
			newestCommit:  "HEAD",
			oldestCommit:  "HEAD~2",
			expectError:   true,
		},
		{
			name:          "fails when newest commit cannot be resolved",
			commitsNumber: 1,
			newestCommit:  "HEAD~2",
			oldestCommit:  "HEAD",
			expectError:   true,
		},
	} {
		suite.Run(t.name, func() {
			repoName := fmt.Sprintf("test-%d", i)
			dirPath := filepath.Join(suite.tmpDir, repoName)
			_, worktree, err := initializeRepoAndCommit(dirPath, t.commitsNumber)
			require.NoErrorf(suite.T(), err, "error creating test repository %s", repoName)
			// suite.T().Logf("repo dir is: %s", worktree.Filesystem.Root())

			gv, err := New(worktree.Filesystem.Root())
			require.NoError(suite.T(), err)
			commits, err := gv.CommitsBetween(t.oldestCommit, t.newestCommit, suite.logger)
			if t.expectError {
				require.Error(suite.T(), err)
			} else {
				require.Len(suite.T(), commits, t.expectedNumberOfCommits)
			}
		})
	}
}

func initializeRepoAndCommit(repoPath string, commitsNumber int) (*git.Repository, *git.Worktree, error) {
	// the repo worktree filesystem. It has to be osfs so that we can give it a path
	fs := osfs.New(repoPath)
	// the filesystem for git database
	storerFS := osfs.New(filepath.Join(repoPath, ".git"))
	storer := filesystem.NewStorage(storerFS, cache.NewObjectLRUDefault())
	// initialize the git repo at the filesystem "fs" and using "storer" as the git database
	repo, err := git.Init(storer, fs)
	if err != nil {
		return repo, nil, err
	}

	w, err := repo.Worktree()
	if err != nil {
		return repo, nil, err
	}

	for i := 1; i <= commitsNumber; i++ {
		filePath := fmt.Sprintf("file-%d.txt", i)
		newFile, err := fs.Create(filePath)
		if err != nil {
			return repo, w, err
		}
		_, err = newFile.Write([]byte("this is a dummy line"))
		if err != nil {
			return repo, w, err
		}
		err = newFile.Close()
		if err != nil {
			return repo, w, err
		}
		_, err = w.Add(filePath)
		if err != nil {
			return repo, w, err
		}
		_, err = w.Commit(fmt.Sprintf("Added file %d", i), &git.CommitOptions{})
		if err != nil {
			return repo, w, err
		}
	}

	return repo, w, nil
}

func TestGitViewTestSuite(t *testing.T) {
	suite.Run(t, new(GitViewTestSuite))
}
