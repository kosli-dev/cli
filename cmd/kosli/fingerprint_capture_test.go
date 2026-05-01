package main

import (
	"testing"

	"github.com/kosli-dev/cli/internal/docker"
	"github.com/kosli-dev/cli/internal/version"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FingerprintCaptureTestSuite asserts the customer-facing contract that
// `kosli fingerprint` produces shell-capturable output: stdout is exactly
// the fingerprint, stderr is exactly empty. This is the contract Kosli
// users rely on when they write `FP=$(kosli fingerprint ... 2>&1)` in CI.
//
// The contract must hold even when the CLI has internal reasons to want
// to write to stderr (e.g. an outdated-version notice). Any future code
// path that writes to stderr from a fingerprint invocation breaks this
// contract and breaks customer pipelines, regardless of cause.
type FingerprintCaptureTestSuite struct {
	suite.Suite
	dockerImage string
}

const (
	// SHA256 of cmd/kosli/testdata/file1, which contains "hello world!".
	file1Fingerprint = "7509e5bda0c762d2bac7f90d758b5b2263fa01ccbc542ab5e3df163be08e6ca9"

	// SHA256 of cmd/kosli/testdata/folder1, pinned in fingerprint_test.go.
	folder1Fingerprint = "c43808cb04c6e66c4c6fc1f972dd67c3b9b71c81e0a0c78730da3699922d17be"

	// Realistic notice the version-check goroutine emits when a newer
	// release exists. Stubbed in via SetCheckForUpdateOverride.
	fakeUpdateNotice = "\nA new version of the Kosli CLI is available: v9.99.0 (you have v0.0.1)\nUpgrade: https://docs.kosli.com/getting_started/install/\n"
)

// SetupSuite pulls the alpine test image used by the docker variant. Same
// pattern as FingerprintTestSuite in fingerprint_test.go — the image is
// pinned by digest so the assertion can be exact.
func (suite *FingerprintCaptureTestSuite) SetupSuite() {
	suite.dockerImage = "library/alpine@sha256:e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"
	err := docker.PullDockerImage(suite.dockerImage)
	require.NoError(suite.T(), err)
}

// TestFingerprintFile_CaptureCleanliness pins the contract for the
// most common shell-capture pattern: `kosli fingerprint <file> --artifact-type=file`.
// stdout must be exactly the fingerprint + newline; stderr must be exactly empty.
func (suite *FingerprintCaptureTestSuite) TestFingerprintFile_CaptureCleanliness() {
	// Stub the update check to deterministically return a notice, so the
	// test fails if any code path forwards that notice to stderr — which
	// is exactly what was breaking customer CI pipelines.
	defer version.SetCheckForUpdateOverride(func(string) (string, error) {
		return fakeUpdateNotice, nil
	})()

	_, combined, stdout, stderr, err := executeCommandC(
		"fingerprint --artifact-type file testdata/file1")
	suite.Require().NoError(err)

	// Contract 1: stdout is exactly the fingerprint and nothing else.
	suite.Equal(file1Fingerprint+"\n", stdout,
		"stdout must contain only the fingerprint — anything else breaks shell capture")

	// Contract 2: stderr is exactly empty on the success path.
	// Stricter than NotContains because the threat is general — any new
	// stderr writer (deprecation notice, telemetry warning, framework
	// log) would silently break `2>&1` capture in customer CI.
	suite.Equal("", stderr,
		"stderr must be empty — any output here pollutes 2>&1 capture pipelines")

	// Contract 3: combined stdout+stderr (what `2>&1` capture sees) parses
	// as a fingerprint. This is the customer's actual usage:
	//   FP=$(kosli fingerprint ... 2>&1)
	// If this fails, customer CI fails.
	suite.Equal(file1Fingerprint+"\n", combined,
		"combined output (the 2>&1 capture pattern) must be exactly the fingerprint")
}

// TestFingerprintDir_CaptureCleanliness covers the directory variant. The
// original cyber-dojo bug fired here because the dir path runs long enough
// for the background version-check goroutine to complete and write to
// stderr before the command exits. Same three contracts as the file variant.
func (suite *FingerprintCaptureTestSuite) TestFingerprintDir_CaptureCleanliness() {
	defer version.SetCheckForUpdateOverride(func(string) (string, error) {
		return fakeUpdateNotice, nil
	})()

	_, combined, stdout, stderr, err := executeCommandC(
		"fingerprint --artifact-type dir testdata/folder1")
	suite.Require().NoError(err)

	suite.Equal(folder1Fingerprint+"\n", stdout,
		"stdout must contain only the fingerprint — anything else breaks shell capture")

	suite.Equal("", stderr,
		"stderr must be empty — any output here pollutes 2>&1 capture pipelines")

	suite.Equal(folder1Fingerprint+"\n", combined,
		"combined output (the 2>&1 capture pattern) must be exactly the fingerprint")
}

// TestFingerprintDocker_CaptureCleanliness covers the docker variant. The
// docker code path goes through internal/docker.GetImageFingerprint, which
// hits the local Docker daemon and is entirely separate from the file/dir
// hashing path — so it could legitimately introduce its own stderr writers
// (Docker API warnings, daemon connection logs, etc.) that the file/dir
// tests would not catch. Pinning the contract here protects that surface.
//
// Mirrors the docker test in fingerprint_test.go: alpine pinned by digest,
// so the resulting fingerprint is stable and the assertion can be exact.
func (suite *FingerprintCaptureTestSuite) TestFingerprintDocker_CaptureCleanliness() {
	defer version.SetCheckForUpdateOverride(func(string) (string, error) {
		return fakeUpdateNotice, nil
	})()

	const alpineFingerprint = "e15947432b813e8ffa90165da919953e2ce850bef511a0ad1287d7cb86de84b5"

	_, combined, stdout, stderr, err := executeCommandC(
		"fingerprint --artifact-type docker " + suite.dockerImage)
	suite.Require().NoError(err)

	suite.Equal(alpineFingerprint+"\n", stdout,
		"stdout must contain only the fingerprint — anything else breaks shell capture")

	suite.Equal("", stderr,
		"stderr must be empty — any output here pollutes 2>&1 capture pipelines")

	suite.Equal(alpineFingerprint+"\n", combined,
		"combined output (the 2>&1 capture pattern) must be exactly the fingerprint")
}

// TestFingerprintFile_DebugModeIsAllowedToWriteStderr pins the *other*
// half of the contract: stderr is an opt-in channel for debug output, not
// silent across the board. This stops a future contributor from "fixing"
// a TestFingerprintFile_CaptureCleanliness failure by silencing all stderr
// — they need to keep the debug channel working.
//
// stdout MUST still be exactly the fingerprint, even in debug mode, because
// `FP=$(kosli fingerprint --debug=true ...)` (without 2>&1) is still a
// supported pattern that the customer might use when troubleshooting.
func (suite *FingerprintCaptureTestSuite) TestFingerprintFile_DebugModeIsAllowedToWriteStderr() {
	defer version.SetCheckForUpdateOverride(func(string) (string, error) {
		return fakeUpdateNotice, nil
	})()

	_, _, stdout, stderr, err := executeCommandC(
		"fingerprint --artifact-type file testdata/file1 --debug=true")
	suite.Require().NoError(err)

	// stdout invariant holds even under --debug: the fingerprint and only
	// the fingerprint. This protects `$(...)` capture (which doesn't
	// include stderr) from being broken by debug output.
	suite.Equal(file1Fingerprint+"\n", stdout,
		"stdout must be the fingerprint even in debug mode — protects $(...) capture")

	// In debug mode the fingerprint command's own debug output MUST reach
	// stderr. Asserting on the fingerprint-specific log line (not just any
	// debug output) ensures this catches a regression where the logger is
	// silenced *during* the fingerprint operation — e.g. someone "fixing"
	// a CaptureCleanliness failure by routing logger.ErrOut to io.Discard
	// in the fingerprint code path. Earlier framework-level debug logs
	// from PreRunE would mask that, so we assert on the fingerprint marker.
	suite.Contains(stderr, "calculated fingerprint",
		"fingerprint-specific debug output must reach stderr in --debug mode — "+
			"if this fails, someone has silenced the logger inside the fingerprint code path")
}

func TestFingerprintCaptureTestSuite(t *testing.T) {
	suite.Run(t, new(FingerprintCaptureTestSuite))
}
