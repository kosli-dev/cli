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
	"github.com/kosli-dev/cli/internal/testHelpers"
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
			actual := ExtractRepoURLFromRemote(t.remoteURL)
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

	_, err = gv.GetCommitInfoFromCommitSHA("58a9461c5a42d83bd5731485a72ddae542ac99d8")
	require.Error(suite.T(), err)
	expected := "failed to resolve git reference 58a9461c5a42d83bd5731485a72ddae542ac99d8: reference not found"
	require.Equal(suite.T(), expected, err.Error())

	_, err = gv.GetCommitInfoFromCommitSHA("HEAD~2")
	require.Error(suite.T(), err)
	expected = "failed to resolve git reference HEAD~2: EOF"
	require.Equal(suite.T(), expected, err.Error())

	ci, err := gv.GetCommitInfoFromCommitSHA("HEAD")
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Added file 1", ci.Message)
	require.Equal(suite.T(), "master", ci.Branch)
	require.Empty(suite.T(), ci.Parents)
}

func (suite *GitViewTestSuite) TestMatchPatternInCommitMessageORBranchName() {
	_, workTree, fs, err := testHelpers.InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.T(), err)

	for _, t := range []struct {
		name          string
		pattern       string
		commitMessage string
		wantError     bool
		want          []string
		commitSha     string
	}{
		{
			name:          "One Jira reference found",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "EX-1 test commit",
			want:          []string{"EX-1"},
			wantError:     false,
		},
		{
			name:          "Two Jira references found",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "EX-1 ABC-22 test commit",
			want:          []string{"EX-1", "ABC-22"},
			wantError:     false,
		},
		{
			name:          "No Jira references found",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "test commit",
			want:          []string{},
			wantError:     false,
		},
		{
			name:          "No Jira references found, despite something that looks similar to Jira reference",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "Ea-1 test commit",
			want:          []string{},
			wantError:     false,
		},
		{
			name:      "Commit not found, expect an error",
			pattern:   "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitSha: "3b7420d0392114794591aaefcd84d7b100b8d095",
			wantError: true,
		},
		{
			name:          "GitHub reference found",
			pattern:       "#[0-9]+",
			commitMessage: "#324 test commit",
			want:          []string{"#324"},
			wantError:     false,
		},
	} {
		suite.Run(t.name, func() {

			if t.commitSha == "" {
				t.commitSha, err = testHelpers.CommitToRepo(workTree, fs, t.commitMessage)
				require.NoError(suite.T(), err)
			}

			gitView, err := New(suite.tmpDir)
			require.NoError(suite.T(), err)

			actual, _, err := gitView.MatchPatternInCommitMessageORBranchName(t.pattern, t.commitSha)
			require.True(suite.T(), (err != nil) == t.wantError)
			require.ElementsMatch(suite.T(), t.want, actual)

		})
	}
}

func (suite *GitViewTestSuite) TestResolveRevision() {
	_, workTree, fs, err := testHelpers.InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.T(), err)

	FirstCommitSha, err := testHelpers.CommitToRepo(workTree, fs, "Test commit message 1")
	require.NoError(suite.T(), err)

	SecondCommitSha, err := testHelpers.CommitToRepo(workTree, fs, "Test commit message 2")
	require.NoError(suite.T(), err)

	ThirdCommitSha, err := testHelpers.CommitToRepo(workTree, fs, "Test commit message 3")
	require.NoError(suite.T(), err)

	for _, t := range []struct {
		name           string
		commitSHAOrRef string
		wantError      bool
		want           string
	}{
		{
			name:           "HEAD reference resolved",
			commitSHAOrRef: "HEAD",
			want:           ThirdCommitSha,
			wantError:      false,
		},
		{
			name:           "~1 reference resolved",
			commitSHAOrRef: "HEAD~1",
			want:           SecondCommitSha,
			wantError:      false,
		},
		{
			name:           "^^ reference resolved",
			commitSHAOrRef: "HEAD^^",
			want:           FirstCommitSha,
			wantError:      false,
		},
		{
			name:           "Short sha reference resolved",
			commitSHAOrRef: ThirdCommitSha[0:7],
			want:           ThirdCommitSha,
			wantError:      false,
		},
		{
			name:           "Fail if sha not found",
			commitSHAOrRef: "123456",
			wantError:      true,
		},
	} {
		suite.Run(t.name, func() {

			gitView, err := New(suite.tmpDir)
			require.NoError(suite.T(), err)

			actual, err := gitView.ResolveRevision(t.commitSHAOrRef)
			require.True(suite.T(), (err != nil) == t.wantError)
			require.Equal(suite.T(), t.want, actual)

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
