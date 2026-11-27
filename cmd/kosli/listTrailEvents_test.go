package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

func TestListTrailEventsAlias(t *testing.T) {
	stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
	_, _, _, _, err := executeCommandC(fmt.Sprintf(
		"list te my-trail --flow my-flow --host %s --org my-org --api-token secret --max-api-retries 0", stub.server.URL))
	require.NoError(t, err)
	require.Equal(t, "/api/v2/trails/my-org/my-flow/my-trail/events", stub.path)
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

func TestListTrailEventsPaging(t *testing.T) {
	t.Run("a page holds 50 events by default", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		_, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Equal(t, "50", stub.query.Get("per_page"))
		require.False(t, stub.query.Has("after"))
		require.False(t, stub.query.Has("before"))
	})

	t.Run("--page-limit sets the page size", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		_, err := stub.run("my-trail --flow my-flow --page-limit 7")
		require.NoError(t, err)
		require.Equal(t, "7", stub.query.Get("per_page"))
	})

	for _, limit := range []string{"0", "101"} {
		t.Run("--page-limit "+limit+" fails without a request", func(t *testing.T) {
			stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
			out, err := stub.run("my-trail --flow my-flow --page-limit " + limit)
			require.Error(t, err)
			require.Contains(t, out, "Error: --page-limit must be between 1 and 100\n")
			require.Zero(t, stub.requests)
		})
	}

	t.Run("--after sends the cursor of the next page", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		_, err := stub.run("my-trail --flow my-flow --after abc=")
		require.NoError(t, err)
		require.Equal(t, "abc=", stub.query.Get("after"))
		require.False(t, stub.query.Has("before"))
	})

	t.Run("--before sends the cursor of the previous page", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		_, err := stub.run("my-trail --flow my-flow --before abc=")
		require.NoError(t, err)
		require.Equal(t, "abc=", stub.query.Get("before"))
		require.False(t, stub.query.Has("after"))
	})

	t.Run("--after with --before fails without a request", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, emptyTrailEventsPage)
		out, err := stub.run("my-trail --flow my-flow --after a --before b")
		require.Error(t, err)
		require.Contains(t, out, "Error: only one of --after, --before is allowed\n")
		require.Zero(t, stub.requests)
	})

	t.Run("the table ends with the cursors of the pages either side", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[{"type":"trail_reported","timestamp":1700000000}],"_links":{
			"self":{"href":"http://kosli/api/v2/trails/o/f/t/events?after=cur%3D"},
			"prev":{"href":"http://kosli/api/v2/trails/o/f/t/events?sort=newest&before=prv%3D"},
			"next":{"href":"http://kosli/api/v2/trails/o/f/t/events?sort=newest&after=nxt%3D"}}}`)
		out, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Equal(t, ""+
			"EVENT          TYPE   TIMESTAMP                      COMMIT  COMPLIANCE\n"+
			"Trail started  trail  Sat, 16 Jan 2016 00:00:00 UTC          \n"+
			"\n"+
			"Next page: --after nxt=\n"+
			"Previous page: --before prv=\n", out)
	})

	t.Run("an empty page still prints the cursor that leads back", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[],"_links":{
			"self":{"href":"http://kosli/api/v2/trails/o/f/t/events?after=cur%3D"},
			"prev":{"href":"http://kosli/api/v2/trails/o/f/t/events?before=prv%3D"}}}`)
		out, err := stub.run("my-trail --flow my-flow --after cur=")
		require.NoError(t, err)
		require.Equal(t, "No events were found.\n\nPrevious page: --before prv=\n", out)
	})

	t.Run("the API's message for an invalid cursor is printed", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusBadRequest, `{"message":"Invalid cursor."}`)
		out, err := stub.run("my-trail --flow my-flow --after nonsense")
		require.Error(t, err)
		require.Contains(t, out, "Invalid cursor.")
	})
}

const testTimestamp = "Sat, 16 Jan 2016 00:00:00 UTC"

