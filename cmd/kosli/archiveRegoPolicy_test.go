package main

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const archivedRegoPolicy = `{"identifier":"pr-approved","name":"Pull request approved","archived":true}`

// ArchiveRegoPolicyCommandTestSuite runs each case for both archive and
// unarchive, which differ only in the endpoint they post to.
type ArchiveRegoPolicyCommandTestSuite struct {
	suite.Suite
}

var regoPolicyArchiveVerbs = []string{"archive", "unarchive"}

func (suite *ArchiveRegoPolicyCommandTestSuite) verbCmd(verb, host, args string) string {
	return fmt.Sprintf("%s rego-policy %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		verb, args, host)
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItIsVisibleAndBeta() {
	for _, verb := range regoPolicyArchiveVerbs {
		root, err := newRootCmd(io.Discard, io.Discard, nil)
		require.NoError(suite.T(), err)
		cmd, _, err := root.Find([]string{verb, "rego-policy"})
		require.NoError(suite.T(), err)

		require.Equal(suite.T(), "rego-policy", cmd.Name(), verb)
		require.False(suite.T(), cmd.Hidden, verb)
		require.Contains(suite.T(), cmd.Annotations, betaCLIAnnotation, verb)
	}
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItNeedsTheOrgAndApiToken() {
	for _, verb := range regoPolicyArchiveVerbs {
		server, fake := newFakeRegoPolicyRead(suite.T(), archivedRegoPolicy)
		noConfig := withoutOrgAndApiToken(suite.T())

		_, combined, _, _, err := executeCommandC(fmt.Sprintf(
			"%s rego-policy pr-approved --api-token test-token --host %s %s", verb, server.URL, noConfig))
		require.Error(suite.T(), err, verb)
		require.Contains(suite.T(), combined, "Error: --org is not set\n", verb)

		_, combined, _, _, err = executeCommandC(fmt.Sprintf(
			"%s rego-policy pr-approved --org test-org --host %s %s", verb, server.URL, noConfig))
		require.Error(suite.T(), err, verb)
		require.Contains(suite.T(), combined, "Error: --api-token is not set\n", verb)
		require.Equal(suite.T(), 0, fake.calls, verb)
	}
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItTakesExactlyOneIdentifier() {
	for _, verb := range regoPolicyArchiveVerbs {
		for _, args := range []string{"", "pr-approved release-ready"} {
			server, fake := newFakeRegoPolicyRead(suite.T(), archivedRegoPolicy)

			_, combined, _, _, err := executeCommandC(suite.verbCmd(verb, server.URL, args))

			require.Error(suite.T(), err, verb+" "+args)
			require.Contains(suite.T(), combined, "accepts 1 arg(s)", verb+" "+args)
			require.Equal(suite.T(), 0, fake.calls, verb+" "+args)
		}
	}
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItPostsToThePolicyEndpointOfTheVerb() {
	for _, verb := range regoPolicyArchiveVerbs {
		server, fake := newFakeRegoPolicyRead(suite.T(), archivedRegoPolicy)

		_, combined, _, _, err := executeCommandC(suite.verbCmd(verb, server.URL, "pr-approved"))

		require.NoError(suite.T(), err, verb)
		require.Equal(suite.T(), http.MethodPost, fake.method, verb)
		require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved/"+verb, fake.escapedPath, verb)
		require.Equal(suite.T(), fmt.Sprintf("rego policy pr-approved was %sd\n", verb), combined, verb)
	}
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItKeepsTheIdentifierInOneSegment() {
	for _, verb := range regoPolicyArchiveVerbs {
		server, fake := newFakeRegoPolicyRead(suite.T(), archivedRegoPolicy)

		_, _, _, _, err := executeCommandC(suite.verbCmd(verb, server.URL, "a/b"))

		require.NoError(suite.T(), err, verb)
		require.Equal(suite.T(), "/api/v2/rego-policies/test-org/a%2Fb/"+verb, fake.escapedPath, verb)
	}
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItSendsNothingOnADryRun() {
	for _, verb := range regoPolicyArchiveVerbs {
		server, fake := newFakeRegoPolicyRead(suite.T(), archivedRegoPolicy)

		_, combined, _, _, err := executeCommandC(suite.verbCmd(verb, server.URL, "pr-approved --dry-run"))

		require.NoError(suite.T(), err, verb)
		require.Equal(suite.T(), 0, fake.calls, verb)
		require.NotContains(suite.T(), combined, "was "+verb+"d", verb)
	}
}

func (suite *ArchiveRegoPolicyCommandTestSuite) TestItReturnsAServerError() {
	for _, verb := range regoPolicyArchiveVerbs {
		server, fake := newFakeRegoPolicyRead(suite.T(), `{"message":"Rego policy 'no-such-policy' not found"}`)
		fake.status = http.StatusNotFound

		_, combined, _, _, err := executeCommandC(suite.verbCmd(verb, server.URL, "no-such-policy"))

		require.Error(suite.T(), err, verb)
		require.Contains(suite.T(), combined, "Rego policy 'no-such-policy' not found", verb)
		require.NotContains(suite.T(), combined, "was "+verb+"d", verb)
	}
}

func TestArchiveRegoPolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ArchiveRegoPolicyCommandTestSuite))
}
