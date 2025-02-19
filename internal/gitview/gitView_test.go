package gitview

import (
	"fmt"
	"os"
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
	require.NoError(suite.Suite.T(), err, "error creating a temporary test directory")
}

// clean up tmpDir after each test
func (suite *GitViewTestSuite) AfterTest() {
	err := os.RemoveAll(suite.tmpDir)
	require.NoErrorf(suite.Suite.T(), err, "error cleaning up the temporary test directory %s", suite.tmpDir)
}

func (suite *GitViewTestSuite) TestNewGitView() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.Suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.Suite.T(), err)
	require.NotNil(suite.Suite.T(), gv)
	require.Equal(suite.Suite.T(), worktree.Filesystem.Root(), gv.repositoryRoot)

	_, err = New(filepath.Join(suite.tmpDir, "non-existing"))
	require.Error(suite.Suite.T(), err)
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
		suite.Suite.Run(t.name, func() {
			repoName := fmt.Sprintf("test-%d", i)
			dirPath := filepath.Join(suite.tmpDir, repoName)
			_, worktree, err := initializeRepoAndCommit(dirPath, t.commitsNumber)
			require.NoErrorf(suite.Suite.T(), err, "error creating test repository %s", repoName)
			// suite.Suite.T().Logf("repo dir is: %s", worktree.Filesystem.Root())

			gv, err := New(worktree.Filesystem.Root())
			require.NoError(suite.Suite.T(), err)
			commits, err := gv.CommitsBetween(t.oldestCommit, t.newestCommit, suite.logger)
			if t.expectError {
				require.Error(suite.Suite.T(), err)
			} else {
				require.Len(suite.Suite.T(), commits, t.expectedNumberOfCommits)
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
		suite.Suite.Run(t.name, func() {
			repoName := fmt.Sprintf("test-%d", i)
			dirPath := filepath.Join(suite.tmpDir, repoName)
			_, worktree, err := initializeRepoAndCommit(dirPath, t.commitsNumber)
			require.NoErrorf(suite.Suite.T(), err, "error creating test repository %s", repoName)
			// suite.Suite.T().Logf("repo dir is: %s", worktree.Filesystem.Root())

			gv, err := New(worktree.Filesystem.Root())
			require.NoError(suite.Suite.T(), err)
			commitsInfo, err := gv.ChangeLog(t.currentCommit, t.previousCommit, suite.logger)
			if t.expectError {
				require.Error(suite.Suite.T(), err)
			} else {
				require.Len(suite.Suite.T(), commitsInfo, t.expectedNumberOfCommits)
			}
		})
	}
}

func (suite *GitViewTestSuite) TestRepoURL() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.Suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.Suite.T(), err)
	// the created repo does not have origin remote yet
	_, err = gv.RepoURL()
	require.Error(suite.Suite.T(), err)
	expectedError := fmt.Sprintf("remote('origin') is not found in git repository: %s", gv.repositoryRoot)
	require.Equal(suite.Suite.T(), expectedError, err.Error())
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
		suite.Suite.Run(t.name, func() {
			actual, _ := ExtractRepoURLFromRemote(t.remoteURL)
			require.Equal(suite.Suite.T(), t.want, actual)
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
		suite.Suite.Run(t.name, func() {
			actual, err := removeUsernamePasswordFromURL(t.inputURL)
			require.Equal(suite.Suite.T(), t.wantError, err != nil)
			require.Equal(suite.Suite.T(), t.want, actual)
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
		suite.Suite.Run(t.name, func() {
			actual := getCommitURL(t.repoURL, t.commitHash)
			require.Equal(suite.Suite.T(), t.want, actual)
		})
	}
}

func (suite *GitViewTestSuite) TestGetCommitInfoFromCommitSHA() {
	dirPath := filepath.Join(suite.tmpDir, "repoName")
	_, worktree, err := initializeRepoAndCommit(dirPath, 1)
	require.NoError(suite.Suite.T(), err)

	gv, err := New(worktree.Filesystem.Root())
	require.NoError(suite.Suite.T(), err)

	_, err = gv.GetCommitInfoFromCommitSHA("58a9461c5a42d83bd5731485a72ddae542ac99d8", true, []string{})
	require.Error(suite.Suite.T(), err)
	expected := "failed to resolve git reference 58a9461c5a42d83bd5731485a72ddae542ac99d8: reference not found"
	require.Equal(suite.Suite.T(), expected, err.Error())

	_, err = gv.GetCommitInfoFromCommitSHA("HEAD~2", true, []string{})
	require.Error(suite.Suite.T(), err)
	expected = "failed to resolve git reference HEAD~2: EOF"
	require.Equal(suite.Suite.T(), expected, err.Error())

	commitInfo, err := gv.GetCommitInfoFromCommitSHA("HEAD", false, []string{})
	require.NoError(suite.Suite.T(), err)
	require.Equal(suite.Suite.T(), "Added file 1", commitInfo.Message)
	require.Equal(suite.Suite.T(), "master", commitInfo.Branch)
	require.Empty(suite.Suite.T(), commitInfo.Parents)
	require.Empty(suite.Suite.T(), commitInfo.URL)

	commitInfo, err = gv.GetCommitInfoFromCommitSHA("HEAD", false, []string{"author", "message", "branch"})
	require.NoError(suite.Suite.T(), err)
	require.Equal(suite.Suite.T(), redactedCommitInfoValue, commitInfo.Author)
	require.Equal(suite.Suite.T(), redactedCommitInfoValue, commitInfo.Message)
	require.Equal(suite.Suite.T(), redactedCommitInfoValue, commitInfo.Branch)
}

func (suite *GitViewTestSuite) TestMatchPatternInCommitMessageORBranchName() {
	_, workTree, fs, err := testHelpers.InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.Suite.T(), err)

	for _, t := range []struct {
		name            string
		pattern         string
		commitMessage   string
		secondarySource string
		wantError       bool
		want            []string
		commitSha       string
		branchName      string
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
			name:          "Jira references found in branch name",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "some test commit",
			branchName:    "EX-5-cool-branch",
			want:          []string{"EX-5"},
			wantError:     false,
		},
		{
			name:            "Jira references found in secondary source",
			pattern:         "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage:   "some test commit",
			secondarySource: "EX-1-test-commit",
			want:            []string{"EX-1"},
			wantError:       false,
		},
		{
			name:            "Jira references found in commit and secondary source",
			pattern:         "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage:   "EX-1 some test commit",
			secondarySource: "EX-2-test-commit",
			want:            []string{"EX-1", "EX-2"},
			wantError:       false,
		},
		{
			name:          "Jira references found in commit and branch name",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "EX-1 some test commit",
			branchName:    "EX-2-test-commit",
			want:          []string{"EX-1", "EX-2"},
			wantError:     false,
		},
		{
			name:          "Same Jira references found in commit and branch name is not duplicated",
			pattern:       "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage: "DUP-1 some test commit",
			branchName:    "DUP-1-test-commit",
			want:          []string{"DUP-1"},
			wantError:     false,
		},
		{
			name:            "Jira references found in commit, branch name and secondary source",
			pattern:         "[A-Z][A-Z0-9]{1,9}-[0-9]+",
			commitMessage:   "ALL-1 some test commit",
			branchName:      "ALL-2-test-commit",
			secondarySource: "ALL-3-some-things",
			want:            []string{"ALL-1", "ALL-2", "ALL-3"},
			wantError:       false,
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
		suite.Suite.Run(t.name, func() {

			if t.commitSha == "" {
				t.commitSha, err = testHelpers.CommitToRepo(workTree, fs, t.commitMessage)
				require.NoError(suite.Suite.T(), err)
			}

			if t.branchName != "" {
				err := testHelpers.CheckoutNewBranch(workTree, t.branchName)
				require.NoError(suite.Suite.T(), err)
				defer testHelpers.CheckoutMaster(workTree, suite.Suite.T())
			}

			gitView, err := New(suite.tmpDir)
			require.NoError(suite.Suite.T(), err)

			actual, _, err := gitView.MatchPatternInCommitMessageORBranchName(t.pattern, t.commitSha, t.secondarySource)
			require.True(suite.Suite.T(), (err != nil) == t.wantError)
			require.ElementsMatch(suite.Suite.T(), t.want, actual)

		})
	}
}

func (suite *GitViewTestSuite) TestResolveRevision() {
	_, workTree, fs, err := testHelpers.InitializeGitRepo(suite.tmpDir)
	require.NoError(suite.Suite.T(), err)

	FirstCommitSha, err := testHelpers.CommitToRepo(workTree, fs, "Test commit message 1")
	require.NoError(suite.Suite.T(), err)

	SecondCommitSha, err := testHelpers.CommitToRepo(workTree, fs, "Test commit message 2")
	require.NoError(suite.Suite.T(), err)

	ThirdCommitSha, err := testHelpers.CommitToRepo(workTree, fs, "Test commit message 3")
	require.NoError(suite.Suite.T(), err)

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
		suite.Suite.Run(t.name, func() {

			gitView, err := New(suite.tmpDir)
			require.NoError(suite.Suite.T(), err)

			actual, err := gitView.ResolveRevision(t.commitSHAOrRef)
			require.True(suite.Suite.T(), (err != nil) == t.wantError)
			require.Equal(suite.Suite.T(), t.want, actual)

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
