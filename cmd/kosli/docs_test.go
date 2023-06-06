package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type DocsCommandTestSuite struct {
	suite.Suite
}

func (suite *DocsCommandTestSuite) TestDocsCmd() {
	global = &GlobalOpts{}
	tempDirName, err := os.MkdirTemp("", "generatedDocs")
	if err != nil {
		suite.T().Fail()
	}
	defer os.RemoveAll(tempDirName)
	o := &docsOptions{
		dest:            tempDirName,
		topCmd:          newReportArtifactCmd(os.Stdout),
		generateHeaders: true,
	}
	o.run()
	actualFile := filepath.Join(tempDirName, "artifact.md")
	require.FileExists(suite.T(), actualFile)
	compareTwoFile(actualFile, goldenPath("output/docs/artifact.md"))
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDocsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DocsCommandTestSuite))
}
