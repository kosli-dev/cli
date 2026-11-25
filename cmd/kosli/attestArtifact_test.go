package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type AttestArtifactCommandTestSuite struct {
	flowName  string
	trailName string
	suite.Suite
	defaultKosliArguments string
	builtArtifact         string
}

func (suite *AttestArtifactCommandTestSuite) SetupTest() {
	suite.flowName = "attest-artifact"
	suite.trailName = "test-123"
	suite.builtArtifact = suite.T().TempDir() + "/built.bin"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Org:      "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --flow %s --trail %s --repo-root ../.. --host %s --org %s --api-token %s", suite.flowName, suite.trailName, global.Host, global.Org, global.ApiToken)
	CreateFlowWithTemplate(suite.flowName, "testdata/valid_template.yml", suite.T())
	BeginTrail(suite.trailName, suite.flowName, "", suite.T())
}

func (suite *AttestArtifactCommandTestSuite) TestAttestArtifactCmd() {
	tests := []cmdTestCase{
		{
			wantError: true,
			name:      "fails when more arguments are provided",
			cmd:       fmt.Sprintf("attest artifact foo bar %s", suite.defaultKosliArguments),
			golden:    "Error: accepts at most 1 arg(s), received 2 [foo bar]\n",
		},
		{
			wantError: true,
			name:      "fails when missing a required flag",
			cmd:       fmt.Sprintf("attest artifact foo --artifact-type file --name bar --commit HEAD --build-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: required flag(s) \"commit-url\" not set\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is invalid sha256 digest",
			cmd:       fmt.Sprintf("attest artifact foo --fingerprint xxxx --name bar --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: xxxx is not a valid SHA256 fingerprint. It should match the pattern ^([a-f0-9]{64})$\nUsage: kosli attest artifact {IMAGE-NAME | FILE-PATH | DIR-PATH} [-- BUILD-COMMAND...] [flags]\n",
		},
		{
			name:   "works when --name does not match artifact name in the template (extra artifact)",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --name bar --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden: "artifact file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest a file artifact",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden: "artifact file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest a file artifact named after --",
			cmd:    fmt.Sprintf("attest artifact --artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com %s -- testdata/file1", suite.defaultKosliArguments),
			golden: "artifact file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest a file artifact built by a build command",
			cmd:    fmt.Sprintf("attest artifact %s --artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com %s -- cp testdata/file1 %s", suite.builtArtifact, suite.defaultKosliArguments, suite.builtArtifact),
			golden: "artifact built.bin was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest an artifact with --fingerprint",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest an artifact with external urls",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --external-url jira=https://jira.kosli.com  %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest an artifact with external urls and fingerprints",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --external-url file=https://kosli.com/file --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9  %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			wantError: true,
			name:      "fails when --external-fingerprint has more items than external urls",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --external-fingerprint file=7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9  %s", suite.defaultKosliArguments),
			golden:    "Error: --external-fingerprints have labels that don't have a URL in --external-url\n",
		},
		{
			wantError:   true,
			name:        "fails (from server) when --external-fingerprint has invalid fingerprint",
			cmd:         fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --external-url file=https://http://www.example.com --external-fingerprint file=7509e5bda0  %s", suite.defaultKosliArguments),
			goldenRegex: "Error: Input payload validation failed: .*7509e5bda0",
		},
		{
			name:   "can attest with annotations against a trail",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --annotate foo=bar --annotate baz=\"data with spaces\" %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			wantError: true,
			name:      "fails when annotation is not valid",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --annotate foo.baz=bar %s", suite.defaultKosliArguments),
			golden:    "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.baz'. Key can only contain [A-Za-z0-9_]\n",
		},
		{
			name:   "can attest a file artifact with redacted commit info",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --redact-commit-info author,branch --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden: "artifact file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			wantError: true,
			name:      "fails when attesting an artifact with invalid redacted commit info",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --artifact-type file --redact-commit-info author,bar --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com  %s", suite.defaultKosliArguments),
			golden:    "Error: bar is not an allowed value for --redact-commit-info\n",
		},
		{
			wantError: true,
			name:      "fails when --repo-url is not a valid URL",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --repo-url not-a-url %s", suite.defaultKosliArguments),
			golden:    "Error: --repo-url 'not-a-url' is not a valid URL\n",
		},
		{
			wantError: true,
			name:      "fails when --repo-provider is not an allowed value",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --repo-provider jenkins %s", suite.defaultKosliArguments),
			golden:    "Error: --repo-provider 'jenkins' is not allowed. Must be one of: github, gitlab, bitbucket, bitbucket_cloud, bitbucket_dc, azure-devops, azure_devops_services, azure_devops_server, git, subversion\n",
		},
		{
			name:   "can attest with all repo flags",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --repo-id test-repo-id --repository test-repo-name --repo-url https://github.com/org/repo --repo-provider github %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
		{
			name:   "can attest without repo-id and repository",
			cmd:    fmt.Sprintf("attest artifact testdata/file1 --fingerprint 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9 --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com %s", suite.defaultKosliArguments),
			golden: "artifact testdata/file1 was attested with fingerprint: 7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

func TestSplitArtifactAndBuildArgs(t *testing.T) {
	for _, tt := range []struct {
		name         string
		dash         int
		args         []string
		wantArtifact []string
		wantBuild    []string
		wantErr      string
	}{
		{name: "no dash keeps all args as artifact args", dash: -1, args: []string{"app"}, wantArtifact: []string{"app"}},
		{name: "single arg after leading dash is the artifact", dash: 0, args: []string{"-odd.tgz"}, wantArtifact: []string{"-odd.tgz"}},
		{name: "args after dash are the build command", dash: 1, args: []string{"app", "go", "build"}, wantArtifact: []string{"app"}, wantBuild: []string{"go", "build"}},
		{name: "several args after leading dash are the build command", dash: 0, args: []string{"go", "build"}, wantArtifact: []string{}, wantBuild: []string{"go", "build"}},
		{name: "trailing dash is an error", dash: 1, args: []string{"app"}, wantErr: "no build command given after --"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			artifactArgs, buildCmd, err := splitArtifactAndBuildArgs(tt.dash, tt.args)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantArtifact, artifactArgs)
			require.Equal(t, tt.wantBuild, buildCmd)
		})
	}
}

// TestAttestArtifactBuildCommandValidation covers cases that fail or stop at
// --dry-run before reaching the server.
func TestAttestArtifactBuildCommandValidation(t *testing.T) {
	sha := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	baseArgs := "--name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --flow attest-artifact --trail test-123 --repo-root ../.. --org docs-cmd-test-user"
	singleHost := baseArgs + " --host http://localhost:8001 --api-token secret-token"
	tests := []cmdTestCase{
		{
			name:        "a single arg after -- is still the artifact name",
			cmd:         fmt.Sprintf("attest artifact --artifact-type file %s --dry-run -- testdata/file1", singleHost),
			goldenRegex: fmt.Sprintf(`"fingerprint": "%s"`, sha),
		},
		{
			wantError: true,
			name:      "fails when -- is not followed by a build command",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --artifact-type file %s --", singleHost),
			golden:    "Error: no build command given after --\n",
		},
		{
			wantError: true,
			name:      "fails when --fingerprint is combined with a build command",
			cmd:       fmt.Sprintf("attest artifact testdata/file1 --fingerprint %s %s -- true", sha, singleHost),
			golden:    "Error: --fingerprint cannot be combined with a build command\n",
		},
	}
	runTestCmd(t, tests)
}

// TestAttestArtifactBuildCommandDryRun runs build commands with --dry-run, so
// no server is needed.
func TestAttestArtifactBuildCommandDryRun(t *testing.T) {
	t.Setenv("KOSLI_TEST_SECRET_TOKEN", "s3cr3t-value")
	sha := "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"
	dir := t.TempDir()
	out := dir + "/out.bin"
	missing := dir + "/missing.bin"
	unchanged := dir + "/unchanged.bin"
	require.NoError(t, os.WriteFile(unchanged, []byte("same"), 0o644))
	shim := dir + "/shim.txt"
	args := "--artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --flow attest-artifact --trail test-123 --repo-root ../.. --host http://localhost:8001 --org docs-cmd-test-user --api-token secret-token --dry-run"

	tests := []cmdTestCase{
		{
			name:        "attests the artifact built by the build command",
			cmd:         fmt.Sprintf("attest artifact %s %s -- cp testdata/file1 %s", out, args, out),
			goldenRegex: fmt.Sprintf(`"fingerprint": "%s"(.|\n)*"build_command": "cp testdata/file1 %s"(.|\n)*"build_duration_seconds": "\d+\.\d"`, sha, out),
		},
		{
			wantError: true,
			name:      "fails with the build's exit code and does not attest",
			cmd:       fmt.Sprintf("attest artifact %s %s -- sh -c 'exit 3'", out, args),
			golden:    "Error: build command failed with exit code 3\n",
		},
		{
			wantError:   true,
			name:        "fails when the build succeeds but the artifact is missing",
			cmd:         fmt.Sprintf("attest artifact %s %s -- true", missing, args),
			goldenRegex: fmt.Sprintf(`^Error: build command succeeded but artifact %s was not found: .*no such file or directory\n$`, missing),
		},
		{
			name:        "warns when the build leaves the artifact unchanged",
			cmd:         fmt.Sprintf("attest artifact %s %s -- true", unchanged, args),
			goldenRegex: fmt.Sprintf(`^\[warning\] artifact %s was not modified by the build command\n(.|\n)*"fingerprint"`, unchanged),
		},
		{
			name:        "does not check whether a dir artifact was modified",
			cmd:         fmt.Sprintf("attest artifact %s %s -- true", dir, strings.Replace(args, "--artifact-type file", "--artifact-type dir", 1)),
			goldenRegex: `^(\[warning\] Repo information[^\n]*\n)?###############`,
		},
		{
			name:        "masks secrets in the build command",
			cmd:         fmt.Sprintf("attest artifact %s %s -- sh -c 'cp testdata/file1 %s' s3cr3t-value secret-token", out, args, out),
			goldenRegex: fmt.Sprintf(`"build_command": "sh -c cp testdata/file1 %s \*\*\* \*\*\*"`, out),
		},
		{
			name:        "user annotations win over build annotations",
			cmd:         fmt.Sprintf("attest artifact %s %s --annotate build_command=custom -- cp testdata/file1 %s", out, args, out),
			goldenRegex: `"build_command": "custom"`,
		},
		{
			name:        "--record-build-command=false omits the build command",
			cmd:         fmt.Sprintf("attest artifact %s %s --record-build-command=false -- cp testdata/file1 %s", out, args, out),
			goldenRegex: `"annotations": \{\s*"build_duration_seconds": "\d+\.\d"\s*\}`,
		},
		{
			name:        "the build command runs with KOSLI_SHIM_DISABLED=1",
			cmd:         fmt.Sprintf(`attest artifact %s %s -- sh -c 'echo "$KOSLI_SHIM_DISABLED" > %s'`, shim, args, shim),
			goldenRegex: `"fingerprint": "4355a46b19d348dc2f57c046f8ef63d4538ebb936000f3c9ee954a27460dd865"`,
		},
	}
	runTestCmd(t, tests)
}

// TestAttestArtifactMissingArtifactFailsUnderDryRun drives innerMain, which the
// golden harness bypasses and which turns other errors into exit 0 under --dry-run.
func TestAttestArtifactMissingArtifactFailsUnderDryRun(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.bin")
	args := []string{
		"attest", "artifact", missing, "--artifact-type", "file", "--name", "cli",
		"--commit", "HEAD", "--build-url", "http://www.example.com", "--commit-url", "http://www.example.com",
		"--flow", "attest-artifact", "--trail", "test-123", "--repo-root", "../..",
		"--host", "http://localhost:8001", "--org", "docs-cmd-test-user", "--api-token", "secret-token",
		"--dry-run", "--", "true",
	}
	cmd, err := newRootCmd(io.Discard, io.Discard, args)
	require.NoError(t, err)

	code, ok := buildExitCode(innerMain(cmd, append([]string{"kosli"}, args...)))
	require.True(t, ok)
	require.Equal(t, 1, code)
}

// TestAttestArtifactFailOnAttestError sends the attestation to a port nothing
// listens on, so the Kosli request fails without a server.
func TestAttestArtifactFailOnAttestError(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.bin")
	annotated := filepath.Join(t.TempDir(), "annotated.bin")
	args := "--artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --flow attest-artifact --trail test-123 --repo-root ../.. --host http://localhost:1 --max-api-retries 0 --org docs-cmd-test-user --api-token secret-token"

	tests := []cmdTestCase{
		{
			name:        "warns instead of failing when the build succeeded and the flag is false",
			cmd:         fmt.Sprintf("attest artifact %s %s --fail-on-attest-error=false -- cp testdata/file1 %s", out, args, out),
			goldenRegex: `\[warning\] the build succeeded but the attestation was not recorded: .*connection refused`,
		},
		{
			wantError:   true,
			name:        "fails on a failed Kosli request by default",
			cmd:         fmt.Sprintf("attest artifact %s %s -- cp testdata/file1 %s", out, args, out),
			goldenRegex: `Error: .*connection refused`,
		},
		{
			wantError:   true,
			name:        "still fails without a build command",
			cmd:         fmt.Sprintf("attest artifact testdata/file1 %s --fail-on-attest-error=false", args),
			goldenRegex: `Error: .*connection refused`,
		},
		{
			wantError: true,
			name:      "still fails on invalid input after the build",
			cmd:       fmt.Sprintf("attest artifact %s %s --fail-on-attest-error=false --annotate foo.bar=baz -- cp testdata/file1 %s", annotated, args, annotated),
			golden:    "Error: --annotate flag should be in the format key=value. Invalid key: 'foo.bar'. Key can only contain [A-Za-z0-9_]\n",
		},
		{
			wantError: true,
			name:      "still fails with the build's exit code when the build fails",
			cmd:       fmt.Sprintf("attest artifact %s %s --fail-on-attest-error=false -- sh -c 'exit 3'", out, args),
			golden:    "Error: build command failed with exit code 3\n",
		},
	}
	runTestCmd(t, tests)
}

// TestAttestArtifactFailOnAttestErrorByStatus checks which Kosli responses
// --fail-on-attest-error=false softens: an outage does, a setup mistake does not.
func TestAttestArtifactFailOnAttestErrorByStatus(t *testing.T) {
	kosliAnswering := func(status int) string {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			_, _ = fmt.Fprintf(w, `{"message": "status %d from the stub"}`, status)
		}))
		t.Cleanup(server.Close)
		return server.URL
	}
	out := filepath.Join(t.TempDir(), "out.bin")
	args := "--artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --flow attest-artifact --trail test-123 --repo-root ../.. --max-api-retries 0 --org docs-cmd-test-user --api-token secret-token --fail-on-attest-error=false"

	tests := []cmdTestCase{
		{
			name:        "warns on a server error",
			cmd:         fmt.Sprintf("attest artifact %s %s --host %s -- cp testdata/file1 %s", out, args, kosliAnswering(http.StatusServiceUnavailable), out),
			goldenRegex: `\[warning\] the build succeeded but the attestation was not recorded: .*503`,
		},
		{
			name:        "warns on rate limiting",
			cmd:         fmt.Sprintf("attest artifact %s %s --host %s -- cp testdata/file1 %s", out, args, kosliAnswering(http.StatusTooManyRequests), out),
			goldenRegex: `\[warning\] the build succeeded but the attestation was not recorded: .*429`,
		},
		{
			name:        "warns on a lock conflict",
			cmd:         fmt.Sprintf("attest artifact %s %s --host %s -- cp testdata/file1 %s", out, args, kosliAnswering(http.StatusConflict), out),
			goldenRegex: `\[warning\] the build succeeded but the attestation was not recorded: .*409`,
		},
		{
			wantError:   true,
			name:        "still fails on an invalid token",
			cmd:         fmt.Sprintf("attest artifact %s %s --host %s -- cp testdata/file1 %s", out, args, kosliAnswering(http.StatusUnauthorized), out),
			goldenRegex: `Error: status 401 from the stub`,
		},
		{
			wantError:   true,
			name:        "still fails on an unknown flow",
			cmd:         fmt.Sprintf("attest artifact %s %s --host %s -- cp testdata/file1 %s", out, args, kosliAnswering(http.StatusNotFound), out),
			goldenRegex: `Error: status 404 from the stub`,
		},
	}
	runTestCmd(t, tests)
}

// fakeKosliHost records what each attestation request carried.
type fakeKosliHost struct {
	url      string
	mu       sync.Mutex
	tokens   []string
	clients  []string
	payloads []string
}

func (h *fakeKosliHost) requests() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.tokens)
}

