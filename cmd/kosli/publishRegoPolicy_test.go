package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/regopolicies"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	regoPolicyPending = `{"identifier":"pr-approved","version":1,"status":"pending",` +
		`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe"}`
	regoPolicyValid = `{"identifier":"pr-approved","version":1,"status":"valid",` +
		`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000002.25,` +
		`"opa_version":"1.21.1","entrypoint":"policy"}`
	regoPolicyRejected = `{"identifier":"pr-approved","version":1,"status":"rejected",` +
		`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000002.25,` +
		`"opa_version":"1.21.1","error":{"kind":"compile",` +
		`"message":"policy.rego:4: rego_parse_error: unexpected eof token"}}`
	regoPolicyFailed = `{"identifier":"pr-approved","version":1,"status":"failed",` +
		`"digest":"sha256:abc123","input_schema":"trail","comment":"",` +
		`"created_at":1757000000.5,"created_by":"Jane Doe","validated_at":1757000040.0,` +
		`"error":{"kind":"lambda_timeout","message":"validation timed out after 3 attempts"}}`
)

// fakeRegoPolicies answers version reads with the next of reads, repeating the last.
type fakeRegoPolicies struct {
	// Empty means the publish answers with a pending version.
	publishAnswer string
	published     []map[string]any
	publishURL    string
	reads         []string
	// Zero means a version read answers 200.
	readStatus int
	readCount  int
	readURL    string
	unexpected int
}

func newFakeRegoPolicies(t *testing.T, reads ...string) (*httptest.Server, *fakeRegoPolicies) {
	t.Helper()
	fake := &fakeRegoPolicies{reads: reads}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/rego-policies/"):
			raw, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			var body map[string]any
			require.NoError(t, json.Unmarshal(raw, &body))
			fake.published = append(fake.published, body)
			fake.publishURL = r.URL.Path
			w.WriteHeader(http.StatusAccepted)
			answer := fake.publishAnswer
			if answer == "" {
				answer = regoPolicyPending
			}
			_, _ = fmt.Fprint(w, answer)

		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/versions/"):
			fake.readURL = r.URL.Path
			status := fake.readStatus
			if status == 0 {
				status = http.StatusOK
			}
			w.WriteHeader(status)
			_, _ = fmt.Fprint(w, fake.reads[min(fake.readCount, len(fake.reads)-1)])
			fake.readCount++

		default:
			fake.unexpected++
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server, fake
}

type PublishRegoPolicyCommandTestSuite struct {
	suite.Suite
	originalWait regopolicies.WaitOptions
}

func (suite *PublishRegoPolicyCommandTestSuite) SetupTest() {
	suite.originalWait = regoPolicyWaitOptions
	regoPolicyWaitOptions = regopolicies.WaitOptions{
		Timeout: 2 * time.Second,
		Initial: time.Millisecond,
		Max:     2 * time.Millisecond,
	}
}

func (suite *PublishRegoPolicyCommandTestSuite) TearDownTest() {
	regoPolicyWaitOptions = suite.originalWait
}

