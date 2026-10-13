package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	regoPoliciesPage = `{"page":1,"per_page":15,"total_pages":1,"total_count":2,"rego_policies":[` +
		`{"identifier":"pr-approved","name":"Pull request approved","description":"Approved PRs only",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","updated_at":1757000100.5},` +
		`{"identifier":"release-ready","name":"Release ready",` +
		`"created_at":1757000200.5,"created_by":"John Roe","updated_at":1757000300.5}]}`
	regoPoliciesEmptyPage = `{"page":%d,"per_page":15,"total_pages":0,"total_count":0,"rego_policies":[]}`
)

type ListRegoPoliciesCommandTestSuite struct {
	suite.Suite
}

func (suite *ListRegoPoliciesCommandTestSuite) SetupTest() {
	// Makes formattedTimestamp print a fixed UTC time.
	suite.T().Setenv("KOSLI_TESTS", "true")
}

func (suite *ListRegoPoliciesCommandTestSuite) listCmd(host, args string) string {
	return fmt.Sprintf("list rego-policies %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItIsHiddenAndBeta() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	cmd, _, err := root.Find([]string{"list", "rego-policies"})
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), "rego-policies", cmd.Name())
	require.True(suite.T(), cmd.Hidden)
	require.Contains(suite.T(), cmd.Annotations, betaCLIAnnotation)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItNeedsTheOrgAndApiToken() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)
	noConfig := withoutOrgAndApiToken(suite.T())

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"list rego-policies --api-token test-token --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --org is not set\n")

	_, combined, _, _, err = executeCommandC(fmt.Sprintf(
		"list rego-policies --org test-org --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --api-token is not set\n")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItTakesNoArguments() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "pr-approved"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `unknown command "pr-approved"`)
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItReadsTheFirstPageOf15ByDefault() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)

	_, _, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "/api/v2/rego-policies/test-org", fake.escapedPath)
	require.Equal(suite.T(), url.Values{"page": {"1"}, "per_page": {"15"}}, fake.query)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItPrintsATable() {
	server, _ := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(),
		"IDENTIFIER     NAME                   DESCRIPTION        CREATED BY  LAST UPDATED\n"+
			"pr-approved    Pull request approved  Approved PRs only  Jane Doe    Sat, 16 Jan 2016 00:00:00 UTC\n"+
			"release-ready  Release ready                             John Roe    Sat, 16 Jan 2016 00:00:00 UTC\n"+
			"\n"+
			"Showing page 1 of 1, total 2 rego policies\n",
		combined)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItSaysWhenThereAreNoPolicies() {
	server, _ := newFakeRegoPolicyRead(suite.T(), fmt.Sprintf(regoPoliciesEmptyPage, 1))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "No rego policies were found.\n", combined)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItNamesThePageWhenAPageAfterTheFirstIsEmpty() {
	server, _ := newFakeRegoPolicyRead(suite.T(), fmt.Sprintf(regoPoliciesEmptyPage, 3))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--page 3"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "No rego policies were found at page number 3.\n", combined)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItPrintsTheServerAnswerAsJSON() {
	server, _ := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), regoPoliciesPage, combined)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItSendsThePagingSearchAndSort() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)

	_, _, _, _, err := executeCommandC(suite.listCmd(server.URL,
		"--page 2 --page-limit 50 --search approved --sort identifier --sort-direction asc"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), url.Values{
		"page":           {"2"},
		"per_page":       {"50"},
		"search":         {"approved"},
		"sort_by":        {"identifier"},
		"sort_direction": {"asc"},
	}, fake.query)
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItRefusesAPageOrPageLimitThatIsNotPositive() {
	for _, args := range []string{"--page 0", "--page-limit 0"} {
		server, fake := newFakeRegoPolicyRead(suite.T(), regoPoliciesPage)

		_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, args))

		require.Error(suite.T(), err, args)
		require.Contains(suite.T(), combined, "must be a positive integer", args)
		require.Equal(suite.T(), 0, fake.calls, args)
	}
}

func (suite *ListRegoPoliciesCommandTestSuite) TestItReturnsAServerError() {
	server, fake := newFakeRegoPolicyRead(suite.T(),
		`{"message":"User 'jane' does not have access to organization 'test-org'"}`)
	fake.status = http.StatusForbidden

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "does not have access to organization 'test-org'")
}

func TestListRegoPoliciesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListRegoPoliciesCommandTestSuite))
}
