package main

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

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
}

func TestPrintApiKeysAsTable(t *testing.T) {
	// The rotate command aggregates one or more rotated keys into a JSON array.
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
type ServiceAccountApiKeysCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *ServiceAccountApiKeysCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *ServiceAccountApiKeysCommandTestSuite) TestCreateApiKeyCmd() {
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "create builds the right url and payload (dry-run)",
			cmd:         "service-account api-keys create --service-account test-sa --description 'ci key' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user/test-sa/api-keys.*"description": "ci key"`,
		},
		{
			wantError:   false,
			name:        "create with a date --expires-at converts to an epoch timestamp (dry-run)",
			cmd:         "service-account api-keys create --service-account test-sa --description 'ci key' --expires-at 2026-12-31 --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)"description": "ci key".*"expires_at": 1798675200`,
		},
		{
			wantError:   false,
			name:        "the sa/ak aliases work",
			cmd:         "sa ak create --service-account test-sa --description 'ci key' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys`,
		},
		{
			wantError:   false,
			name:        "the create aliases and -s shorthand work",
			cmd:         "sa ak c -s test-sa --description 'ci key' --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys`,
		},
		{
			wantError: true,
			name:      "create fails when --service-account is missing",
			cmd:       "service-account api-keys create --description 'ci key'" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
		{
			wantError: true,
			name:      "create fails when --description is missing",
			cmd:       "service-account api-keys create --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"description\" not set\n",
		},
		{
			wantError:   true,
			name:        "create fails with an invalid --expires-at value",
			cmd:         "service-account api-keys create --service-account test-sa --description 'ci key' --expires-at not-a-date --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `Error: invalid --expires-at value`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ServiceAccountApiKeysCommandTestSuite) TestRotateApiKeyCmd() {
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "rotate builds the right url and default grace period (dry-run)",
			cmd:         "service-account api-keys rotate key-123 --service-account test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123/rotate.*"grace_period_hours": 24`,
		},
		{
			wantError:   false,
			name:        "rotate honours a custom --grace-period-hours (dry-run)",
			cmd:         "service-account api-keys rotate key-123 --service-account test-sa --grace-period-hours 48 --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `"grace_period_hours": 48`,
		},
		{
			wantError:   false,
			name:        "the rotate alias and -s shorthand work",
			cmd:         "sa ak ro key-123 -s test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123/rotate`,
		},
		{
			wantError:   false,
			name:        "the -g and -e shorthands work (dry-run)",
			cmd:         "sa ak ro key-123 -s test-sa -g 1 -e 2026-6-5 --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)"grace_period_hours": 1.*"expires_at": 1780617600`,
		},
		{
			wantError:   false,
			name:        "rotate accepts multiple KEY-IDs (dry-run)",
			cmd:         "service-account api-keys rotate key-1 key-2 --service-account test-sa --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)api-keys/key-1/rotate.*api-keys/key-2/rotate`,
		},
		{
			wantError: true,
			name:      "rotate fails when KEY-ID argument is missing",
			cmd:       "service-account api-keys rotate --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
		{
			wantError: true,
			name:      "rotate fails when --service-account is missing",
			cmd:       "service-account api-keys rotate key-123" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ServiceAccountApiKeysCommandTestSuite) TestRevokeApiKeyCmd() {
	tests := []cmdTestCase{
		{
			wantError: false,
			name:      "revoke without confirmation (empty stdin) is cancelled and makes no call",
			cmd:       "service-account api-keys revoke key-123 --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Are you sure you want to revoke API key(s) key-123 for service account test-sa? [y/N] revocation of API key(s) key-123 was cancelled\n",
		},
		{
			wantError:   false,
			name:        "revoke accepts multiple KEY-IDs (dry-run)",
			cmd:         "service-account api-keys revoke key-1 key-2 --service-account test-sa --assume-yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `(?s)api-keys/key-1.*api-keys/key-2`,
		},
		{
			wantError:   false,
			name:        "revoke with --assume-yes and --dry-run builds the right url",
			cmd:         "service-account api-keys revoke key-123 --service-account test-sa --assume-yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123`,
		},
		{
			wantError:   false,
			name:        "revoke with --yes bypasses confirmation too",
			cmd:         "service-account api-keys revoke key-123 --service-account test-sa --yes --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123`,
		},
		{
			wantError:   false,
			name:        "the del alias and -s shorthand work",
			cmd:         "sa ak del key-123 -s test-sa -y --dry-run" + suite.defaultKosliArguments,
			goldenRegex: `service-accounts/docs-cmd-test-user/test-sa/api-keys/key-123`,
		},
		{
			wantError: true,
			name:      "revoke fails when KEY-ID argument is missing",
			cmd:       "service-account api-keys revoke --service-account test-sa" + suite.defaultKosliArguments,
			golden:    "Error: requires at least 1 arg(s), only received 0\n",
		},
		{
			wantError: true,
			name:      "revoke fails when --service-account is missing",
			cmd:       "service-account api-keys revoke key-123" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestApiKeysSuccessOutput stubs successful (2xx) API responses to verify that
// create/list/rotate render the server's response on the happy path.
func (suite *ServiceAccountApiKeysCommandTestSuite) TestApiKeysSuccessOutput() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys").
		Reply(201).
		BodyString(`{"id":"id-1","key":"sk_created","description":"ci","created_at":1780584129.5,"expires_at":0}`)
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys").
		Reply(200).
		BodyString(`[{"id":"id-1","description":"ci","created_at":1780584129.5,"expires_at":0,"last_used_at":0}]`)
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k1/rotate").
		Reply(201).
		BodyString(`{"id":"k1","key":"sk_one","description":"one","created_at":1,"expires_at":0}`)
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/k2/rotate").
		Reply(201).
		BodyString(`{"id":"k2","key":"sk_two","description":"two","created_at":1,"expires_at":0}`)

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "create prints the new key value",
			cmd:         "service-account api-keys create -s test-sa -d ci --output json" + args,
			goldenRegex: `sk_created`,
		},
		{
			wantError:   false,
			name:        "list prints the returned keys",
			cmd:         "service-account api-keys list -s test-sa --output json" + args,
			goldenRegex: `id-1`,
		},
		{
			wantError:   false,
			name:        "rotate of multiple keys prints all new key values",
			cmd:         "service-account api-keys rotate k1 k2 -s test-sa --output json" + args,
			goldenRegex: `(?s)sk_one.*sk_two`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestRevokeApiKeyNotFound stubs the API with a 404 to verify that revoking a
// non-existing key surfaces the server's "API key not found" error instead of
// reporting success.
func (suite *ServiceAccountApiKeysCommandTestSuite) TestRevokeApiKeyNotFound() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Delete("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/missing-key").
		Reply(404).
		BodyString(`{"message": "API key not found"}`)

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "revoke surfaces a 404 from the API as an error",
			cmd:         "service-account api-keys revoke missing-key --service-account test-sa --assume-yes" + args,
			goldenRegex: `Error: API key not found`,
		},
	}

	runTestCmd(suite.T(), tests)
}

// TestApiErrorsAreSurfaced stubs the API with 4xx responses to verify that
// create/rotate/list surface the server's error message instead of succeeding.
func (suite *ServiceAccountApiKeysCommandTestSuite) TestApiErrorsAreSurfaced() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/missing-sa/api-keys").
		Reply(404).
		BodyString(`{"message": "Service account not found"}`)
	fake.NewHandler().
		Post("/api/v2/service-accounts/docs-cmd-test-user/test-sa/api-keys/missing-key/rotate").
		Reply(404).
		BodyString(`{"message": "API key not found"}`)
	fake.NewHandler().
		Get("/api/v2/service-accounts/docs-cmd-test-user/missing-sa/api-keys").
		Reply(403).
		BodyString(`{"message": "You don't have permission to access this resource"}`)

	args := fmt.Sprintf(" --host %s --org %s --api-token %s", fake.Server.URL, global.Org, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   true,
			name:        "create surfaces a 404 from the API as an error",
			cmd:         "service-account api-keys create --service-account missing-sa --description x" + args,
			goldenRegex: `Error: Service account not found`,
		},
		{
			wantError:   true,
			name:        "rotate surfaces a 404 from the API as an error",
			cmd:         "service-account api-keys rotate missing-key --service-account test-sa" + args,
			goldenRegex: `Error: API key not found`,
		},
		{
			wantError:   true,
			name:        "list surfaces a 403 from the API as an error",
			cmd:         "service-account api-keys list --service-account missing-sa" + args,
			goldenRegex: `Error: You don't have permission to access this resource`,
		},
	}

	runTestCmd(suite.T(), tests)
}

func (suite *ServiceAccountApiKeysCommandTestSuite) TestListApiKeysCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "list fails when --service-account is missing",
			cmd:       "service-account api-keys list" + suite.defaultKosliArguments,
			golden:    "Error: required flag(s) \"service-account\" not set\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestServiceAccountApiKeysCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceAccountApiKeysCommandTestSuite))
}