func (suite *PublishRegoPolicyCommandTestSuite) publishCmd(host, args string) string {
	return fmt.Sprintf("publish rego-policy %s --host %s --org test-org --api-token test-token --max-api-retries 0",
		args, host)
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItIsListedInHelp() {
	_, combined, _, _, err := executeCommandC("--help")

	require.NoError(suite.T(), err)
	require.Contains(suite.T(), combined, publishDesc)
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItIsNoLongerACreateCommand() {
	root, err := newRootCmd(io.Discard, io.Discard, nil)
	require.NoError(suite.T(), err)
	create, _, err := root.Find([]string{"create"})
	require.NoError(suite.T(), err)

	for _, sub := range create.Commands() {
		require.NotEqual(suite.T(), "rego-policy", sub.Name())
	}
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItPublishesTheFileAndReportsTheValidVersion() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		`pr-approved --path testdata/policies/allow-all.rego --input-schema trail `+
			`--name "Pull request approved" --description "Approved PRs only" --comment "Initial version"`))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, fake.unexpected)
	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved", fake.publishURL)
	require.Len(suite.T(), fake.published, 1)

	published := fake.published[0]
	require.Equal(suite.T(), "trail", published["input_schema"])
	require.Equal(suite.T(), "Pull request approved", published["name"])
	require.Equal(suite.T(), "Approved PRs only", published["description"])
	require.Equal(suite.T(), "Initial version", published["comment"])
	files := published["files"].(map[string]any)
	require.Len(suite.T(), files, 1)
	require.Contains(suite.T(), files["allow-all.rego"], "package policy")

	require.Equal(suite.T(), "/api/v2/rego-policies/test-org/pr-approved/versions/1", fake.readURL)
	require.Contains(suite.T(), combined, "rego policy 'pr-approved' version 1 is valid (sha256:abc123)")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItLeavesOutTheNameAndDescriptionWhenNotGiven() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail"))

	require.NoError(suite.T(), err)
	require.NotContains(suite.T(), fake.published[0], "name")
	require.NotContains(suite.T(), fake.published[0], "description")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItWaitsThroughPendingReads() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyPending, regoPolicyPending, regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 3, fake.readCount)
	require.Equal(suite.T(), "rego policy 'pr-approved' version 1 is valid (sha256:abc123)\n", combined)
}

// The same directory must give the same digest here as in `evaluate policy`.
func (suite *PublishRegoPolicyCommandTestSuite) TestADirectoryTravelsKeyedByRelativePath() {
	dir := suite.T().TempDir()
	require.NoError(suite.T(), os.MkdirAll(filepath.Join(dir, "lib"), 0o755))
	require.NoError(suite.T(), os.WriteFile(filepath.Join(dir, "policy.rego"), []byte("package policy\n"), 0o644))
	require.NoError(suite.T(), os.WriteFile(filepath.Join(dir, "lib", "helpers.rego"), []byte("package lib\n"), 0o644))
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path "+dir+" --input-schema trails"))

	require.NoError(suite.T(), err)
	require.Equal(suite.T(), map[string]any{
		"policy.rego":      "package policy\n",
		"lib/helpers.rego": "package lib\n",
	}, fake.published[0]["files"])
	require.Equal(suite.T(), "trails", fake.published[0]["input_schema"])
}

// A policy and a library kept elsewhere are published as one version.
func (suite *PublishRegoPolicyCommandTestSuite) TestSeveralPathsJoinOneBundle() {
	lib := suite.T().TempDir()
	require.NoError(suite.T(), os.WriteFile(filepath.Join(lib, "helpers.rego"), []byte("package lib\n"), 0o644))
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego -p "+lib+" --input-schema trail"))

	require.NoError(suite.T(), err)
	require.Len(suite.T(), fake.published, 1)
	files := fake.published[0]["files"].(map[string]any)
	require.Len(suite.T(), files, 2)
	require.Contains(suite.T(), files["allow-all.rego"], "package policy")
	require.Equal(suite.T(), "package lib\n", files["helpers.rego"])
}

func (suite *PublishRegoPolicyCommandTestSuite) publishExpectingError(reads ...string) (string, error) {
	server, _ := newFakeRegoPolicies(suite.T(), reads...)
	_, combined, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail"))
	require.Error(suite.T(), err)
	require.NotContains(suite.T(), combined, "is valid")
	return combined, err
}

