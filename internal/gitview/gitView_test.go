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

func (suite *GitViewTestSuite) TestNewGitView() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), gv)
	require.Equal(suite.T(), worktree.Filesystem.Root(), gv.repositoryRoot)

	_, err = New(filepath.Join(suite.tmpDir, "non-existing"))
	require.Error(suite.T(), err)
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
			name:                    "can list commits when the repo has 4 commits and newest != oldest",
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

func (suite *GitViewTestSuite) TestChangeLog() {
	for i, t := range []struct {
		name                    string
		currentCommit           string
		previousCommit          string
		commitsNumber           int
		expectedNumberOfCommits int
		expectError             bool
	}{
		{
			name:                    "can get changelog when the repo has only one commit and current == previous",
			commitsNumber:           1,
			currentCommit:           "HEAD",
			previousCommit:          "HEAD",
			expectedNumberOfCommits: 1,
		},
		{
			name:                    "can get changelog when the repo has 3 commits and current == previous",
			commitsNumber:           3,
			currentCommit:           "HEAD",
			previousCommit:          "HEAD",
			expectedNumberOfCommits: 1,
		},
		{
			name:                    "can get changelog when the repo has 3 commits and current != previous",
			commitsNumber:           3,
			currentCommit:           "HEAD",
			previousCommit:          "HEAD~2",
			expectedNumberOfCommits: 2,
		},
		{
			name:                    "can get changelog when the repo has 4 commits and current != previous",
			commitsNumber:           4,
			currentCommit:           "HEAD",
			previousCommit:          "HEAD~1",
			expectedNumberOfCommits: 1,
		},
		{
			name:                    "when previous commit cannot be resolved, the current commit alone is returned",
			commitsNumber:           1,
			currentCommit:           "HEAD",
			previousCommit:          "HEAD~2",
			expectedNumberOfCommits: 1,
		},
		{
			name:           "fails when current commit cannot be resolved",
			commitsNumber:  1,
			currentCommit:  "HEAD~2",
			previousCommit: "HEAD",
			expectError:    true,
		},
		{
			name:                    "can get changelog when previous commit is not supplied",
			commitsNumber:           2,
			currentCommit:           "HEAD",
			expectedNumberOfCommits: 1,
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
			commitsInfo, err := gv.ChangeLog(t.currentCommit, t.previousCommit, suite.logger)
			if t.expectError {
				require.Error(suite.T(), err)
			} else {
				require.Len(suite.T(), commitsInfo, t.expectedNumberOfCommits)
			}
		})
	}
}

func (suite *GitViewTestSuite) TestRepoURL() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.T(), err)
	// the created repo does not have origin remote yet
	_, err = gv.RepoUrl()
	require.Error(suite.T(), err)
	expectedError := fmt.Sprintf("remote('origin') is not found in git repository: %s", gv.repositoryRoot)
	require.Equal(suite.T(), expectedError, err.Error())
}

func (suite *GitViewTestSuite) TestExtractRepoURLFromRemote() {
	for _, t := range []struct {
		name      string
		remoteURL string
		want      string
	}{
		{
			name:      "SSH remote",
			remoteURL: "git@github.com:kosli-dev/cli.git",
			want:      "https://github.com/kosli-dev/cli",
		},
		{
			name:      "HTTP remote",
			remoteURL: "https://github.com/kosli-dev/cli.git",
			want:      "https://github.com/kosli-dev/cli",
		},
	} {
		suite.Run(t.name, func() {
			actual := extractRepoURLFromRemote(t.remoteURL)
			require.Equal(suite.T(), t.want, actual)
		})
	}
}

func (suite *GitViewTestSuite) TestNewCommitInfoFromGitCommit() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.T(), err)

	_, err = gv.newCommitInfoFromGitCommit("58a9461c5a42d83bd5731485a72ddae542ac99d8")
	require.Error(suite.T(), err)
	expected := "failed to resolve git reference 58a9461c5a42d83bd5731485a72ddae542ac99d8: reference not found"
	require.Equal(suite.T(), expected, err.Error())

	_, err = gv.newCommitInfoFromGitCommit("HEAD~2")
	require.Error(suite.T(), err)
	expected = "failed to resolve git reference HEAD~2: EOF"
	require.Equal(suite.T(), expected, err.Error())

	ci, err := gv.newCommitInfoFromGitCommit("HEAD")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Added file 1", ci.Message)
	require.Equal(suite.T(), "master", ci.Branch)
	require.Empty(suite.T(), ci.Parents)
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
