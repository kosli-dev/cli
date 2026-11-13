package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
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

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all"))

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

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "controls  Controls  N/A        false")
}

// manyProductUpdates is a list of count updates, u0 first, with the first unread of them unread.
func manyProductUpdates(count, unread int) string {
	updates := make([]string, count)
	for i := range updates {
		updates[i] = fmt.Sprintf(`{"id":"u%d","title":"Update %d","url":"https://docs.kosli.com/changelog#u%d",`+
			`"summary":"","published_at":1757000000.5,"unread":%t}`, i, i, i, i < unread)
	}
	return fmt.Sprintf(`{"unread_count":%d,"product_updates":[%s]}`, unread, strings.Join(updates, ","))
}

func (suite *ListProductUpdatesCommandTestSuite) TestItListsOnlyUnreadUpdatesByDefault() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(),
		"ID             TITLE          PUBLISHED                      UNREAD  URL\n"+
			"rego-policies  Rego policies  Sat, 16 Jan 2016 00:00:00 UTC  true    https://docs.kosli.com/changelog#rego-policies\n"+
			"\n"+
			"1 unread product update(s)\n",
		combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestJSONHoldsOnlyUnreadUpdatesByDefault() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), `{"unread_count":1,"product_updates":[`+
		`{"id":"rego-policies","title":"Rego policies","url":"https://docs.kosli.com/changelog#rego-policies",`+
		`"summary":"Store and version Rego policies in Kosli.","published_at":1757000000.5,"unread":true}]}`, combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestJSONKeepsUnknownFieldsAndHTMLCharacters() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK,
		`{"unread_count":1,"feed_url":"https://docs.kosli.com/changelog","product_updates":[`+
			`{"id":"a","summary":"Rego & <OPA>","unread":true},{"id":"b","unread":false}]}`)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), `{"unread_count":1,"feed_url":"https://docs.kosli.com/changelog","product_updates":[`+
		`{"id":"a","summary":"Rego & <OPA>","unread":true}]}`, combined)
	require.Contains(suite.T(), combined, "Rego & <OPA>")
}

func (suite *ListProductUpdatesCommandTestSuite) TestItSaysWhenNothingIsUnread() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, manyProductUpdates(3, 0))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "No unread product updates. Use --all to list the ones you have seen.\n", combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestDismissMarksTheUnreadUpdatesListedAsSeen() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = productUpdatesList

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all --dismiss"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "/api/v2/product-updates/test-org/seen", fake.escapedPath)
	require.Len(suite.T(), fake.posted, 1)
	require.JSONEq(suite.T(), `{"ids":["rego-policies"]}`, fake.posted[0])
	require.Contains(suite.T(), combined, "1 unread of 2 product updates\n1 product update(s) dismissed\n")
}

func (suite *ListProductUpdatesCommandTestSuite) TestDismissLeavesUpdatesBeyondTheLimitUnread() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = manyProductUpdates(12, 11)

	_, _, _, _, err := executeCommandC(suite.listCmd(server.URL, "--dismiss"))

	require.NoError(suite.T(), err)
	require.Len(suite.T(), fake.posted, 1)
	require.Equal(suite.T(), 10, strings.Count(fake.posted[0], `"u`))
	require.NotContains(suite.T(), fake.posted[0], `"u10"`)
}

func (suite *ListProductUpdatesCommandTestSuite) TestDismissSendsNothingWhenNothingIsUnread() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = manyProductUpdates(3, 0)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--dismiss"))

	require.NoError(suite.T(), err)
	require.Empty(suite.T(), fake.posted)
	require.NotContains(suite.T(), combined, "dismissed")
}

func (suite *ListProductUpdatesCommandTestSuite) TestDismissKeepsJSONOutputClean() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = manyProductUpdates(2, 2)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--dismiss --output json"))

	require.NoError(suite.T(), err)
	require.Len(suite.T(), fake.posted, 1)
	require.JSONEq(suite.T(), manyProductUpdates(2, 2), combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestAFailedDismissIsAnError() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = manyProductUpdates(2, 2)
	fake.failOnPost = 1

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--dismiss"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "boom")
}

func (suite *ListProductUpdatesCommandTestSuite) TestItListsTheFirst10ByDefault() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, manyProductUpdates(12, 11))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, ""))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "\nu9 ")
	require.NotContains(suite.T(), combined, "\nu10 ")
	require.Contains(suite.T(), combined,
		"\nShowing 10 of 11 unread product updates. Use --limit to list more.\n")
}

func (suite *ListProductUpdatesCommandTestSuite) TestLimitSetsHowManyAreListed() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, manyProductUpdates(12, 1))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all --limit 2"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "\nu1 ")
	require.NotContains(suite.T(), combined, "\nu2 ")
	require.Contains(suite.T(), combined, "\nShowing 2 of 12 product updates, 1 unread in total.")
}

func (suite *ListProductUpdatesCommandTestSuite) TestLimit0ListsAll() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, manyProductUpdates(12, 1))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all --limit 0"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "\nu11 ")
	require.Contains(suite.T(), combined, "\n1 unread of 12 product updates\n")
}

func (suite *ListProductUpdatesCommandTestSuite) TestLimitAlsoAppliesToJSON() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, manyProductUpdates(12, 11))

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--limit 3 --output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), strings.Replace(manyProductUpdates(3, 3), `"unread_count":3`, `"unread_count":11`, 1), combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItRefusesANegativeLimit() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--limit -1"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "illegal")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItSaysWhenThereAreNoProductUpdates() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesEmptyList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "No product updates were found.\n", combined)
}

func (suite *ListProductUpdatesCommandTestSuite) TestItPrintsTheServerAnswerAsJSON() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusOK, productUpdatesList)

	_, combined, _, _, err := executeCommandC(suite.listCmd(server.URL, "--all --output json"))

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

// The root command already refuses a flag value starting with "-"; this pins
// the command's own rule, which must hold without that guard.
func TestListProductUpdatesRefusesANegativeLimitItself(t *testing.T) {
	o := &listProductUpdatesOptions{limit: -1}

	err := o.validate(newListProductUpdatesCmd(io.Discard))

	require.ErrorContains(t, err, "--limit must be 0 or more")
}

func TestListProductUpdatesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListProductUpdatesCommandTestSuite))
}
