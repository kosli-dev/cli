package gitview

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/go-git/go-billy/v5/osfs"
	git "github.com/go-git/go-git/v5"
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

func (suite *GitViewTestSuite) TestNewGitViewFromWorktree() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, _, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.T(), err)

	worktreePath := filepath.Join(suite.tmpDir, "myWorktree")
	cmd := exec.Command("git", "worktree", "add", "-b", "worktree-branch", worktreePath)
	cmd.Dir = dirPath
	output, err := cmd.CombinedOutput()
	require.NoError(suite.T(), err, "git worktree add failed: %s", string(output))

	gv, err := New(worktreePath)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), gv)

	branchName, err := gv.BranchName()
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "worktree-branch", branchName)

	commitInfo, err := gv.GetCommitInfoFromCommitSHA("HEAD", true, []string{})
	require.NoError(suite.T(), err)
	require.NotEmpty(suite.T(), commitInfo.Sha1)
	require.Equal(suite.T(), "worktree-branch", commitInfo.Branch)
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
	_, err = gv.RepoURL()
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
		{
			name:      "HTTP remote with username and password",
			remoteURL: "https://kosli:xxx@github.com/kosli-dev/cli.git",
			want:      "https://github.com/kosli-dev/cli",
		},
	} {
		suite.Run(t.name, func() {
			actual, _ := ExtractRepoURLFromRemote(t.remoteURL)
			require.Equal(suite.T(), t.want, actual)
		})
	}
}

func (suite *GitViewTestSuite) TestRemoveUsernamePasswordFromURL() {
	for _, t := range []struct {
		name      string
		inputURL  string
		want      string
		wantError bool
	}{
		{
			name:     "url with username",
			inputURL: "https://kosli@dev.azure.com/kosli/kosli-azure/_git/cli",
			want:     "https://dev.azure.com/kosli/kosli-azure/_git/cli",
		},
		{
			name:     "url with username and password",
			inputURL: "https://kosli:xxxx@dev.azure.com/kosli/kosli-azure/_git/cli",
			want:     "https://dev.azure.com/kosli/kosli-azure/_git/cli",
		},
		{
			name:     "clean url",
			inputURL: "https://dev.azure.com/kosli/kosli-azure/_git/cli",
			want:     "https://dev.azure.com/kosli/kosli-azure/_git/cli",
		},
		{
			name:      "invalid url returns error",
			inputURL:  "://not.a url@",
			wantError: true,
		},
	} {
		suite.Run(t.name, func() {
			actual, err := removeUsernamePasswordFromURL(t.inputURL)
			require.Equal(suite.T(), t.wantError, err != nil)
			require.Equal(suite.T(), t.want, actual)
		})
	}
}

func (suite *GitViewTestSuite) TestGetCommitURL() {
	for _, t := range []struct {
		name       string
		repoURL    string
		commitHash string
		want       string
	}{
		{
			name:       "github",
			repoURL:    "https://github.com/kosli-dev/cli",
			commitHash: "089615f84caedd6280689da694e71052cbdfb84d",
			want:       "https://github.com/kosli-dev/cli/commit/089615f84caedd6280689da694e71052cbdfb84d",
		},
		{
			name:       "gitlab",
			repoURL:    "https://gitlab.com/kosli/merkely-gitlab-demo",
			commitHash: "089615f84caedd6280689da694e71052cbdfb84d",
			want:       "https://gitlab.com/kosli/merkely-gitlab-demo/-/commit/089615f84caedd6280689da694e71052cbdfb84d",
		},
		{
			name:       "bitbucket",
			repoURL:    "https://bitbucket.org/kosli-dev/cli-test",
			commitHash: "089615f84caedd6280689da694e71052cbdfb84d",
			want:       "https://bitbucket.org/kosli-dev/cli-test/commits/089615f84caedd6280689da694e71052cbdfb84d",
		},
		{
			name:       "azure",
			repoURL:    "https://dev.azure.com/kosli/kosli-azure/_git/cli",
			commitHash: "089615f84caedd6280689da694e71052cbdfb84d",
			want:       "https://dev.azure.com/kosli/kosli-azure/_git/cli/commit/089615f84caedd6280689da694e71052cbdfb84d",
		},
		{
			name:       "github enterprise",
			repoURL:    "https://custom-domain-name.com/kosli-dev/cli",
			commitHash: "089615f84caedd6280689da694e71052cbdfb84d",
			want:       "https://custom-domain-name.com/kosli-dev/cli/commit/089615f84caedd6280689da694e71052cbdfb84d",
		},
	} {
		suite.Run(t.name, func() {
			actual := getCommitURL(t.repoURL, t.commitHash)
			require.Equal(suite.T(), t.want, actual)
		})
	}
}

