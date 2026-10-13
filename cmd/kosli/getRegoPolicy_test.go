package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	regoPolicyWithCurrentVersion = `{"identifier":"pr-approved","name":"Pull request approved",` +
		`"description":"Approved PRs only","newest_version":{"version":3,"status":"rejected"},` +
		`"current_version":{"version":2,"digest":"sha256:abc123","input_schema":"trail",` +
		`"params_schema":{"type":"object"}},` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","updated_at":1757000100.5}`
	regoPolicyWithNoValidVersion = `{"identifier":"pr-approved","name":"Pull request approved",` +
		`"newest_version":{"version":1,"status":"pending"},` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","updated_at":1757000000.5}`
)

// fakeRegoPolicyRead records the read and answers with body and status.
type fakeRegoPolicyRead struct {
	body        string
	status      int
	escapedPath string
	calls       int
}

func newFakeRegoPolicyRead(t *testing.T, body string) (*httptest.Server, *fakeRegoPolicyRead) {
	t.Helper()
	fake := &fakeRegoPolicyRead{body: body, status: http.StatusOK}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fake.calls++
		fake.escapedPath = r.URL.EscapedPath()
		w.WriteHeader(fake.status)
		_, _ = fmt.Fprint(w, fake.body)
	}))
	t.Cleanup(server.Close)
	return server, fake
}

type GetRegoPolicyCommandTestSuite struct {
	suite.Suite
}

func (suite *GetRegoPolicyCommandTestSuite) SetupTest() {
	// Makes formattedTimestamp print a fixed UTC time.
	suite.T().Setenv("KOSLI_TESTS", "true")
}

func (suite *GetRegoPolicyCommandTestSuite) getCmd(host, args string) string {
	return fmt.Sprintf("get rego-policy %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItIsHiddenAndBeta() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	cmd, _, err := root.Find([]string{"get", "rego-policy"})
	require.NoError(suite.T(), err)

	require.Equal(suite.T(), "rego-policy", cmd.Name())
	require.True(suite.T(), cmd.Hidden)
	require.Contains(suite.T(), cmd.Annotations, betaCLIAnnotation)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItTakesExactlyOneIdentifier() {
	for _, args := range []string{"", "pr-approved other"} {
		server, fake := newFakeRegoPolicyRead(suite.T(), regoPolicyWithCurrentVersion)

		_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, args))

		require.Error(suite.T(), err, args)
		require.Contains(suite.T(), combined, "accepts 1 arg(s)", args)
		require.Equal(suite.T(), 0, fake.calls, args)
	}
}

func (suite *GetRegoPolicyCommandTestSuite) TestItPrintsThePolicyAndItsCurrentVersion() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPolicyWithCurrentVersion)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved", fake.escapedPath)
	require.Equal(suite.T(),
		"Identifier:       pr-approved\n"+
			"Name:             Pull request approved\n"+
			"Description:      Approved PRs only\n"+
			"Created by:       Jane Doe\n"+
			"Created at:       Sat, 16 Jan 2016 00:00:00 UTC\n"+
			"Last updated:     Sat, 16 Jan 2016 00:00:00 UTC\n"+
			"Newest version:   3 (rejected)\n"+
			"Current version:  2\n"+
			"  Digest:         sha256:abc123\n"+
			"  Input schema:   trail\n"+
			"  Params schema:  {\"type\":\"object\"}\n",
		combined)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItSaysWhenNoVersionIsValidYet() {
	server, _ := newFakeRegoPolicyRead(suite.T(), regoPolicyWithNoValidVersion)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved"))

	require.NoError(suite.T(), err)
	require.NotContains(suite.T(), combined, "Description:")
	require.Contains(suite.T(), combined, "Newest version:   1 (pending)\n")
	require.Contains(suite.T(), combined, "Current version:  none, as no version is valid yet\n")
	require.NotContains(suite.T(), combined, "Digest:")
}

func (suite *GetRegoPolicyCommandTestSuite) TestItPrintsThePolicyAsJSON() {
	server, _ := newFakeRegoPolicyRead(suite.T(), regoPolicyWithCurrentVersion)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), regoPolicyWithCurrentVersion, combined)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItPrintsAValidVersion() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --version 1"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved/versions/1", fake.escapedPath)
	require.Equal(suite.T(),
		"Identifier:    pr-approved\n"+
			"Version:       1\n"+
			"Status:        valid\n"+
			"Digest:        sha256:abc123\n"+
			"Input schema:  trail\n"+
			"Published by:  Jane Doe\n"+
			"Published at:  Sat, 16 Jan 2016 00:00:00 UTC\n"+
			"Validated at:  Sat, 16 Jan 2016 00:00:00 UTC\n"+
			"OPA version:   1.21.1\n"+
			"Entrypoint:    policy\n",
		combined)
}

// get only reads, so a version that is not valid is a successful answer.
func (suite *GetRegoPolicyCommandTestSuite) TestItPrintsWhyAVersionIsNotValidAndExitsZero() {
	server, _ := newFakeRegoPolicyRead(suite.T(), regoPolicyRejected)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --version 1"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "Status:        rejected\n")
	require.Contains(suite.T(), combined,
		"Error:         compile: policy.rego:4: rego_parse_error: unexpected eof token\n")
	require.NotContains(suite.T(), combined, "Entrypoint:")
}

func (suite *GetRegoPolicyCommandTestSuite) TestItPrintsTheCommentAndParamsSchemaOfAVersion() {
	body := `{"identifier":"pr-approved","version":2,"status":"valid","digest":"sha256:abc123",` +
		`"input_schema":"trails","comment":"Also require two approvers",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000002.25,` +
		`"opa_version":"1.21.1","entrypoint":"policy","params_schema":{"type":"object"}}`
	server, _ := newFakeRegoPolicyRead(suite.T(), body)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --version 2"))

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, "Comment:        Also require two approvers\n")
	require.Contains(suite.T(), combined, "Params schema:  {\"type\":\"object\"}\n")
}

func (suite *GetRegoPolicyCommandTestSuite) TestItPrintsAVersionAsJSON() {
	server, _ := newFakeRegoPolicyRead(suite.T(), regoPolicyRejected)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --version 1 --output json"))

	require.NoError(suite.T(), err)
	require.JSONEq(suite.T(), regoPolicyRejected, combined)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItRefusesAVersionThatIsNotPositive() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --version 0"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "--version must be a positive integer")
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItKeepsTheIdentifierInOneSegment() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPolicyWithCurrentVersion)

	_, _, _, _, err := executeCommandC(suite.getCmd(server.URL, "a/b"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/a%2Fb", fake.escapedPath)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItRefusesAnIdentifierThatIsNotASegment() {
	server, fake := newFakeRegoPolicyRead(suite.T(), regoPolicyWithCurrentVersion)

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, ".."))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, `".." is not a valid rego policy identifier`)
	require.Equal(suite.T(), 0, fake.calls)
}

func (suite *GetRegoPolicyCommandTestSuite) TestItReturnsANotFoundInTheServersWords() {
	server, fake := newFakeRegoPolicyRead(suite.T(),
		`{"message":"Rego policy 'pr-approved' has no version 9 in organization 'test-org'"}`)
	fake.status = http.StatusNotFound

	_, combined, _, _, err := executeCommandC(suite.getCmd(server.URL, "pr-approved --version 9"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), combined, "Rego policy 'pr-approved' has no version 9 in organization 'test-org'")
}

func TestGetRegoPolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(GetRegoPolicyCommandTestSuite))
}
