package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const auditLogsResponse = `{"page":1,"per_page":15,"total_pages":1,"total_count":2,"events":[` +
	`{"id":"evt-1","user":"Alice Example","changed_by":"Bob Admin","role":"admin","previous_role":"member","source":"app","created_at":1780584129.6878593},` +
	`{"id":"evt-2","user":"Carol New","changed_by":"External IdP","role":"member","previous_role":"","source":"scim","created_at":1780584130.5}]}`

const emptyAuditLogsResponse = `{"page":1,"per_page":15,"total_pages":0,"total_count":0,"events":[]}`

func TestPrintAuditLogsListAsTable(t *testing.T) {
	var buf bytes.Buffer
	err := printAuditLogsListAsTable(auditLogsResponse, &buf, 1)
	require.NoError(t, err)

	out := buf.String()
	for _, want := range []string{
		"WHEN", "USER", "ROLE", "PREVIOUS ROLE", "CHANGED BY", "SOURCE",
		"Alice Example", "Bob Admin", "Carol New", "External IdP", "scim",
		"Showing page 1 of 1, total 2 events",
	} {
		require.Contains(t, out, want)
	}
	require.Regexp(t, `Alice Example\s+admin\s+member\s+Bob Admin\s+app`, out)
	require.Regexp(t, `Carol New\s+member\s+External IdP\s+scim`, out)
}

func TestPrintAuditLogsListAsTableWithoutCreatedAt(t *testing.T) {
	// the server leaves created_at out of an event that has none
	raw := `{"page":1,"per_page":15,"total_pages":1,"total_count":1,"events":[` +
		`{"id":"evt-3","user":"Dave Old","changed_by":"Bob Admin","role":"member","previous_role":"","source":"script"}]}`

	var buf bytes.Buffer
	err := printAuditLogsListAsTable(raw, &buf, 1)
	require.NoError(t, err)
	require.Regexp(t, `N/A\s+Dave Old\s+member\s+Bob Admin\s+script`, buf.String())
}

type ListAuditLogsCommandTestSuite struct {
	suite.Suite
	fake                  *httpfake.HTTPFake
	defaultKosliArguments string
}