func newFakeKosliHost(t *testing.T, status int) *fakeKosliHost {
	host := &fakeKosliHost{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		host.mu.Lock()
		host.tokens = append(host.tokens, r.Header.Get("Authorization"))
		host.clients = append(host.clients, r.Header.Get("X-Kosli-Client"))
		host.payloads = append(host.payloads, string(body))
		host.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = fmt.Fprintf(w, `{"message": "status %d from %s"}`, status, r.Host)
	}))
	t.Cleanup(server.Close)
	host.url = server.URL
	return host
}

// TestAttestArtifactBuildCommandMultiHost runs one build and attests it to every
// host in a comma-separated --host/--api-token list.
func TestAttestArtifactBuildCommandMultiHost(t *testing.T) {
	run := func(t *testing.T, first, second *fakeKosliHost, extraFlags string) (output string, failed bool) {
		dir := t.TempDir()
		out := filepath.Join(dir, "out.bin")
		builds := filepath.Join(dir, "builds.txt")
		cmd := fmt.Sprintf("attest artifact %s --artifact-type file --name cli --commit HEAD --build-url http://www.example.com --commit-url http://www.example.com --flow attest-artifact --trail test-123 --repo-root ../.. --org docs-cmd-test-user --max-api-retries 0 --host %s,%s --api-token token-one,token-two %s -- sh -c 'echo build >> %s; cp testdata/file1 %s' token-one token-two",
			out, first.url, second.url, extraFlags, builds, out)
		_, output, _, _, err := executeCommandC(cmd)
		buildLog, readErr := os.ReadFile(builds)
		require.NoError(t, readErr)
		require.Equal(t, "build\n", string(buildLog), "the build must run exactly once")
		return output, err != nil
	}

	t.Run("attests to every host with its own token", func(t *testing.T) {
		first, second := newFakeKosliHost(t, http.StatusCreated), newFakeKosliHost(t, http.StatusCreated)
		output, failed := run(t, first, second, "")
		require.False(t, failed, output)
		require.Equal(t, []string{"Bearer token-one"}, first.tokens)
		require.Equal(t, []string{"Bearer token-two"}, second.tokens)
		require.NotEmpty(t, first.clients[0], "usage headers must reach the first host")
		require.NotEmpty(t, second.clients[0], "usage headers must reach the second host")
		require.Equal(t, 1, strings.Count(output, "was attested with fingerprint"), output)
		for _, payload := range append(first.payloads, second.payloads...) {
			require.Contains(t, payload, `"build_command":"sh -c echo build`)
			require.NotContains(t, payload, "token-one")
			require.NotContains(t, payload, "token-two")
		}
	})

	t.Run("attests to the other hosts and names the one that failed", func(t *testing.T) {
		first, second := newFakeKosliHost(t, http.StatusCreated), newFakeKosliHost(t, http.StatusServiceUnavailable)
		output, failed := run(t, first, second, "")
		require.True(t, failed)
		require.Equal(t, 1, first.requests())
		require.Equal(t, 1, second.requests())
		require.Contains(t, output, "["+second.url+"]")
		require.NotContains(t, output, "["+first.url+"]")
	})

	t.Run("warns about a host that is down with --fail-on-attest-error=false", func(t *testing.T) {
		first, second := newFakeKosliHost(t, http.StatusServiceUnavailable), newFakeKosliHost(t, http.StatusCreated)
		output, failed := run(t, first, second, "--fail-on-attest-error=false")
		require.False(t, failed, output)
		require.Regexp(t, `\[warning\] \[`+regexp.QuoteMeta(first.url)+`\] the build succeeded but the attestation was not recorded`, output)
		require.Equal(t, 1, second.requests())
	})

	t.Run("treats a host whose token is DRY_RUN as a dry run", func(t *testing.T) {
		first, second := newFakeKosliHost(t, http.StatusUnauthorized), newFakeKosliHost(t, http.StatusCreated)
		output, failed := run(t, first, second, "--api-token DRY_RUN,token-two")
		require.False(t, failed, output)
		require.Equal(t, 0, first.requests(), "a host with a DRY_RUN token must not be sent a request")
		require.Equal(t, []string{"Bearer token-two"}, second.tokens)
		require.Contains(t, output, "THIS IS A DRY-RUN")
	})

	t.Run("still fails on a 4xx from one host with --fail-on-attest-error=false", func(t *testing.T) {
		first, second := newFakeKosliHost(t, http.StatusCreated), newFakeKosliHost(t, http.StatusUnauthorized)
		output, failed := run(t, first, second, "--fail-on-attest-error=false")
		require.True(t, failed)
		require.Contains(t, output, "["+second.url+"]")
	})
}

