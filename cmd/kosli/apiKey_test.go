package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// apiKeyFixture reads an api-key response fixture from testdata/service-account.
// The fixtures hold the canonical API response bodies so the response contract
// lives in one place (see the README there).
func apiKeyFixture(t *testing.T, name string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "service-account", name))
	require.NoError(t, err, "failed to read fixture %s", name)
	return string(body)
}

func TestPrintApiKeyAsTable(t *testing.T) {
	// The API returns timestamps as floating-point epoch seconds (with a
	// fractional part), so the response struct/table rendering must accept them.
	raw := `{"id":"key-1","key":"sk_secret_value","description":"ci key","created_at":1780584129.6878593,"expires_at":0,"grace_period_expires_at":1780670529.5}`

	var buf bytes.Buffer
	err := printApiKeyAsTable(raw, &buf, 0)
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "key-1")
	require.Contains(t, out, "sk_secret_value")
	require.Contains(t, out, "ci key")
	require.Contains(t, out, "Old Key Valid Until")
	// expires_at of 0 means "no expiry" and must render as N/A, not epoch zero
	require.Regexp(t, `Expires At:\s+N/A`, out)
	require.NotContains(t, out, "1970")
}

func TestPrintApiKeysAsTable(t *testing.T) {
	// The update --rotate command aggregates one or more rotated keys into a JSON array.
	raw := `[{"id":"key-1","key":"sk_one","description":"first","created_at":1780584129.5,"expires_at":0},` +
		`{"id":"key-2","key":"sk_two","description":"second","created_at":1780584130.5,"expires_at":0}]`

	var buf bytes.Buffer
	err := printApiKeysAsTable(raw, &buf, 0)
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "key-1")
	require.Contains(t, out, "sk_one")
	require.Contains(t, out, "key-2")
	require.Contains(t, out, "sk_two")
}

func TestPrintApiKeysListAsTable(t *testing.T) {
	// The list endpoint returns key metadata only (no secret key value).
	raw := `[{"id":"key-1","description":"first","created_at":1780584129.5,"expires_at":0,"last_used_at":0},` +
		`{"id":"key-2","description":"second","created_at":1780584130.5,"expires_at":0,"last_used_at":0}]`

	var buf bytes.Buffer
	err := printApiKeysListAsTable(raw, &buf, 0)
	require.NoError(t, err)

	out := buf.String()
	for _, want := range []string{"ID", "DESCRIPTION", "CREATED", "EXPIRES", "LAST USED", "key-1", "first", "key-2", "second"} {
		require.Contains(t, out, want)
	}
	// expires_at and last_used_at of 0 must render as N/A, not epoch zero (1970)
	require.Contains(t, out, "N/A")
	require.NotContains(t, out, "1970")
}