func (suite *ListAuditLogsCommandTestSuite) SetupTest() {
	suite.fake = httpfake.New(httpfake.WithTesting(suite.T()))
	global = &GlobalOpts{
		ApiToken: "secret",
		Org:      "test-org",
		Host:     suite.fake.Server.URL,
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ListAuditLogsCommandTestSuite) TearDownTest() {
	suite.fake.Close()
}

func (suite *ListAuditLogsCommandTestSuite) replyWith(body string) *httpfake.Request {
	request := suite.fake.NewHandler().Get("/api/v2/membership-audit-log/test-org")
	request.Reply(200).BodyString(body)
	return request
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsCmd() {
	suite.replyWith(auditLogsResponse).
		AssertQueryValue("page", "1").
		AssertQueryValue("per_page", "15")

	tests := []cmdTestCase{
		{
			name:        "listing audit logs prints a table",
			cmd:         "list audit-logs" + suite.defaultKosliArguments,
			goldenRegex: `(?s)WHEN\s+USER\s+ROLE\s+PREVIOUS ROLE\s+CHANGED BY\s+SOURCE.*Alice Example.*Showing page 1 of 1, total 2 events`,
		},
		{
			name:       "listing audit logs with --output json prints the response",
			cmd:        "list audit-logs --output json" + suite.defaultKosliArguments,
			goldenJson: []jsonCheck{{"events", "length:2"}, {"events.[0].id", "evt-1"}, {"total_count", float64(2)}},
		},
		{
			name:        "the audit-log alias works",
			cmd:         "list audit-log" + suite.defaultKosliArguments,
			goldenRegex: `Alice Example`,
		},
		{
			name:        "the al alias works",
			cmd:         "list al" + suite.defaultKosliArguments,
			goldenRegex: `Alice Example`,
		},
		{
			name:        "the als alias works",
			cmd:         "list als" + suite.defaultKosliArguments,
			goldenRegex: `Alice Example`,
		},
		{
			wantError: true,
			name:      "providing an argument causes an error",
			cmd:       "list audit-logs xxx" + suite.defaultKosliArguments,
			golden:    "Error: unknown command \"xxx\" for \"kosli list audit-logs\"\n",
		},
		{
			wantError:   true,
			name:        "--page must be a positive integer",
			cmd:         "list audit-logs --page 0" + suite.defaultKosliArguments,
			goldenRegex: "^Error: page number must be a positive integer",
		},
		{
			wantError:   true,
			name:        "--page-limit must be a positive integer",
			cmd:         "list audit-logs --page-limit 0" + suite.defaultKosliArguments,
			goldenRegex: "^Error: page limit must be a positive integer",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsSendsPaginationParams() {
	suite.replyWith(emptyAuditLogsResponse).
		AssertQueryValue("page", "3").
		AssertQueryValue("per_page", "5")

	tests := []cmdTestCase{
		{
			name:   "--page and --page-limit are sent and an empty page names its number",
			cmd:    "list audit-logs --page 3 --page-limit 5" + suite.defaultKosliArguments,
			golden: "No audit log events were found at page number 3.\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsWhenThereAreNone() {
	suite.replyWith(emptyAuditLogsResponse)

	tests := []cmdTestCase{
		{
			name:   "an empty audit log reports that no events were found",
			cmd:    "list audit-logs" + suite.defaultKosliArguments,
			golden: "No audit log events were found.\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsSendsFilterAndSortParams() {
	suite.replyWith(auditLogsResponse).
		AssertQueryValue("search", "alice").
		AssertQueryValue("event_type", "changed").
		AssertQueryValue("sort_by", "user").
		AssertQueryValue("sort_dir", "asc").
		AssertCustom(httpfake.CustomAssertor(func(r *http.Request) error {
			query := r.URL.Query()
			if got := query["role"]; !slices.Equal(got, []string{"admin", "member"}) {
				return fmt.Errorf("role query values: got %v", got)
			}
			if got := query["source"]; !slices.Equal(got, []string{"app", "scim"}) {
				return fmt.Errorf("source query values: got %v", got)
			}
			return nil
		}))

	tests := []cmdTestCase{
		{
			name: "filter and sort flags are sent as query params",
			cmd: "list audit-logs --search alice --event-type changed --role admin --role member" +
				" --source app --source scim --sort-by user --sort-direction asc" + suite.defaultKosliArguments,
			goldenRegex: `Alice Example`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsOmitsUnsetFilters() {
	suite.replyWith(auditLogsResponse).
		AssertCustom(httpfake.CustomAssertor(func(r *http.Request) error {
			for _, key := range []string{"search", "event_type", "role", "source", "sort_by", "sort_dir"} {
				if r.URL.Query().Has(key) {
					return fmt.Errorf("query param %q should not be sent when its flag is unset", key)
				}
			}
			return nil
		}))

	tests := []cmdTestCase{
		{
			name:        "unset filter and sort flags are left out of the query",
			cmd:         "list audit-logs" + suite.defaultKosliArguments,
			goldenRegex: `Alice Example`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsRejectsInvalidFlagValues() {
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "--event-type rejects an unknown value",
			cmd:         "list audit-logs --event-type deleted" + suite.defaultKosliArguments,
			goldenRegex: `^Error: invalid --event-type "deleted": valid values are \[added, changed, removed\]`,
		},
		{
			wantError:   true,
			name:        "--sort-by rejects an unknown column",
			cmd:         "list audit-logs --sort-by id" + suite.defaultKosliArguments,
			goldenRegex: `^Error: invalid --sort-by "id": valid values are \[when, user, role, changed_by, source\]`,
		},
		{
			wantError:   true,
			name:        "--sort-direction rejects an unknown value",
			cmd:         "list audit-logs --sort-direction up" + suite.defaultKosliArguments,
			goldenRegex: `^Error: invalid --sort-direction "up": valid values are \[asc, desc\]`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ListAuditLogsCommandTestSuite) TestListAuditLogsCmdIsRegisteredAndNotBeta() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	cmd, _, err := root.Find([]string{"list", "audit-logs"})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "audit-logs", cmd.Name())
	require.NotContains(suite.T(), cmd.Annotations, betaCLIAnnotation)
}

func TestListAuditLogsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListAuditLogsCommandTestSuite))
}
