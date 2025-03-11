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
	// If this test fails, a simple way to retrieve a new generated master is to:
	// - add an import for fmt
	// - uncomment the fmt.Printf() call below
	// - comment out the line defer os.RemoveAll(tempDirName)
	// Then:
	// - make test_integration_single TARGET=TestDocsCommandTestSuite
	// will tell you where the new snyk.md master file lives.
	// Then copy it to ./cmd/kosli/testdata/output/docs/
	// and undo the changes above.
	global = &GlobalOpts{}
	tempDirName, err := os.MkdirTemp("", "generatedDocs")
	//fmt.Printf("tempDirName :%s:\n\n\n\n\n", tempDirName)
	require.NoError(suite.Suite.T(), err)
	defer os.RemoveAll(tempDirName)

	o := &docsOptions{
		dest:            tempDirName,
		topCmd:          newAttestSnykCmd(os.Stdout),
		generateHeaders: true,
	}
	err = o.run()
	require.NoError(suite.Suite.T(), err)

	actualFile := filepath.Join(tempDirName, "snyk.md")
	require.FileExists(suite.Suite.T(), actualFile)
	err = compareTwoFiles(actualFile, goldenPath("output/docs/snyk.md"))
	require.NoError(suite.Suite.T(), err)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestDocsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DocsCommandTestSuite))
}