func TestParseExpiresAt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "empty returns zero", input: "", want: 0},
		{name: "bare epoch is passed through", input: "1798675200", want: 1798675200},
		{name: "date only is parsed as UTC midnight", input: "2026-06-04", want: time.Date(2026, 6, 4, 0, 0, 0, 0, time.UTC).Unix()},
		{name: "date with unpadded day/month is parsed", input: "2026-6-5", want: time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC).Unix()},
		{name: "date and time is parsed as UTC", input: "2026-06-04 15:04:05", want: time.Date(2026, 6, 4, 15, 4, 5, 0, time.UTC).Unix()},
		{name: "RFC3339 is parsed", input: "2026-06-04T15:04:05Z", want: time.Date(2026, 6, 4, 15, 4, 5, 0, time.UTC).Unix()},
		{name: "invalid value errors", input: "not-a-date", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseExpiresAt(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// Define the suite, and absorb the built-in basic suite functionality from testify.
type ApiKeyCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *ApiKeyCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ApiKeyCommandTestSuite) TestCreateApiKeyCmd() {
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "create builds the right url and payload (dry-run)",
			cmd:         "create api-key --service-account test-sa --description 'ci key' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user/test-sa/api-keys.*"description": "ci key"`,
		},
		{
			wantError:   false,
			name:        "create with a date --expires-at converts to an epoch timestamp (dry-run)",
			cmd:         "create api-key --service-account test-sa --description 'ci key' --expires-at 2026-12-31 --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)"description": "ci key".*"expires_at": 1798675200`,
		},
		{
			wantError:   false,
			name:        "the api-key alias (ak) and -s shorthand work",
			cmd:         "create ak -s test-sa --description 'ci key' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys`,
		},
		{
			wantError: true,
			name:      "create fails when --service-account is missing",
			cmd:       "create api-key --description 'ci key'" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
		{
			wantError: true,
			name:      "create fails when --description is missing",
			cmd:       "create api-key --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"description\" not set\n",
		},
		{
			wantError:   true,
			name:        "create fails with an invalid --expires-at value",
			cmd:         "create api-key --service-account test-sa --description 'ci key' --expires-at not-a-date --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `Error: invalid --expires-at value`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ApiKeyCommandTestSuite) TestUpdateApiKeyCmd() {
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "rotate without --grace-period-hours sends an empty payload (server owns the default)",
			cmd:         "update api-key key-123 --rotate --service-account test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123/rotate.*real run:\s*\{\}`,
		},
		{
			wantError:   false,
			name:        "rotate honours a custom --grace-period-hours (dry-run)",
			cmd:         "update api-key key-123 --rotate --service-account test-sa --grace-period-hours 48 --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `"grace_period_hours": 48`,
		},
		{
			wantError:   false,
			name:        "the api-key alias (ak) and -s shorthand work",
			cmd:         "update ak key-123 --rotate -s test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123/rotate`,
		},
		{
			wantError:   false,
			name:        "the update verb aliases (up, u) work",
			cmd:         "u ak key-123 --rotate -s test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123/rotate`,
		},
		{
			wantError:   false,
			name:        "the -R, -g and -e shorthands work (dry-run)",
			cmd:         "update ak key-123 -R -s test-sa -g 1 -e 2026-6-5 --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)"grace_period_hours": 1.*"expires_at": 1780617600`,
		},
		{
			wantError:   false,
			name:        "update --rotate accepts multiple KEY-IDs (dry-run)",
			cmd:         "update api-key key-1 key-2 --rotate --service-account test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)api-keys/key-1/rotate.*api-keys/key-2/rotate`,
		},
		{
			wantError:   true,
			name:        "update without --rotate has nothing to do",
			cmd:         "update api-key key-123 --service-account test-sa" + suite.defaultKosliArguments,
			goldenRegex: `Error: nothing to update`,
		},
		{
			wantError: true,
			name:      "update fails when KEY-ID argument is missing",
			cmd:       "update api-key --rotate --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
		{
			wantError: true,
			name:      "update fails when --service-account is missing",
			cmd:       "update api-key key-123 --rotate" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ApiKeyCommandTestSuite) TestDeleteApiKeyCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "delete without confirmation (empty stdin) is cancelled and makes no call",
			cmd:       "delete api-key key-123 --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Are you sure you want to delete API key(s) key-123 for service account test-sa? [y/N]\ndeletion of API key(s) key-123 was cancelled\n",
		},
		{
			wantError:   false,
			name:        "delete accepts multiple KEY-IDs (dry-run)",
			cmd:         "delete api-key key-1 key-2 --service-account test-sa --assume-yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)api-keys/key-1.*api-keys/key-2`,
		},
		{
			wantError:   false,
			name:        "delete with --assume-yes and --dry-run builds the right url",
			cmd:         "delete api-key key-123 --service-account test-sa --assume-yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123`,
		},
		{
			wantError:   false,
			name:        "the --yes alias bypasses confirmation too",
			cmd:         "delete api-key key-123 --service-account test-sa --yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123`,
		},
		{
			wantError:   false,
			name:        "the api-key alias (ak), -s and -y shorthands work",
			cmd:         "delete ak key-123 -s test-sa -y --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123`,
		},
		{
			wantError: true,
			name:      "delete fails when KEY-ID argument is missing",
			cmd:       "delete api-key --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
		{
			wantError: true,
			name:      "delete fails when --service-account is missing",
			cmd:       "delete api-key key-123" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestApiKeysSuccessOutput stubs successful (2xx) API responses to verify that
// create/list/update render the server's response on the happy path.
func (suite *ApiKeyCommandTestSuite) TestApiKeysSuccessOutput() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys").
		Reply(201).
		BodyString(apiKeyFixture(suite.T(), "created_api_key.json"))
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys").
		Reply(200).
		BodyString(apiKeyFixture(suite.T(), "listed_api_keys.json"))
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k1/rotate").
		Reply(201).
		BodyString(apiKeyFixture(suite.T(), "rotated_api_key.json"))
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k2/rotate").
		Reply(201).
		BodyString(apiKeyFixture(suite.T(), "rotated_api_key.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "create prints the new key value",
			cmd:         "create api-key -s test-sa -d ci --output json" + args,
			goldenRegex: `sk_created`,
		},
		{
			wantError:   false,
			name:        "list prints the returned keys",
			cmd:         "list api-keys -s test-sa --output json" + args,
			goldenRegex: `id-1`,
		},
		{
			wantError: false,
			name:      "update --rotate of multiple keys prints all rotated keys",
			cmd:       "update api-key k1 k2 --rotate -s test-sa --output json" + args,
			goldenJson: []jsonCheck{
				{Path: "", Want: "length:2"},
				{Path: "[0].key", Want: "sk_one"},
			},
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestUpdatePartialFailure verifies that when one key in a multi-key rotate
// fails, the keys already rotated are still printed (their values are only
// returned once) before the error is surfaced.
func (suite *ApiKeyCommandTestSuite) TestUpdatePartialFailure() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k1/rotate").
		Reply(201).
		BodyString(apiKeyFixture(suite.T(), "rotated_api_key.json"))
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k2/rotate").
		Reply(404).
		BodyString(apiKeyFixture(suite.T(), "error_api_key_not_found.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "rotate prints already-rotated keys then surfaces the error",
			cmd:         "update api-key k1 k2 --rotate -s test-sa --output json" + args,
			goldenRegex: `(?s)sk_one.*Error: API key not found`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestDeletePartialFailure verifies that when one key in a multi-key delete
// fails, the keys already deleted are reported (deletion is destructive and
// one-way) before the error is surfaced.
func (suite *ApiKeyCommandTestSuite) TestDeletePartialFailure() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k1").
		Reply(200).
		BodyString(apiKeyFixture(suite.T(), "revoke_success.json"))
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k2").
		Reply(404).
		BodyString(apiKeyFixture(suite.T(), "error_api_key_not_found.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "delete reports deleted keys before a later key fails",
			cmd:         "delete api-key k1 k2 -s test-sa --assume-yes" + args,
			goldenRegex: `(?s)API key k1 for service account test-sa was deleted.*already deleted before this failure: k1.*failed to delete API key k2.*API key not found`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestDeleteApiKeyNotFound stubs the API with a 404 to verify that deleting a
// non-existing key surfaces the server's "API key not found" error instead of
// reporting success.
func (suite *ApiKeyCommandTestSuite) TestDeleteApiKeyNotFound() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/missing-key").
		Reply(404).
		BodyString(apiKeyFixture(suite.T(), "error_api_key_not_found.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "delete surfaces a 404 from the API as an error",
			cmd:         "delete api-key missing-key --service-account test-sa --assume-yes" + args,
			goldenRegex: `(?s)failed to delete API key missing-key.*API key not found`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestApiErrorsAreSurfaced stubs the API with 4xx responses to verify that
// create/update/list surface the server's error message instead of succeeding.
func (suite *ApiKeyCommandTestSuite) TestApiErrorsAreSurfaced() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/missing-sa/api-keys").
		Reply(404).
		BodyString(apiKeyFixture(suite.T(), "error_service_account_not_found.json"))
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/missing-key/rotate").
		Reply(404).
		BodyString(apiKeyFixture(suite.T(), "error_api_key_not_found.json"))
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user/missing-sa/api-keys").
		Reply(403).
		BodyString(apiKeyFixture(suite.T(), "error_forbidden.json"))

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "create surfaces a 404 from the API as an error",
			cmd:         "create api-key --service-account missing-sa --description x" + args,
			goldenRegex: `Error: Service account not found`,
		},
		{
			wantError:   true,
			name:        "update --rotate surfaces a 404 from the API as an error",
			cmd:         "update api-key missing-key --rotate --service-account test-sa" + args,
			goldenRegex: `Error: API key not found`,
		},
		{
			wantError:   true,
			name:        "list surfaces a 403 from the API as an error",
			cmd:         "list api-keys --service-account missing-sa" + args,
			goldenRegex: `Error: You don't have permission to access this resource`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ApiKeyCommandTestSuite) TestListApiKeysCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "list fails when --service-account is missing",
			cmd:       "list api-keys" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestApiKeyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ApiKeyCommandTestSuite))
}