func (suite *GitViewTestSuite) TestGetCommitInfoFromCommitSHA() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.T(), err)

	_, err = gv.GetCommitInfoFromCommitSHA("58a9461c5a42d83bd5731485a72ddae542ac99d8", true, []string{})
	require.Error(suite.T(), err)
	expected := "failed to resolve git reference 58a9461c5a42d83bd5731485a72ddae542ac99d8: reference not found"
	require.Equal(suite.T(), expected, err.Error())

	_, err = gv.GetCommitInfoFromCommitSHA("HEAD~2", true, []string{})
	require.Error(suite.T(), err)
	expected = "failed to resolve git reference HEAD~2: EOF"
	require.Equal(suite.T(), expected, err.Error())

	commitInfo, err := gv.GetCommitInfoFromCommitSHA("HEAD", false, []string{})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "Added file 1", commitInfo.Message)
	require.Equal(suite.T(), "master", commitInfo.Branch)
	require.Empty(suite.T(), commitInfo.Parents)
	require.Empty(suite.T(), commitInfo.URL)

	commitInfo, err = gv.GetCommitInfoFromCommitSHA("HEAD", false, []string{"author", "message", "branch"})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), redactedCommitInfoValue, commitInfo.Author)
	require.Equal(suite.T(), redactedCommitInfoValue, commitInfo.Message)
	require.Equal(suite.T(), redactedCommitInfoValue, commitInfo.Branch)
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

func (suite *GitViewTestSuite) TestGetTrailerValues() {
	for _, tt := range []struct {
		name     string
		message  string
		key      string
		expected []string
	}{
		{
			name:     "no trailers returns empty slice",
			message:  "fix: something\n\nsome body text",
			key:      "Jira",
			expected: []string{},
		},
		{
			name:     "single matching trailer",
			message:  "fix: something\n\nJira: BX-123",
			key:      "Jira",
			expected: []string{"BX-123"},
		},
		{
			name:     "key match is case-insensitive",
			message:  "fix: something\n\njira: BX-123",
			key:      "Jira",
			expected: []string{"BX-123"},
		},
		{
			name:     "multiple occurrences of same key",
			message:  "fix: something\n\nJira: BX-123\nJira: BX-456",
			key:      "Jira",
			expected: []string{"BX-123", "BX-456"},
		},
		{
			name:     "non-matching trailers are ignored",
			message:  "fix: something\n\nJira: BX-123\nOna-Environment-Id: ONA-456",
			key:      "Jira",
			expected: []string{"BX-123"},
		},
		{
			name:     "whitespace trimmed from value",
			message:  "fix: something\n\nJira:   BX-123  ",
			key:      "Jira",
			expected: []string{"BX-123"},
		},
		{
			name:     "leading whitespace on line is tolerated",
			message:  "fix: something\n\n    Jira: BX-123",
			key:      "Jira",
			expected: []string{"BX-123"},
		},
		{
			name:     "key supplied with trailing colon still matches",
			message:  "fix: something\n\nJira: BX-123",
			key:      "Jira:",
			expected: []string{"BX-123"},
		},
		{
			name:     "key with surrounding whitespace still matches",
			message:  "fix: something\n\nJira: BX-123",
			key:      "  Jira  ",
			expected: []string{"BX-123"},
		},
		{
			name:     "line with empty value is skipped",
			message:  "fix: something\n\nJira:",
			key:      "Jira",
			expected: []string{},
		},
		{
			name:     "key whose lowercase is shorter than the original still extracts correct value",
			message:  "fix: something\n\nİ: BX-123",
			key:      "İ",
			expected: []string{"BX-123"},
		},
	} {
		suite.Run(tt.name, func() {
			result := GetTrailerValues(tt.message, tt.key)
			require.Equal(suite.T(), tt.expected, result)
		})
	}
}

func TestGitViewTestSuite(t *testing.T) {
	suite.Run(t, new(GitViewTestSuite))
}