func (suite *PublishRegoPolicyCommandTestSuite) TestARejectedVersionNamesTheErrorInThePolicy() {
	_, err := suite.publishExpectingError(regoPolicyRejected)

	require.Equal(suite.T(),
		"rego policy 'pr-approved' version 1 was rejected (compile): "+
			"policy.rego:4: rego_parse_error: unexpected eof token",
		err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestARejectedVersionWithNoReasonSaysSo() {
	_, err := suite.publishExpectingError(
		strings.Replace(regoPolicyValid, `"status":"valid"`, `"status":"rejected"`, 1))

	require.Equal(suite.T(),
		"rego policy 'pr-approved' version 1 was rejected, and Kosli gave no reason",
		err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAFailedVersionIsNotBlamedOnThePolicy() {
	_, err := suite.publishExpectingError(regoPolicyFailed)

	require.Equal(suite.T(),
		"validation of rego policy 'pr-approved' version 1 failed on the Kosli side (lambda_timeout): "+
			"validation timed out after 3 attempts; "+
			"this is not a fault in the policy, publish again to retry",
		err.Error())
	require.NotContains(suite.T(), err.Error(), "rejected")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAFailedVersionWithNoReasonIsStillNotBlamedOnThePolicy() {
	_, err := suite.publishExpectingError(
		strings.Replace(regoPolicyValid, `"status":"valid"`, `"status":"failed"`, 1))

	require.Equal(suite.T(),
		"validation of rego policy 'pr-approved' version 1 failed on the Kosli side, with no reason given; "+
			"this is not a fault in the policy, publish again to retry",
		err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAnAnswerWithNoVersionIsNotReadBack() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)
	fake.publishAnswer = `{"identifier":"pr-approved","status":"pending"}`

	_, _, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail"))

	require.EqualError(suite.T(), err,
		"the Kosli server accepted the publish but named no version, so it cannot be read back")
	require.Equal(suite.T(), 0, fake.readCount)
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAFailedReadNamesTheVersionThatWasPublished() {
	for _, test := range []struct {
		name, output string
	}{
		{"table", "table"},
		{"json", "json"},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeRegoPolicies(suite.T(), `{"message":"Rego policy 'pr-approved' does not exist"}`)
			fake.readStatus = http.StatusNotFound

			_, _, stdout, _, err := executeCommandC(suite.publishCmd(server.URL,
				"pr-approved --path testdata/policies/allow-all.rego --input-schema trail --output "+test.output))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(),
				"rego policy 'pr-approved' version 1 was published, but could not be read back: ")
			require.Contains(suite.T(), err.Error(), "does not exist")
			if test.output == "json" {
				var printed map[string]any
				require.NoError(suite.T(), json.Unmarshal([]byte(stdout), &printed), "stdout is JSON: %s", stdout)
				require.Equal(suite.T(), "pending", printed["status"])
			} else {
				require.Empty(suite.T(), stdout)
			}
		})
	}
}

func (suite *PublishRegoPolicyCommandTestSuite) TestJsonOutputIsThePendingVersionWhenTheWaitExpires() {
	regoPolicyWaitOptions.Timeout = 20 * time.Millisecond
	server, _ := newFakeRegoPolicies(suite.T(), regoPolicyPending)

	_, _, stdout, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail --output json"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "still pending")
	var printed map[string]any
	require.NoError(suite.T(), json.Unmarshal([]byte(stdout), &printed), "stdout is JSON: %s", stdout)
	require.Equal(suite.T(), "pending", printed["status"])
	require.Equal(suite.T(), float64(1), printed["version"])
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAWaitThatExpiresSaysHowToReadTheVersionLater() {
	regoPolicyWaitOptions.Timeout = 20 * time.Millisecond

	_, err := suite.publishExpectingError(regoPolicyPending)

	require.Contains(suite.T(), err.Error(), "version 1 of rego policy pr-approved is still pending after")
	require.Regexp(suite.T(),
		`read it later with GET http://127\.0\.0\.1:\d+/api/v2/rego-policies/test-org/pr-approved/versions/1$`,
		err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) refusedBeforeSending(args string) error {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, _, _, err := executeCommandC(suite.publishCmd(server.URL, args))

	require.Error(suite.T(), err)
	require.Empty(suite.T(), fake.published, "nothing may be published")
	require.Equal(suite.T(), 0, fake.unexpected)
	return err
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAnUnknownInputSchemaIsRefused() {
	err := suite.refusedBeforeSending(
		"pr-approved --path testdata/policies/allow-all.rego --input-schema many")

	require.Equal(suite.T(), `--input-schema must be one of [trail trails], got "many"`, err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAPolicyAtAURLIsRefused() {
	err := suite.refusedBeforeSending(
		"pr-approved --path https://example.com/policy.rego --input-schema trail")

	require.Equal(suite.T(), "--path takes a file or a directory on this machine, not a URL", err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAPathInsideAnotherPathIsRefused() {
	err := suite.refusedBeforeSending(
		"pr-approved --path testdata/policies --path testdata/policies/allow-all.rego --input-schema trail")

	require.Equal(suite.T(),
		"--path testdata/policies/allow-all.rego is inside --path testdata/policies, so its files would be sent twice",
		err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestABundleOverTheFileCapIsRefused() {
	dir := suite.T().TempDir()
	for i := range maxPolicyBundleFiles + 1 {
		name := filepath.Join(dir, fmt.Sprintf("p%03d.rego", i))
		require.NoError(suite.T(), os.WriteFile(name, []byte("package p\n"), 0o644))
	}

	err := suite.refusedBeforeSending("pr-approved --path " + dir + " --input-schema trail")

	require.Equal(suite.T(), "policy bundle holds 101 files, over the limit of 100", err.Error())
}

// An unset variable in `--description "$DESC"` must not clear a description.
func (suite *PublishRegoPolicyCommandTestSuite) TestAnEmptyDescriptionIsRefused() {
	err := suite.refusedBeforeSending(
		`pr-approved --path testdata/policies/allow-all.rego --input-schema trail --description ""`)

	require.Contains(suite.T(), err.Error(), "flag '--description' was given an empty value")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestJsonOutputIsPrintedForAVersionThatIsNotValid() {
	for _, test := range []struct {
		name, body, status, kind string
	}{
		{"rejected", regoPolicyRejected, "rejected", "compile"},
		{"failed", regoPolicyFailed, "failed", "lambda_timeout"},
	} {
		suite.Run(test.name, func() {
			server, _ := newFakeRegoPolicies(suite.T(), test.body)

			_, _, stdout, _, err := executeCommandC(suite.publishCmd(server.URL,
				"pr-approved --path testdata/policies/allow-all.rego --input-schema trail --output json"))

			require.Error(suite.T(), err)
			require.Contains(suite.T(), err.Error(), test.status)
			var printed map[string]any
			require.NoError(suite.T(), json.Unmarshal([]byte(stdout), &printed), "stdout is JSON: %s", stdout)
			require.Equal(suite.T(), test.status, printed["status"])
			require.Equal(suite.T(), test.kind, printed["error"].(map[string]any)["kind"])
		})
	}
}

func (suite *PublishRegoPolicyCommandTestSuite) TestTableOutputPrintsNothingForAVersionThatIsNotValid() {
	server, _ := newFakeRegoPolicies(suite.T(), regoPolicyRejected)

	_, _, stdout, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail"))

	require.Error(suite.T(), err)
	require.Empty(suite.T(), stdout)
}

func (suite *PublishRegoPolicyCommandTestSuite) TestJsonOutputIsTheValidVersion() {
	server, _ := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, _, stdout, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail --output json"))
	require.NoError(suite.T(), err)

	var printed map[string]any
	require.NoError(suite.T(), json.Unmarshal([]byte(stdout), &printed), "stdout is JSON: %s", stdout)
	require.Equal(suite.T(), map[string]any{
		"identifier":   "pr-approved",
		"version":      float64(1),
		"status":       "valid",
		"digest":       "sha256:abc123",
		"input_schema": "trail",
		"comment":      "",
		"created_at":   1757000000.5,
		"created_by":   "Jane Doe",
		"validated_at": 1757000002.25,
		"opa_version":  "1.21.1",
		"entrypoint":   "policy",
	}, printed, "fields the server left out stay out")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestAnUnknownOutputFormatIsRefused() {
	err := suite.refusedBeforeSending(
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail --output yaml")

	require.Equal(suite.T(), "unsupported output format: yaml. Valid formats are: [table, json]", err.Error())
}

func (suite *PublishRegoPolicyCommandTestSuite) TestADryRunSendsNothingAndWaitsForNothing() {
	server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)

	_, combined, _, _, err := executeCommandC(suite.publishCmd(server.URL,
		"pr-approved --path testdata/policies/allow-all.rego --input-schema trail --dry-run"))

	require.NoError(suite.T(), err)
	require.Empty(suite.T(), fake.published)
	require.Equal(suite.T(), 0, fake.readCount)
	require.Contains(suite.T(), combined, "THIS IS A DRY-RUN")
	require.Contains(suite.T(), combined, "/api/v2/rego-policies/test-org/pr-approved")
	require.NotContains(suite.T(), combined, "is valid")
}

// A new version keeps the comment it was published with, so a different
// comment on the answer means the server stored nothing.
func (suite *PublishRegoPolicyCommandTestSuite) TestItSaysWhenAPublishStoredNothingNew() {
	const unchanged = "rego policy 'pr-approved' version 1 is valid (sha256:abc123); nothing new was stored, " +
		"as the files and input schema are the same as this version"
	withComment := func(body, comment string) string {
		return strings.Replace(body, `"comment":""`, `"comment":"`+comment+`"`, 1)
	}
	for _, test := range []struct {
		name, answer, flags, want string
		reads                     int
	}{
		{"valid, no comment given", regoPolicyValid, "",
			unchanged + "\n", 0},
		{"valid, same comment", withComment(regoPolicyValid, "First"), ` --comment "First"`,
			unchanged + "\n", 0},
		{"valid, other comment", regoPolicyValid, ` --comment "Second try"`,
			unchanged + ", so --comment was not kept\n", 0},
		{"pending, other comment", withComment(regoPolicyPending, "First"), ` --comment "Second try"`,
			unchanged + ", so --comment was not kept\n", 1},
		{"new version, same comment", withComment(regoPolicyPending, "Second try"), ` --comment "Second try"`,
			"rego policy 'pr-approved' version 1 is valid (sha256:abc123)\n", 1},
	} {
		suite.Run(test.name, func() {
			server, fake := newFakeRegoPolicies(suite.T(), regoPolicyValid)
			fake.publishAnswer = test.answer

			_, combined, _, _, err := executeCommandC(suite.publishCmd(server.URL,
				"pr-approved --path testdata/policies/allow-all.rego --input-schema trail"+test.flags))

			require.NoError(suite.T(), err)
			require.Equal(suite.T(), test.reads, fake.readCount, "a terminal answer is not polled")
			require.Equal(suite.T(), test.want, combined)
		})
	}
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItNeedsAnIdentifier() {
	_, _, _, _, err := executeCommandC(suite.publishCmd("http://localhost:1",
		"--path testdata/policies/allow-all.rego --input-schema trail"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "accepts 1 arg(s), received 0")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItNeedsAPath() {
	_, _, _, _, err := executeCommandC(suite.publishCmd("http://localhost:1",
		"pr-approved --input-schema trail"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), `required flag(s) "path" not set`)
}

// The path was once the second argument; it must not be taken as one now.
func (suite *PublishRegoPolicyCommandTestSuite) TestAPathGivenAsAnArgumentIsRefused() {
	_, _, _, _, err := executeCommandC(suite.publishCmd("http://localhost:1",
		"pr-approved testdata/policies/allow-all.rego --input-schema trail"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), "accepts 1 arg(s), received 2")
}

func (suite *PublishRegoPolicyCommandTestSuite) TestItNeedsAnInputSchema() {
	_, _, _, _, err := executeCommandC(suite.publishCmd("http://localhost:1",
		"pr-approved --path testdata/policies/allow-all.rego"))

	require.Error(suite.T(), err)
	require.Contains(suite.T(), err.Error(), `required flag(s) "input-schema" not set`)
}

func TestPublishRegoPolicyCommandTestSuite(t *testing.T) {
	suite.Run(t, new(PublishRegoPolicyCommandTestSuite))
}