func TestMaskSecrets(t *testing.T) {
	for _, tt := range []struct {
		name    string
		cmdline string
		environ []string
		extra   []string
		want    string
	}{
		{name: "masks values of secret-named env vars", cmdline: "deploy --token abcdef123", environ: []string{"MY_TOKEN=abcdef123"}, want: "deploy --token ***"},
		{name: "matches names case-insensitively", cmdline: "login hunter22", environ: []string{"db_password=hunter22"}, want: "login ***"},
		{name: "ignores other env vars", cmdline: "build abcdef123", environ: []string{"MY_VALUE=abcdef123"}, want: "build abcdef123"},
		{name: "ignores values shorter than 6 characters", cmdline: "build abc", environ: []string{"MY_KEY=abc"}, want: "build abc"},
		{name: "ignores empty values", cmdline: "build", environ: []string{"MY_KEY="}, extra: []string{""}, want: "build"},
		{name: "masks the longer of two overlapping secrets whole", cmdline: "run abcdef-extra", environ: []string{"A_KEY=abcdef", "B_KEY=abcdef-extra"}, want: "run ***"},
		{name: "always masks extra values", cmdline: "run tok", extra: []string{"tok"}, want: "run ***"},
		{name: "masks other common secret names", cmdline: "push npm-auth-1 gh-pat-1 db-pass-1 pem-private-1 tls-cert-1",
			environ: []string{"NPM_AUTH=npm-auth-1", "GH_PAT=gh-pat-1", "DB_PASS=db-pass-1", "SIGNING_PRIVATE=pem-private-1", "CLIENT_CERT=tls-cert-1"},
			want:    "push *** *** *** *** ***"},
		{name: "does not treat PATH as a personal access token", cmdline: "run /usr/local/bin", environ: []string{"PATH=/usr/local/bin"}, want: "run /usr/local/bin"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, maskSecrets(tt.cmdline, tt.environ, tt.extra...))
		})
	}
}