func TestListTrailEventsTable(t *testing.T) {
	t.Run("the columns are those of the events list in the Kosli UI", func(t *testing.T) {
		stub := newTrailEventsStub(t, http.StatusOK, `{"events":[{"type":"trail_reported","timestamp":1700000000}],"_links":{}}`)
		out, err := stub.run("my-trail --flow my-flow")
		require.NoError(t, err)
		require.Equal(t, []string{"EVENT", "TYPE", "TIMESTAMP", "COMMIT", "COMPLIANCE"},
			strings.Fields(strings.SplitN(out, "\n", 2)[0]))
	})

	for _, tc := range []struct {
		name  string
		event string
		want  []string
	}{
		{
			name:  "a trail starting",
			event: `{"type":"trail_reported","timestamp":1700000000,"git_commit_info":{"sha1":"0fc1ba9876f91b215679f3649b8668085d820ab5"}}`,
			want:  []string{"Trail started", "trail", testTimestamp, "0fc1ba9", ""},
		},
		{
			name:  "a trail update without a commit",
			event: `{"type":"trail_updated","timestamp":1700000000}`,
			want:  []string{"Trail updated", "trail", testTimestamp, "", ""},
		},
		{
			name: "a compliant trail attestation",
			event: `{"type":"trail_attestation_reported","timestamp":1700000000,"template_reference_name":"bar",
				"attestation_type":"generic","is_compliant":true,"type_mismatch":false,
				"git_commit_info":{"sha1":"0fc1ba9876f91b215679f3649b8668085d820ab5"}}`,
			want: []string{"bar", "generic", testTimestamp, "0fc1ba9", "COMPLIANT"},
		},
		{
			name: "a non-compliant artifact attestation of a custom type",
			event: `{"type":"artifact_attestation_reported","timestamp":1700000000,"target_artifact":"cli",
				"template_reference_name":"scan","attestation_type":"custom:my-scan","is_compliant":false,"type_mismatch":false}`,
			want: []string{"cli.scan", "custom:my-scan", testTimestamp, "", "NON-COMPLIANT"},
		},
		{
			name: "an attestation for an artifact not yet reported, of a system type",
			event: `{"type":"trail_attestation_for_artifact_reported","timestamp":1700000000,"target_artifact":"cli",
				"template_reference_name":"tests","attestation_type":"system:junit","is_compliant":true,"type_mismatch":false}`,
			want: []string{"cli.tests", "junit", testTimestamp, "", "COMPLIANT"},
		},
		{
			name: "a compliant attestation of the wrong type",
			event: `{"type":"trail_attestation_reported","timestamp":1700000000,"template_reference_name":"bar",
				"attestation_type":"generic","is_compliant":true,"type_mismatch":true}`,
			want: []string{"bar", "generic", testTimestamp, "", "NON-COMPLIANT"},
		},
		{
			name: "an override of an artifact attestation",
			event: `{"type":"attestation_overridden","timestamp":1700000000,"target_artifact":"cli",
				"attestation_name":"scan","new_compliance_status":true}`,
			want: []string{"cli.scan", "override", testTimestamp, "", "COMPLIANT"},
		},
		{
			name:  "an override of a trail attestation to non-compliant",
			event: `{"type":"attestation_overridden","timestamp":1700000000,"attestation_name":"bar","new_compliance_status":false}`,
			want:  []string{"bar", "override", testTimestamp, "", "NON-COMPLIANT"},
		},
		{
			name:  "an artifact",
			event: `{"type":"artifact_creation_reported","timestamp":1700000000,"template_reference_name":"cli","artifact_name":"cli-image"}`,
			want:  []string{"cli", "artifact", testTimestamp, "", ""},
		},
		{
			name:  "a deployment",
			event: `{"type":"artifact_deployment_reported","timestamp":1700000000,"deployment_number":3,"environment_name":"production"}`,
			want:  []string{"Deployment #3 to production", "deployment", testTimestamp, "", ""},
		},
		{
			name:  "an approval that was reviewed",
			event: `{"type":"artifact_approval_reported","timestamp":1700000000,"approval_number":2,"reviewer":"sami"}`,
			want:  []string{"Approval #2 created by sami", "approval", testTimestamp, "", ""},
		},
		{
			name:  "an approval awaiting review",
			event: `{"type":"artifact_approval_reported","timestamp":1700000000,"approval_number":2}`,
			want:  []string{"Approval #2 requested", "approval", testTimestamp, "", ""},
		},
		{
			name: "an artifact starting to run",
			event: `{"type":"artifact_started_running","timestamp":1700000000,"template_reference_name":"cli",
				"environment_name":"labs-prod","snapshot_index":28}`,
			want: []string{"cli started running in labs-prod#28", "environment", testTimestamp, "", ""},
		},
		{
			name: "an artifact that stopped running",
			event: `{"type":"artifact_stopped_running","timestamp":1700000000,"template_reference_name":"cli",
				"environment_name":"labs-prod","snapshot_index":29}`,
			want: []string{"cli exited from labs-prod#29", "environment", testTimestamp, "", ""},
		},
		{
			name:  "an event type this CLI does not know",
			event: `{"type":"trail_archived","timestamp":1700000000}`,
			want:  []string{"trail_archived", "", testTimestamp, "", ""},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			event := trailEvent{}
			require.NoError(t, json.Unmarshal([]byte(tc.event), &event))
			row, err := event.row()
			require.NoError(t, err)
			require.Equal(t, tc.want, row)
		})
	}

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
			name: "listing events with --page-limit 1 returns one event and a link to the next page",
			cmd:  fmt.Sprintf(`list trail-events %s --flow %s --page-limit 1 --output json %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenJson: []jsonCheck{
				{"events", "length:1"},
				{"events.[0].type", "trail_attestation_reported"},
				{"_links.next.href", "not-nil"},
			},
		},
		{
			name:        "listing events with --page-limit 1 prints the cursor of the next page",
			cmd:         fmt.Sprintf(`list trail-events %s --flow %s --page-limit 1 %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "bar +generic +Sat, 16 Jan 2016 00:00:00 UTC +0fc1ba9 +COMPLIANT\n\nNext page: --after \\S+\n$",
		},
		{
			wantError:   true,
			name:        "listing events fails with an invalid cursor",
			cmd:         fmt.Sprintf(`list trail-events %s --flow %s --after nonsense %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "^Error: Invalid cursor\\.",
		},
		{
			wantError:   true,
			name:        "listing events fails with both cursors",
			cmd:         fmt.Sprintf(`list trail-events %s --flow %s --after a --before b %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "^Error: only one of --after, --before is allowed\n",
		},
		{
			wantError:   true,
			name:        "listing events fails with a --page-limit over 100",
			cmd:         fmt.Sprintf(`list trail-events %s --flow %s --page-limit 101 %s`, suite.trailName, suite.flowName, suite.defaultKosliArguments),
			goldenRegex: "^Error: --page-limit must be between 1 and 100\n",
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
