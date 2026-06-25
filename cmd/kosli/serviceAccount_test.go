package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// serviceAccountFixture reads a service-account response fixture from
// testdata/service-account. The fixtures hold the canonical API response
// bodies so the response contract lives in one place (see the README there).
func serviceAccountFixture(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "service-account", name))
	require.NoError(t, err, "failed to read fixture %s", name)
	return string(body)
}

func TestPrintServiceAccountAsTable(t *testing.T) {
	// Timestamps come back as floating-point epoch seconds.
	raw := `{"name":"ci-bot","description":"CI service account","privilege":"member","created_at":1780584129.6878593}`

	var buf bytes.Buffer
	err := printServiceAccountAsTable(raw, &buf, 0)
	require.NoError(t, err)

	out := buf.String()
	for _, want := range []string{"Name:", "ci-bot", "Description:", "CI service account", "Privilege:", "member", "Created At:"} {
		require.Contains(t, out, want)
	}
}

func TestPrintServiceAccountAsTableEmptyDescription(t *testing.T) {
	raw := `{"name":"ci-bot","description":"","privilege":"member","created_at":1780584129.5}`

	var buf bytes.Buffer
	err := printServiceAccountAsTable(raw, &buf, 0)
	require.NoError(t, err)
	require.Regexp(t, `Description:\s+N/A`, buf.String())
}

func TestPrintServiceAccountsListAsTable(t *testing.T) {
	raw := `[{"name":"ci-bot","description":"first","privilege":"member","created_at":1780584129.5},` +
		`{"name":"deployer","description":"","privilege":"admin","created_at":1780584130.5}]`

	var buf bytes.Buffer
	err := printServiceAccountsListAsTable(raw, &buf, 0)
	require.NoError(t, err)

	out := buf.String()
	for _, want := range []string{"NAME", "DESCRIPTION", "PRIVILEGE", "CREATED", "ci-bot", "first", "deployer", "admin"} {
		require.Contains(t, out, want)
	}
	// the deployer's empty description must render as N/A, not blank
	require.Contains(t, out, "N/A")
}

