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

type DismissProductUpdatesCommandTestSuite struct {
	suite.Suite
}

func (suite *DismissProductUpdatesCommandTestSuite) dismissCmd(host, args string) string {
	return fmt.Sprintf("dismiss product-updates %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItIsVisibleAndBeta() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	cmd, _, err := root.Find([]string{"dismiss", "product-updates"})
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), "product-updates", cmd.Name())
	require.False(suite.T(), cmd.Hidden)
	require.Contains(suite.T(), cmd.Annotations, betaCLIAnnotation)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItHasShortAliases() {
	for _, alias := range []string{"dism pu", "dism pr-up", "dismiss pu", "dm pu"} {
		server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

		_, _, _, _, err := executeCommandC(fmt.Sprintf(
			"%s controls --host %s --org test-org --api-token test-token --max-api-retries 0", alias, server.URL))

		require.NoError(suite.T(), err, alias)
		require.Equal(suite.T(), "/api/v2/product-updates/test-org/seen", fake.escapedPath, alias)
	}
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItNeedsTheOrgAndApiToken() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	noConfig := withoutOrgAndApiToken(suite.T())

	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"dismiss product-updates controls --api-token test-token --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --org is not set\n")

	_, combined, _, _, err = executeCommandC(fmt.Sprintf(
		"dismiss product-updates controls --org test-org --host %s %s", server.URL, noConfig))
	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --api-token is not set\n")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItNeedsAtLeastOneIDOrAll() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, ""))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: provide at least one PRODUCT-UPDATE-ID, or --all\n")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItRefusesIDsTogetherWithAll() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "controls --all"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: --all cannot be combined with PRODUCT-UPDATE-ID arguments\n")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestAllDismissesEveryUnreadUpdate() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = productUpdatesList

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "--all"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "/api/v2/product-updates/test-org/seen", fake.escapedPath)
	require.Len(suite.T(), fake.posted, 1)
	require.JSONEq(suite.T(), `{"ids":["rego-policies"]}`, fake.posted[0])
	require.Equal(suite.T(), "1 product update(s) dismissed\n", combined)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestAllSendsNothingWhenNothingIsUnread() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = productUpdatesEmptyList

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "--all"))

	require.NoError(suite.T(), err)
	require.Empty(suite.T(), fake.posted)
	require.Equal(suite.T(), "No unread product updates to dismiss.\n", combined)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestAllSendsAtMost200IDsPerRequest() {
	updates := make([]string, 201)
	for i := range updates {
		updates[i] = fmt.Sprintf(`{"id":"update-%d","unread":true}`, i)
	}
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = `{"unread_count":201,"product_updates":[` + strings.Join(updates, ",") + `]}`

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "--all"))

	require.NoError(suite.T(), err)
	require.Len(suite.T(), fake.posted, 2)
	require.Equal(suite.T(), 200, strings.Count(fake.posted[0], "update-"))
	require.JSONEq(suite.T(), `{"ids":["update-200"]}`, fake.posted[1])
	require.Equal(suite.T(), "201 product update(s) dismissed\n", combined)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestAFailedLaterBatchSaysHowManyWereDismissed() {
	updates := make([]string, 201)
	for i := range updates {
		updates[i] = fmt.Sprintf(`{"id":"update-%d","unread":true}`, i)
	}
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = `{"unread_count":201,"product_updates":[` + strings.Join(updates, ",") + `]}`
	fake.failOnPost = 2

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "--all"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "dismissed 200 of 201 product update(s) before failing: boom")
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItSendsEachIDOnceAndSkipsEmptyOnes() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, `controls "" controls rego-policies`))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), `{"ids":["controls","rego-policies"]}`, fake.requestBody)
	require.Equal(suite.T(), "2 product update(s) dismissed\n", combined)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItRefusesOnlyEmptyIDs() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, `""`))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Error: provide at least one PRODUCT-UPDATE-ID, or --all\n")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestAllWithADryRunSendsNothing() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")
	fake.listBody = productUpdatesList

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "--all --dry-run"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, fake.calls)
	require.Contains(suite.T(), combined, "every unread product update would be dismissed\n")
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItSendsTheIDs() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, _, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "rego-policies controls"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), http.MethodPost, fake.method)
	require.Equal(suite.T(), "/api/v2/product-updates/test-org/seen", fake.escapedPath)
	require.JSONEq(suite.T(), `{"ids":["rego-policies","controls"]}`, fake.requestBody)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItReportsHowManyWereDismissed() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "rego-policies controls"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "2 product update(s) dismissed\n", combined)
}

func (suite *DismissProductUpdatesCommandTestSuite) TestADryRunSendsNothing() {
	server, fake := newFakeProductUpdates(suite.T(), http.StatusCreated, "null")

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "controls --dry-run"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, fake.calls)
	require.NotContains(suite.T(), combined, "dismissed")
}

func (suite *DismissProductUpdatesCommandTestSuite) TestItReturnsAServerError() {
	server, _ := newFakeProductUpdates(suite.T(), http.StatusForbidden,
		`{"message":"Product updates is a beta feature that is not enabled for organization 'test-org'"}`)

	_, combined, _, _, err := executeCommandC(suite.dismissCmd(server.URL, "controls"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "not enabled for organization 'test-org'")
	require.NotContains(suite.T(), combined, "dismissed")
}

func TestDismissProductUpdatesCommandTestSuite(t *testing.T) {
	suite.Run(t, new(DismissProductUpdatesCommandTestSuite))
}
