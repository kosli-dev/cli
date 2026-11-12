package main

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	productUpdatesList = `{"unread_count":1,"product_updates":[` +
		`{"id":"rego-policies","title":"Rego policies","url":"https://docs.kosli.com/changelog#rego-policies",` +
		`"summary":"Store and version Rego policies in Kosli.","published_at":1757000000.5,"unread":true},` +
		`{"id":"controls","title":"Controls","url":"https://docs.kosli.com/changelog#controls",` +
		`"summary":"","published_at":1750000000.5,"unread":false}]}`
	productUpdatesEmptyList = `{"unread_count":0,"product_updates":[]}`
)

type ListProductUpdatesCommandTestSuite struct {
	suite.Suite
}

func (suite *ListProductUpdatesCommandTestSuite) SetupTest() {
	// Makes formattedTimestamp print a fixed UTC time.
	suite.T().Setenv("KOSLI_TESTS", "true")
}

func (suite *ListProductUpdatesCommandTestSuite) listCmd(host, args string) string {
	return fmt.Sprintf("list product-updates %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItIsVisibleAndBeta() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	cmd, _, err := root.Find([]string{"list", "product-updates"})
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), "product-updates", cmd.Name())
	require.False(suite.T(), cmd.Hidden)
	require.Contains(suite.T(), cmd.Annotations, betaCLIAnnotation)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItHasShortAliases() {
	for _, alias := range []string{"pu", "pr-up"} {
		server, fake := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"list %s --host %s --org test-org --api-token test-token --max-api-retries 0", alias, server.URL))

		require.NoError(suite.T(), err, alias)
		require.Equal(suite.T(), "/api/v2/product-updates/test-org", fake.escapedPath, alias)
	}
}

func (suite *ListProductUpdatesCommandTestSuite) TestItNeedsTheOrgAndApiToken() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)
	noConfig := withoutOrgAndApiToken(suite.T())

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"list product-updates --api-token test-token --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --org is not set\n")

	_, combined, _, _, err = executeCommandC(fmt.Sprintf(
		"list product-updates --org test-org --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --api-token is not set\n")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItTakesNoArguments() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "rego-policies"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `unknown command "rego-policies"`)
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItReadsTheOrgsProductUpdates() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, _, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.MethodGet, fake.method)
	require.Equal(suite.T(), "/api/v2/product-updates/test-org", fake.escapedPath)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItPrintsATable() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(),
		"ID             TITLE          PUBLISHED                      UNREAD  URL\n"+
			"rego-policies  Rego policies  Sat, 16 Jan 2016 00:00:00 UTC  true    https://docs.kosli.com/changelog#rego-policies\n"+
			"controls       Controls       Sat, 16 Jan 2016 00:00:00 UTC  false   https://docs.kosli.com/changelog#controls\n"+
			"\n"+
			"1 unread of 2 product updates\n",
		combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItPrintsNAForAMissingPublishedTime() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK,
		`{"unread_count":0,"product_updates":[{"id":"controls","title":"Controls",`+
			`"url":"https://docs.kosli.com/changelog#controls","summary":"","unread":false}]}`)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "controls  Controls  N/A        false")
}

func (suite *ListProductUpdatesCommandTestSuite) TestItSaysWhenThereAreNoProductUpdates() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesEmptyList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "No product updates were found.\n", combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItPrintsTheServerAnswerAsJSON() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), productUpdatesList, combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItReturnsAServerError() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusForbidden,
		`{"message":"Product updates is a beta feature that is not enabled for organization 'test-org'"}`)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "not enabled for organization 'test-org'")
}

func TestListProductUpdatesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListProductUpdatesCommandTestSuite))
}
