package main

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/maxcnunes/httpfake"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPrintDefaultOrgAsTable(t *testing.T) {
	raw := `{"default_org_name":"test-org"}`

	var buf bytes.Buffer
	err := printDefaultOrgAsTable(raw, &buf, 0)
	require.NoError(t, err)

	out := buf.String()
	for _, want := range []string{"Default organization:", "test-org"} {
		require.Contains(t, out, want)
	}
}

type GetDefaultOrgCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
}

func (suite *GetDefaultOrgCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --api-token %s", global.Host, global.ApiToken)
}

func (suite *GetDefaultOrgCommandTestSuite) TestGetDefaultOrgCmd() {
	fake := httpfake.New()
	defer fake.Close()
	fake.NewHandler().
		Get("/api/v2/user/default-org").
		Reply(200).
		BodyString(`{"default_org_name":"test-org"}`)

	args := fmt.Sprintf(" --host %s --api-token %s", fake.Server.URL, global.ApiToken)
	tests := []cmdTestCase{
		{
			wantError:   false,
			name:        "get default-org prints the org name as a table",
			cmd:         "get default-org" + args,
			goldenRegex: `Default organization:\s+test-org`,
		},
		{
			wantError:   false,
			name:        "get default-org supports --output json",
			cmd:         "get default-org --output json" + args,
			goldenRegex: `(?s)"default_org_name":\s*"test-org"`,
		},
		{
			wantError: true,
			name:      "get default-org fails when an argument is provided",
			cmd:       "get default-org extra-arg" + args,
			golden:    "Error: unknown command \"extra-arg\" for \"kosli get default-org\"\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestGetDefaultOrgCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetDefaultOrgCommandTestSuite))
}
