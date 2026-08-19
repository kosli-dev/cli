package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CreateAttestationTypeTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *CreateAttestationTypeTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)
}

func (suite *CreateAttestationTypeTestSuite) TestCustomAttestationTypeCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when no arguments are provided",
			cmd:       "create attestation-type" + suite.defaultKosliArguments,
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			name:   "type name is provided",
			cmd:    "create attestation-type wibble" + suite.defaultKosliArguments,
			golden: "attestation-type wibble was created\n",
		},
		{
			name:   "type description is provided",
			cmd:    "create attestation-type wibble-2 --description 'description of attestation type'" + suite.defaultKosliArguments,
			golden: "attestation-type wibble-2 was created\n",
		},
		{
			name:   "type schema is provided",
			cmd:    "create attestation-type wibble-4 --schema testdata/person-schema.json" + suite.defaultKosliArguments,
			golden: "attestation-type wibble-4 was created\n",
		},
		{
			name:   "type jq evaluator is provided",
			cmd:    `create attestation-type wibble-5 --jq '.age > 21' --jq '.age < 50'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-5 was created\n",
		},
		{
			name:   `jq evaluators can include bare "`,
			cmd:    `create attestation-type wibble-6 --jq '.name | startswith("B")'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-6 was created\n",
		},
		{
			name:   "summary json is provided",
			cmd:    `create attestation-type wibble-7 --summary-json '[{"name":"Critical","expression":".critical_count"}]'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-7 was created\n",
		},
		{
			name:   "summary json expressions can contain commas and equals",
			cmd:    `create attestation-type wibble-8 --summary-json '[{"name":"Tool","expression":"[.a, .b] | map(select(.x == 1)) | length"}]'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-8 was created\n",
		},
		{
			name:   "empty summary json array is accepted",
			cmd:    `create attestation-type wibble-9 --summary-json '[]'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-9 was created\n",
		},
		{
			// Deliberately asymmetric with the empty --summary case below: a blank
			// JSON blob reads as "not given", a blank key=value entry as malformed.
			name:   "empty summary json string is accepted",
			cmd:    `create attestation-type wibble-10 --summary-json ''` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-10 was created\n",
		},
		{
			name:   "summary json and jq rules can be combined",
			cmd:    `create attestation-type wibble-11 --jq '.critical_count == 0' --summary-json '[{"name":"Critical","expression":".critical_count"}]'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-11 was created\n",
		},
		{
			name:   "summary json and schema can be combined",
			cmd:    `create attestation-type wibble-12 --schema testdata/person-schema.json --summary-json '[{"name":"Age","expression":".age"}]'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-12 was created\n",
		},
		{
			wantError: true,
			name:      "fails when summary json is malformed",
			cmd:       `create attestation-type wibble-bad --summary-json '[{"name":'` + suite.defaultKosliArguments,
			golden:    "Error: --summary-json is not valid JSON: unexpected end of JSON input\n",
		},
		{
			wantError: true,
			name:      "fails when summary json is an object not an array",
			cmd:       `create attestation-type wibble-bad --summary-json '{"name":"Critical","expression":".critical_count"}'` + suite.defaultKosliArguments,
			golden:    "Error: --summary-json must be a JSON array of {name, expression} entries\n",
		},
		{
			wantError: true,
			name:      "fails when a summary entry has no name",
			cmd:       `create attestation-type wibble-bad --summary-json '[{"expression":".critical_count"}]'` + suite.defaultKosliArguments,
			golden:    "Error: --summary-json entry 1 is missing a name\n",
		},
		{
			wantError: true,
			name:      "fails when a summary entry has no expression",
			cmd:       `create attestation-type wibble-bad --summary-json '[{"name":"Critical"},{"name":"Tool","expression":".t"}]'` + suite.defaultKosliArguments,
			golden:    "Error: --summary-json entry 1 is missing an expression\n",
		},
		{
			name:   "repeatable summary is provided",
			cmd:    `create attestation-type wibble-13 --summary 'Critical=.critical_count' --summary 'Tool=.scanner.name'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-13 was created\n",
		},
		{
			name:   "repeatable summary expressions can contain commas and equals",
			cmd:    `create attestation-type wibble-14 --summary 'Tool=[.a, .b] | map(select(.x == 1)) | length'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-14 was created\n",
		},
		{
			name:   "repeatable summary and jq rules can be combined",
			cmd:    `create attestation-type wibble-15 --jq '.critical_count == 0' --summary 'Critical=.critical_count'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-15 was created\n",
		},
		{
			name:   "repeatable summary and schema can be combined",
			cmd:    `create attestation-type wibble-16 --schema testdata/person-schema.json --summary 'Age=.age'` + suite.defaultKosliArguments,
			golden: "attestation-type wibble-16 was created\n",
		},
		{
			wantError: true,
			name:      "fails when a repeatable summary entry has no equals sign",
			cmd:       `create attestation-type wibble-bad --summary 'Critical'` + suite.defaultKosliArguments,
			golden:    "Error: --summary entry 1 must be in the form NAME=EXPRESSION\n",
		},
		{
			wantError: true,
			name:      "fails when a repeatable summary entry has no name",
			cmd:       `create attestation-type wibble-bad --summary '=.critical_count'` + suite.defaultKosliArguments,
			golden:    "Error: --summary entry 1 is missing a name\n",
		},
		{
			wantError: true,
			name:      "fails when a repeatable summary entry has no expression",
			cmd:       `create attestation-type wibble-bad --summary 'Critical='` + suite.defaultKosliArguments,
			golden:    "Error: --summary entry 1 is missing an expression\n",
		},
		{
			wantError: true,
			name:      "fails when a repeatable summary value is a bare expression",
			cmd:       `create attestation-type wibble-bad --summary '.failing == 0'` + suite.defaultKosliArguments,
			golden:    "Error: --summary entry 1 expression cannot start with '='\n",
		},
		{
			// Unlike --summary-json '', which is a no-op. Matters in CI, where
			// --summary "$VAR" with VAR unset fails instead of silently no-opping.
			wantError: true,
			name:      "fails when a repeatable summary value is empty",
			cmd:       `create attestation-type wibble-bad --summary ''` + suite.defaultKosliArguments,
			golden:    "Error: --summary entry 1 must be in the form NAME=EXPRESSION\n",
		},
		{
			wantError: true,
			name:      "fails when both summary flags are provided",
			cmd:       `create attestation-type wibble-bad --summary 'Critical=.critical_count' --summary-json '[{"name":"Critical","expression":".critical_count"}]'` + suite.defaultKosliArguments,
			golden:    "Error: only one of --summary, --summary-json is allowed\n",
		},
		{
			// MuXRequiredFlags keys off flag.Changed, so an explicitly empty
			// --summary-json still trips the exclusion despite being a no-op.
			wantError: true,
			name:      "fails when both summary flags are provided and summary json is empty",
			cmd:       `create attestation-type wibble-bad --summary 'Critical=.critical_count' --summary-json ''` + suite.defaultKosliArguments,
			golden:    "Error: only one of --summary, --summary-json is allowed\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCreateAttestationTypeTestSuite(t *testing.T) {
	suite.Run(t, new(CreateAttestationTypeTestSuite))
}

func TestParseSummaryJSON(t *testing.T) {
	t.Run("omitted flag leaves summary out of the payload", func(t *testing.T) {
		payload := CreateAttestationTypePayload{TypeName: "wibble"}
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		require.NotContains(t, string(body), "summary")
	})

	t.Run("empty array leaves summary out of the payload", func(t *testing.T) {
		summary, err := parseSummaryJSON("[]")
		require.NoError(t, err)
		require.Empty(t, summary)

		payload := CreateAttestationTypePayload{TypeName: "wibble", Summary: summary}
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		require.NotContains(t, string(body), "summary")
	})

	t.Run("entries keep their order and round-trip unmangled", func(t *testing.T) {
		summary, err := parseSummaryJSON(`[{"name":"Critical","expression":"[.a, .b] | map(select(.x == 1)) | length"},{"name":"Tool","expression":".scanner.name"}]`)
		require.NoError(t, err)
		require.Equal(t, []SummaryEntry{
			{Name: "Critical", Expression: "[.a, .b] | map(select(.x == 1)) | length"},
			{Name: "Tool", Expression: ".scanner.name"},
		}, summary)
	})

	// An unset flag arrives here as "". Parsing it must stay a no-op rather than
	// falling through to json.Unmarshal, which rejects an empty string.
	t.Run("blank values are treated as no summary", func(t *testing.T) {
		for _, value := range []string{"", "   ", "\t\n"} {
			summary, err := parseSummaryJSON(value)
			require.NoErrorf(t, err, "value %q", value)
			require.Emptyf(t, summary, "value %q", value)
		}
	})

	t.Run("rejects entries that are not usable", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			value   string
			wantErr string
		}{
			{
				name:    "whitespace-only name",
				value:   `[{"name":"  ","expression":".x"}]`,
				wantErr: "--summary-json entry 1 is missing a name",
			},
			{
				name:    "whitespace-only expression",
				value:   `[{"name":"Critical","expression":" "}]`,
				wantErr: "--summary-json entry 1 is missing an expression",
			},
			{
				// Proves the reported index tracks the entry's position rather
				// than being coincidentally right for a single-entry list.
				name:    "second entry is the bad one",
				value:   `[{"name":"A","expression":".a"},{"name":"B"}]`,
				wantErr: "--summary-json entry 2 is missing an expression",
			},
			{
				name:    "null entry",
				value:   `[null]`,
				wantErr: "--summary-json entry 1 is missing a name",
			},
			{
				name:    "array of numbers",
				value:   `[1,2,3]`,
				wantErr: "--summary-json must be a JSON array of {name, expression} entries",
			},
			{
				name:    "array of strings",
				value:   `["a"]`,
				wantErr: "--summary-json must be a JSON array of {name, expression} entries",
			},
			{
				name:    "trailing data after the array",
				value:   `[] trailing`,
				wantErr: "--summary-json is not valid JSON: invalid character 't' after top-level value",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				summary, err := parseSummaryJSON(tc.value)
				require.EqualError(t, err, tc.wantErr)
				require.Empty(t, summary)
			})
		}
	})
}

func TestParseSummaryFlags(t *testing.T) {
	t.Run("no flags leaves summary out of the payload", func(t *testing.T) {
		summary, err := parseSummaryFlags(nil)
		require.NoError(t, err)
		require.Empty(t, summary)

		payload := CreateAttestationTypePayload{TypeName: "wibble", Summary: summary}
		body, err := json.Marshal(payload)
		require.NoError(t, err)
		require.NotContains(t, string(body), "summary")
	})

	t.Run("entries keep the order the flags were given in", func(t *testing.T) {
		summary, err := parseSummaryFlags([]string{"Critical=.critical_count", "Tool=.scanner.name"})
		require.NoError(t, err)
		require.Equal(t, []SummaryEntry{
			{Name: "Critical", Expression: ".critical_count"},
			{Name: "Tool", Expression: ".scanner.name"},
		}, summary)
	})

	// The whole reason for splitting on the first "=" only: JQ expressions use
	// "==" for comparison, and commas are meaningful inside them. Both must
	// survive into the expression untouched.
	t.Run("splits on the first equals only", func(t *testing.T) {
		summary, err := parseSummaryFlags([]string{"Tool=[.a, .b] | map(select(.x == 1)) | length"})
		require.NoError(t, err)
		require.Equal(t, []SummaryEntry{
			{Name: "Tool", Expression: "[.a, .b] | map(select(.x == 1)) | length"},
		}, summary)
	})

	t.Run("whitespace around name and expression is trimmed", func(t *testing.T) {
		summary, err := parseSummaryFlags([]string{"  Critical  =  .critical_count  "})
		require.NoError(t, err)
		require.Equal(t, []SummaryEntry{
			{Name: "Critical", Expression: ".critical_count"},
		}, summary)
	})

	t.Run("rejects entries that are not usable", func(t *testing.T) {
		for _, tc := range []struct {
			name    string
			values  []string
			wantErr string
		}{
			{
				name:    "no separator",
				values:  []string{"Critical"},
				wantErr: "--summary entry 1 must be in the form NAME=EXPRESSION",
			},
			{
				name:    "empty value",
				values:  []string{""},
				wantErr: "--summary entry 1 must be in the form NAME=EXPRESSION",
			},
			{
				name:    "empty name",
				values:  []string{"=.critical_count"},
				wantErr: "--summary entry 1 is missing a name",
			},
			{
				name:    "whitespace-only name",
				values:  []string{"  =.critical_count"},
				wantErr: "--summary entry 1 is missing a name",
			},
			{
				name:    "empty expression",
				values:  []string{"Critical="},
				wantErr: "--summary entry 1 is missing an expression",
			},
			{
				name:    "whitespace-only expression",
				values:  []string{"Critical=   "},
				wantErr: "--summary entry 1 is missing an expression",
			},
			{
				// A bare jq expression using "==" carries a separator, so it
				// would otherwise parse to {".failing", "= 0"}. No valid jq
				// expression starts with "=" — it is only ever infix.
				name:    "bare expression using ==",
				values:  []string{".failing == 0"},
				wantErr: "--summary entry 1 expression cannot start with '='",
			},
			{
				name:    "doubled separator",
				values:  []string{"Critical== 0"},
				wantErr: "--summary entry 1 expression cannot start with '='",
			},
			{
				// Proves the reported index tracks the flag's position rather
				// than being coincidentally right for a single entry.
				name:    "second entry is the bad one",
				values:  []string{"A=.a", "B"},
				wantErr: "--summary entry 2 must be in the form NAME=EXPRESSION",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				summary, err := parseSummaryFlags(tc.values)
				require.EqualError(t, err, tc.wantErr)
				require.Empty(t, summary)
			})
		}
	})
}
