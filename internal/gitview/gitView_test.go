package gitview

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
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
	var err error
	suite.tmpDir, err = os.MkdirTemp("", "testRepoDir")
	require.NoError(suite.T(), err, "error creating a temporary test directory")
}

func (suite *GitViewTestSuite) TestCommitsBetween() {
	for _, t := range []struct {
		name          string
		repoName      string
		commitsNumber int
		want          int
		expectError   bool
	}{
		{
			name:          "accurately counts 'commits between' as 1 for newest == oldest",
			repoName:      "oneCommit",
			commitsNumber: 1,
			want:          1,
		},
		{
			name:          "accurately counts 'commits between' as 3",
			repoName:      "threeCommits",
			commitsNumber: 3,
			want:          3,
		},
	} {
		suite.Run(t.name, func() {
			dirPath := filepath.Join(suite.tmpDir, t.repoName)
			err := os.Mkdir(dirPath, 0777)
			require.NoErrorf(suite.T(), err, "error creating test dir %s", t.repoName)
			r, err := git.Init(memory.NewStorage(), nil)
			require.NoErrorf(suite.T(), err, "error creating test repository %s", t.repoName)

		})
	}
}

func TestGitViewTestSuite(t *testing.T) {
	suite.Run(t, new(GitViewTestSuite))
}

// accurately counts "commits between" as 1 for newest == oldest
// accurately counts "commits between" as 3 ...