func TestAddBuildAnnotations(t *testing.T) {
	t.Run("adds build command and duration", func(t *testing.T) {
		got := addBuildAnnotations(nil, "go build", 1250*time.Millisecond, true)
		require.Equal(t, map[string]string{"build_command": "go build", "build_duration_seconds": "1.2"}, got)
	})

	t.Run("keeps user annotations", func(t *testing.T) {
		got := addBuildAnnotations(map[string]string{"build_command": "custom"}, "go build", time.Second, true)
		require.Equal(t, map[string]string{"build_command": "custom", "build_duration_seconds": "1.0"}, got)
	})

	t.Run("omits the build command when not recorded", func(t *testing.T) {
		got := addBuildAnnotations(map[string]string{}, "go build", time.Second, false)
		require.Equal(t, map[string]string{"build_duration_seconds": "1.0"}, got)
	})
}

// TestAttestArtifactPayload_RepoInfoOmittedWhenNil ensures that when GitRepoInfo
// is not available (nil), the JSON payload does not include the repo_info field
// (omitempty behavior).
func TestAttestArtifactPayload_RepoInfoOmittedWhenNil(t *testing.T) {
	payload := AttestArtifactPayload{
		Fingerprint: "abc123",
		Filename:    "file1",
		GitCommit:   "sha",
		BuildUrl:    "https://build.example.com",
		CommitUrl:   "https://commit.example.com",
		RepoUrl:     "https://repo.example.com",
		Name:        "cli",
		TrailName:   "trail-1",
		GitRepoInfo: nil, // not available
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	if strings.Contains(string(data), "repo_info") {
		t.Errorf("payload must not include repo_info when GitRepoInfo is nil, got: %s", string(data))
	}
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestAttestArtifactCommandTestSuite(t *testing.T) {
	suite.Run(t, new(AttestArtifactCommandTestSuite))
}