// Define the suite, and absorb the built-in basic suite functionality from testify.
type ServiceAccountCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *ServiceAccountCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ServiceAccountCommandTestSuite) TestCreateServiceAccountCmd() {
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "create builds the right url and payload (dry-run)",
			cmd:         "create service-account ci-bot --privilege member --description 'CI bot' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user.*"name": "ci-bot".*"privilege": "member"`,
		},
		{
			wantError:   false,
			name:        "the service-account alias (sa) works",
			cmd:         "create sa ci-bot --privilege member --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user`,
		},
		{
			wantError: true,
			name:      "create fails when NAME argument is missing",
			cmd:       "create service-account --privilege member" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError: true,
			name:      "create fails when --privilege is missing",
			cmd:       "create service-account ci-bot" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"privilege\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ServiceAccountCommandTestSuite) TestListServiceAccountsCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "list does not accept positional args",
			cmd:       "list service-accounts extra-arg" + suite.defaultKosliArguments,
			golden:    "Error: unknown command \"extra-arg\" for \"kosli list service-accounts\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ServiceAccountCommandTestSuite) TestUpdateServiceAccountCmd() {
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "update sends only the changed description (dry-run)",
			cmd:         "update service-account ci-bot --description 'new desc' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user/ci-bot.*"description": "new desc"`,
		},
		{
			wantError:   false,
			name:        "update sends only the changed privilege (dry-run)",
			cmd:         "update service-account ci-bot --privilege admin --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)"privilege": "admin"`,
		},
		{
			wantError:   true,
			name:        "update fails when neither --description nor --privilege is set",
			cmd:         "update service-account ci-bot" + suite.defaultKosliArguments,
			goldenRegex: `at least one of --description, --privilege is required`,
		},
		{
			wantError: true,
			name:      "update fails when NAME argument is missing",
			cmd:       "update service-account --privilege admin" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ServiceAccountCommandTestSuite) TestDeleteServiceAccountCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "delete without confirmation (empty stdin) is cancelled and makes no call",
			cmd:       "delete service-account ci-bot" + suite.defaultKosliArguments,
			golden:    "Are you sure you want to delete service account(s) ci-bot? [y/N] Deletion of service account(s) ci-bot was cancelled.\n",
		},
		{
			wantError:   false,
			name:        "delete with --assume-yes and --dry-run builds the right url",
			cmd:         "delete service-account ci-bot --assume-yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/ci-bot`,
		},
		{
			wantError:   false,
			name:        "delete accepts multiple names (dry-run)",
			cmd:         "delete service-account sa1 sa2 --assume-yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user/sa1.*service-accounts/docs-cmd-test-user/sa2`,
		},
		{
			wantError:   false,
			name:        "the sa alias, -y shorthand and hidden --yes work",
			cmd:         "delete sa ci-bot -y --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/ci-bot`,
		},
		{
			wantError: true,
			name:      "delete fails when NAME argument is missing",
			cmd:       "delete service-account" + suite.defaultKosliArguments,
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestServiceAccountSuccessOutput stubs successful (2xx) API responses to verify
// that create/list/get/update render the server's response on the happy path.
func (suite *ServiceAccountCommandTestSuite) TestServiceAccountSuccessOutput() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user").
		Reply(201).
		BodyString(serviceAccountFixture(suite.T(), "created_service_account.json"))
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user").
		Reply(200).
		BodyString(serviceAccountFixture(suite.T(), "listed_service_accounts.json"))
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user/ci-bot").
		Reply(200).
		BodyString(serviceAccountFixture(suite.T(), "service_account.json"))
	fake.NewHandler().
		Patch("/api/v2/service-accounts/docs-cmd-test-user/ci-bot").
		Reply(200).
		BodyString(serviceAccountFixture(suite.T(), "updated_service_account.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "create reports the created service account",
			cmd:         "create service-account ci-bot --privilege member" + args,
			goldenRegex: `service account ci-bot was created`,
		},
		{
			wantError:   false,
			name:        "list prints the returned service accounts",
			cmd:         "list service-accounts --output json" + args,
			goldenRegex: `(?s)ci-bot.*deployer`,
		},
		{
			wantError:   false,
			name:        "get prints the service account",
			cmd:         "get service-account ci-bot --output json" + args,
			goldenRegex: `ci-bot`,
		},
		{
			wantError:   false,
			name:        "update reports the updated service account",
			cmd:         "update service-account ci-bot --privilege admin" + args,
			goldenRegex: `service account ci-bot was updated`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestServiceAccountDeleteSuccess stubs a successful delete and verifies the
// CLI reports it.
func (suite *ServiceAccountCommandTestSuite) TestServiceAccountDeleteSuccess() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/ci-bot").
		Reply(200).
		BodyString(serviceAccountFixture(suite.T(), "delete_success.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "delete reports the deleted service account",
			cmd:         "delete service-account ci-bot --assume-yes" + args,
			goldenRegex: `service account ci-bot was deleted!`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestServiceAccountDeletePartialFailure verifies that when one name in a
// multi-name delete fails, the names already deleted are reported before the
// error is surfaced.
func (suite *ServiceAccountCommandTestSuite) TestServiceAccountDeletePartialFailure() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/sa1").
		Reply(200).
		BodyString(serviceAccountFixture(suite.T(), "delete_success.json"))
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/sa2").
		Reply(404).
		BodyString(serviceAccountFixture(suite.T(), "error_service_account_not_found.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "delete reports deleted accounts before a later one fails",
			cmd:         "delete service-account sa1 sa2 --assume-yes" + args,
			goldenRegex: `(?s)service account sa1 was deleted!.*already deleted before this failure: sa1.*failed to delete service account: Service account not found`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestServiceAccountApiErrorsAreSurfaced stubs 4xx responses to verify that the
// commands surface the server's error message instead of succeeding.
func (suite *ServiceAccountCommandTestSuite) TestServiceAccountApiErrorsAreSurfaced() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user/missing-sa").
		Reply(404).
		BodyString(serviceAccountFixture(suite.T(), "error_service_account_not_found.json"))
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user").
		Reply(403).
		BodyString(serviceAccountFixture(suite.T(), "error_forbidden.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "get surfaces a 404 from the API as an error",
			cmd:         "get service-account missing-sa" + args,
			goldenRegex: `Error: Service account not found`,
		},
		{
			wantError:   true,
			name:        "list surfaces a 403 from the API as an error",
			cmd:         "list service-accounts" + args,
			goldenRegex: `Error: You don't have permission to access this resource`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestServiceAccountCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountCommandTestSuite))
}
