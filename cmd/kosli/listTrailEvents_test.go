package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// trailEventsStub serves body with status for every request and records the
// path and query of each request it was asked for.
type trailEventsStub struct {
	server   *httptest.Server
	requests int
	path     string
	query    url.Values
}

func newTrailEventsStub(t *testing.T, status int, body string) *trailEventsStub {
	t.Helper()
	stub := &trailEventsStub{}
	stub.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stub.requests++
		stub.path = r.URL.Path
		stub.query = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(stub.server.Close)
	return stub
}

func (s *trailEventsStub) run(args string) (string, error) {
	_, combined, _, _, err := executeCommandC(fmt.Sprintf(
		"list trail-events %s --host %s --org my-org --api-token secret --max-api-retries 0", args, s.server.URL))
	return combined, err
}

func TestListTrailEventsRequest(t *testing.T) {
	t.Run("json output is the response body of the events endpoint", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[{"type":"trail_reported","timestamp":1.5}],"_links":{"self":{"href":"x"}}}`)
		out, err := stub.run("my-trail --flow my-flow --output json")
		require.NoError(t, err)
		require.Equal(t, "/api/v2/trails/my-org/my-flow/my-trail/events", stub.path)
		require.JSONEq(t, `{"events":[{"type":"trail_reported","timestamp":1.5}],"_links":{"self":{"href":"x"}}}`, out)
	})

	t.Run("the API's error message is printed", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusNotFound, `{"message":"Trail 'my-trail' not found"}`)
		out, err := stub.run("my-trail --flow my-flow --output json")
		require.Error(t, err)
		require.Contains(t, out, "Trail 'my-trail' not found")
	})
}

const emptyTrailEventsPage = `{"events":[],"_links":{"self":{"href":"x"}}}`

func TestListTrailEventsSort(t *testing.T) {
	t.Run("events are requested newest first by default", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		_, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Equal(t, "newest", stub.query.Get("sort"))
	})

	t.Run("--sort oldest requests them oldest first", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		_, err := stub.run("my-trail --flow my-flow --sort oldest")
		require.NoError(t, err)
		require.Equal(t, "oldest", stub.query.Get("sort"))
	})

	t.Run("an unknown order fails without a request", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		out, err := stub.run("my-trail --flow my-flow --sort sideways")
		require.Error(t, err)
		require.Contains(t, out, "Error: --sort must be one of: newest, oldest\n")
		require.Zero(t, stub.requests)
	})
}

func TestListTrailEventsTable(t *testing.T) {
	t.Run("table is the default output", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[
			{"type":"trail_attestation_reported","timestamp":1700000002.5,"template_reference_name":"bar",
			 "git_commit_info":{"sha1":"0fc1ba9876f91b215679f3649b8668085d820ab5"}},
			{"type":"attestation_overridden","timestamp":1700000001,"attestation_name":"foo"},
			{"type":"artifact_attestation_reported","timestamp":1700000001,"template_reference_name":"foo","artifact_name":"cli-image"},
			{"type":"trail_reported","timestamp":1700000000}
		],"_links":{"self":{"href":"x"}}}`)
		out, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Equal(t, ""+
			"TIMESTAMP                      TYPE                           NAME  COMMIT\n"+
			"Sat, 16 Jan 2016 00:00:00 UTC  trail_attestation_reported     bar   0fc1ba9\n"+
			"Sat, 16 Jan 2016 00:00:00 UTC  attestation_overridden         foo   \n"+
			"Sat, 16 Jan 2016 00:00:00 UTC  artifact_attestation_reported  foo   \n"+
			"Sat, 16 Jan 2016 00:00:00 UTC  trail_reported                       \n", out)
	})

	t.Run("an artifact event is named by its artifact when it has no template name", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[
			{"type":"artifact_creation_reported","timestamp":1700000000,"artifact_name":"cli-image"}
		],"_links":{}}`)
		out, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Contains(t, out, "artifact_creation_reported  cli-image")
	})

	t.Run("an empty page says so", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[],"_links":{"self":{"href":"x"}}}`)
		out, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Equal(t, "No events were found.\n", out)
	})
}

type ListTrailEventsCommandTestSuite struct {
	suite.Suite
	flowName              string
	trailName             string
	defaultKosliArguments string
}

func (suite *ListTrailEventsCommandTestSuite) SetupTest() {
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.flowName = "list-trail-events"
	suite.trailName = "trail-name"
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --org %s --api-token %s", global.Host, global.Org, global.ApiToken)

	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
	CreateGenericTrailAttestation(suite.flowName, suite.trailName, "bar", suite.T())
}

func (suite *ListTrailEventsCommandTestSuite) TestListTrailEventsCmd() {
	tests := []cmdTestCase{
		{
			name: "listing events with --output json prints the trail's events",
			cmd:  fmt.Sprintf(`list trail-events %s --flow %s --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"events", "length:2"},
				{"events.[0].type", "trail_attestation_reported"},
				{"events.[1].type", "trail_reported"},
			},
		},
		{
			name: "listing events with --sort oldest starts at the event that began the trail",
			cmd:  fmt.Sprintf(`list trail-events %s --flow %s --sort oldest --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"events.[0].type", "trail_reported"},
				{"events.[1].type", "trail_attestation_reported"},
			},
		},
		{
			wantError:   true,
			name:        "listing events fails with an unknown --sort",
			cmd:         fmt.Sprintf(`list trail-events %s --flow %s --sort sideways %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "^Error: --sort must be one of: newest, oldest\n",
		},
		{
			name:       "listing events prints a table",
			cmd:        fmt.Sprintf(`list trail-events %s --flow %s %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenFile: "output/list/list-trail-events.txt",
		},
		{
			wantError: true,
			name:      "listing events fails without --flow",
			cmd:       fmt.Sprintf(`list trail-events %s %s`, suite.trailName, suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"flow\" not set\n",
		},
		{
			wantError: true,
			name:      "listing events fails without the trail name",
			cmd:       fmt.Sprintf(`list trail-events --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			golden:    "Error: accepts 1 arg(s), received 0\n",
		},
		{
			wantError:   true,
			name:        "listing events of a trail that does not exist fails",
			cmd:         fmt.Sprintf(`list trail-events non-existing --flow %s %s`, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "^Error: .*non-existing.*",
		},
		{
			wantError:   true,
			name:        "listing events in a flow that does not exist fails",
			cmd:         fmt.Sprintf(`list trail-events %s --flow non-existing %s`, suite.trailName, suite.defaultKosliArguments),
			goldenRegex: "^Error: .*non-existing.*",
		},
	}
	runTestCmd(suite.T(), tests)
}

func TestListTrailEventsCommandTestSuite(t *testing.T) {
	suite.Run(t, new(ListTrailEventsCommandTestSuite))
}
