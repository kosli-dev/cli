package azure

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

func TestWarnAboutDigestsSource(t *testing.T) {
	t.Run("logs source warns that the container can write the log", func(t *testing.T) {
		var errOut bytes.Buffer
		warnAboutDigestsSource("logs", logger.NewLogger(io.Discard, &errOut, false))

		require.Contains(t, errOut.String(), "[warning]")
		require.Contains(t, errOut.String(), "--digests-source logs")
		require.Contains(t, errOut.String(), "container")
		require.Contains(t, errOut.String(), "--digests-source acr")
	})

	t.Run("acr source is silent", func(t *testing.T) {
		var errOut bytes.Buffer
		warnAboutDigestsSource("acr", logger.NewLogger(io.Discard, &errOut, false))

		require.Empty(t, errOut.String())
	})
}

const (
	genuineDigest = "1b7c84fc8a533a34ed6e8553976c6b68d97adaa1dbe6499265e7a76ac75801d4"
	spoofedDigest = "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee"
	testSite      = "api-service"
)

// pullBlock is what the App Service platform logs when it pulls an image, as
// captured in design-docs/azure_env_reporting_algorithm.md. minute is the
// "2023-09-28T12:27" prefix shared by every line of one block.
func pullBlock(minute, digest string) []string {
	return []string{
		minute + ":30.909Z INFO  - 3a9444c255ce Extracting 1KB / 1KB",
		minute + ":31.086Z INFO  - 3a9444c255ce Pull complete",
		minute + ":31.201Z INFO  -  Digest: sha256:" + digest,
		minute + ":31.250Z INFO  -  Status: Downloaded newer image for reg.azurecr.io/team/app@sha256:" + digest,
		minute + ":31.282Z INFO  - Pull Image successful, Time taken: 1 Minutes and 8 Seconds",
	}
}

// startBlock is what the platform logs when it starts the container. The
// container's own output can only appear after the "docker run" line.
func startBlock(minute, digest, site string) []string {
	return []string{
		minute + ":33.104Z INFO  - Starting container for site",
		minute + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + site + "_0_5b07493a -e WEBSITE_SITE_NAME=" + site + " reg.azurecr.io/team/app@sha256:" + digest + "  ",
		"",
	}
}

// readyBlock is what the platform logs once the container answers its warmup request.
func readyBlock(minute, site string) []string {
	return []string{
		minute + ":36.389Z INFO  - Initiating warmup request to container " + site + "_0_5b07493a for site " + site,
		minute + ":37.414Z INFO  - Container " + site + "_0_5b07493a for site " + site + " initialized successfully and is ready to serve requests.",
	}
}

func deployment(minute, digest, site string) []string {
	var lines []string
	lines = append(lines, pullBlock(minute, digest)...)
	lines = append(lines, startBlock(minute, digest, site)...)
	lines = append(lines, readyBlock(minute, site)...)
	return lines
}

// containerLine is a line the container wrote to stdout. Docker timestamps it at
// nanosecond precision and the platform adds no level or dash.
func containerLine(minute, text string) string {
	return minute + ":35.123456789Z " + text
}

func readyAt(t *testing.T, minute string) int64 {
	t.Helper()
	ts, err := time.Parse(time.RFC3339Nano, minute+":37.414Z")
	require.NoError(t, err)
	return ts.Unix()
}

func join(blocks ...[]string) []byte {
	var lines []string
	for _, block := range blocks {
		lines = append(lines, block...)
	}
	return []byte(strings.Join(lines, "\n") + "\n")
}

func TestExtractImageFingerprintAndStartedTimestampFromLogs(t *testing.T) {
	first, second := "2023-09-28T12:27", "2023-09-28T13:41"

	for _, tc := range []struct {
		name            string
		logs            []byte
		wantFingerprint string
		wantStartedAt   int64
	}{
		{
			name:            "a genuine pull and start yields its digest and start time",
			logs:            join(deployment(first, genuineDigest, testSite)),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// The stream carries the container's stdout too. Whatever it prints
			// after "docker run" is its own claim, not the platform's.
			name: "a digest line the container prints after it starts is ignored",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlock(first, genuineDigest, testSite),
				[]string{containerLine(first, "Digest: sha256:"+spoofedDigest)},
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a digest line mimicking the platform format after the start is ignored",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlock(first, genuineDigest, testSite),
				[]string{first + ":35.123Z INFO  -  Digest: sha256:" + spoofedDigest},
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a digest line printed after initialization is ignored",
			logs: join(
				deployment(first, genuineDigest, testSite),
				[]string{
					containerLine(second, "Digest: sha256:"+spoofedDigest),
					second + ":38.000Z INFO  -  Digest: sha256:" + spoofedDigest,
				},
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// The container being replaced is still running while the platform
			// pulls its successor, so it can write between the pull and the start.
			// Only the timestamp shape tells its line from the platform's.
			name: "a container-timestamped digest line printed during the next pull is ignored",
			logs: join(
				deployment(first, spoofedDigest, testSite),
				pullBlock(second, genuineDigest),
				[]string{second + ":31.500000000Z INFO  -  Digest: sha256:" + spoofedDigest},
				startBlock(second, genuineDigest, testSite),
				readyBlock(second, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, second),
		},
		{
			// Regression guard against "take the first match": the log window
			// can hold several deployments and the running one is the latest.
			name: "two deployments yield the latest digest",
			logs: join(
				deployment(first, spoofedDigest, testSite),
				deployment(second, genuineDigest, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, second),
		},
		{
			name: "a pull after the last successful start yields nothing",
			logs: join(
				deployment(first, genuineDigest, testSite),
				pullBlock(second, spoofedDigest),
				startBlock(second, spoofedDigest, testSite),
			),
		},
		{
			name: "a start with no initialization line yields the digest with no start time",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlock(first, genuineDigest, testSite),
			),
			wantFingerprint: genuineDigest,
		},
		{
			name: "a digest with no container start yields nothing",
			logs: join(pullBlock(first, genuineDigest)),
		},
		{
			name:            "an initialization line for another site does not set the start time",
			logs:            join(deployment(first, genuineDigest, "other-site")),
			wantFingerprint: genuineDigest,
		},
		{
			name: "a short digest line does not panic",
			logs: join(
				[]string{first + ":31.201Z INFO  -  Digest: sha256:abc"},
				startBlock(first, genuineDigest, testSite),
				readyBlock(first, testSite),
			),
		},
		{
			name: "empty logs yield nothing",
			logs: []byte{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fingerprint, startedAt, err := extractImageFingerprintAndStartedTimestampFromLogs(tc.logs, testSite)
			require.NoError(t, err)
			require.Equal(t, tc.wantFingerprint, fingerprint)
			require.Equal(t, tc.wantStartedAt, startedAt)
		})
	}
}
