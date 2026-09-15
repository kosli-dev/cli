package azure

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	armappservice "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice/v2"
	"github.com/kosli-dev/cli/internal/logger"
	"github.com/stretchr/testify/require"
)

// logsClient is a client in logs mode whose log fetch returns the given log
// instead of calling Azure.
func logsClient(logs []byte) *AzureClient {
	return &AzureClient{
		Credentials:      AzureStaticCredentials{DigestsSource: "logs"},
		dockerLogsForApp: func(string, *logger.Logger) ([]byte, error) { return logs, nil },
	}
}

func TestFingerprintDockerServiceUsesTheLogsSource(t *testing.T) {
	appName, appKind := testSite, "app"
	imageName := testImage
	minute := "2023-09-28T12:27"

	t.Run("reports the digest the platform ran and when it was ready", func(t *testing.T) {
		client := logsClient(join(deployment(minute, genuineDigest, testSite)))
		var errOut bytes.Buffer

		appData, err := client.fingerprintDockerService(
			&armappservice.Site{Name: &appName, Kind: &appKind}, logger.NewLogger(io.Discard, &errOut, false), imageName)

		require.NoError(t, err)
		require.Equal(t, AppData{
			AppName:       appName,
			AppKind:       appKind,
			DigestsSource: "logs",
			Digests:       map[string]string{imageName: genuineDigest},
			StartedAt:     readyAt(t, minute),
		}, appData)
		require.Empty(t, errOut.String())
	})

	t.Run("says so when the log holds no platform digest", func(t *testing.T) {
		client := logsClient(join(pullBlock(minute, genuineDigest)))
		var errOut bytes.Buffer

		appData, err := client.fingerprintDockerService(
			&armappservice.Site{Name: &appName, Kind: &appKind}, logger.NewLogger(io.Discard, &errOut, false), imageName)

		require.NoError(t, err)
		require.Equal(t, map[string]string{imageName: ""}, appData.Digests)
		require.Contains(t, errOut.String(), "[warning]")
		require.Contains(t, errOut.String(), appName)
		require.Contains(t, errOut.String(), "without a fingerprint")
	})
}

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
	// testImage is the configured reference; fixtures pin its repository by digest.
	testImage = "reg.azurecr.io/team/app:v1"
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

// startBlock is what the platform logs when it starts the container.
func startBlock(minute, digest, site string) []string {
	return []string{
		minute + ":33.104Z INFO  - Starting container for site",
		minute + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + site + "_0_5b07493a -e WEBSITE_SITE_NAME=" + site + " reg.azurecr.io/team/app@sha256:" + digest + "  ",
		"",
	}
}

// upToDatePullBlock is the pull of an image the host already has.
func upToDatePullBlock(minute, digest string) []string {
	return []string{
		minute + ":30.909Z INFO  - Pulling image: reg.azurecr.io/team/app:v1",
		minute + ":31.201Z INFO  -  Digest: sha256:" + digest,
		minute + ":31.250Z INFO  -  Status: Image is up to date for reg.azurecr.io/team/app:v1",
		minute + ":31.282Z INFO  - Pull Image successful, Time taken: 0 Minutes and 1 Seconds",
	}
}

// startBlockWithImage is a start whose "docker run" ends with the given image
// reference, followed by an optional startup command.
func startBlockWithImage(minute, site, image string, command ...string) []string {
	return []string{
		minute + ":33.104Z INFO  - Starting container for site",
		minute + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + site + "_0_5b07493a -e WEBSITE_SITE_NAME=" + site + " " + image + " " + strings.Join(command, " ") + " ",
		"",
	}
}

// startBlockByTag is a start whose "docker run" names the image by tag, so the
// line carries no digest of its own.
func startBlockByTag(minute, site string) []string {
	return startBlockWithImage(minute, site, testImage)
}

// startBlockWithCommand is a digest-pinned start followed by the app's
// configured startup command.
func startBlockWithCommand(minute, digest, site, command string) []string {
	return startBlockWithImage(minute, site, "reg.azurecr.io/team/app@sha256:"+digest, command)
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
		name string
		logs []byte
		// image is the reference the app's configuration names; testImage
		// when empty.
		image string
		// site is the app name as configured; testSite when empty.
		site            string
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
			// Whatever the container prints after "docker run" is its own claim.
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
			// A container being replaced still runs during its successor's pull, so
			// only the timestamp shape tells its line from the platform's.
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
			name: "a deployment of another site yields nothing",
			logs: join(deployment(first, genuineDigest, "other-site")),
		},
		{
			name: "the container name in --name= form identifies the site",
			logs: join(
				[]string{
					first + ":33.104Z INFO  - Starting container for site",
					first + ":33.104Z INFO  - docker run -d -p 6693:3000 --name=" + testSite + "_0_5b07493a -e FOO=bar reg.azurecr.io/team/app@sha256:" + genuineDigest + "  ",
				},
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// A slot's container is named "<site>__<slot>_...", which starts with
			// "<site>_" but has no instance digit after the underscore.
			name: "a start whose container name merely begins with the site name is ignored",
			logs: join(
				pullBlock(first, genuineDigest),
				[]string{
					first + ":33.104Z INFO  - Starting container for site",
					first + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + testSite + "__staging_0_5b07493a -e WEBSITE_SITE_NAME=" + testSite + "__staging reg.azurecr.io/team/app@sha256:" + spoofedDigest + "  ",
				},
				readyBlock(first, testSite),
			),
		},
		{
			// App names are hostnames, so the configured spelling and the log's
			// can differ in case.
			name:            "site name case does not matter",
			site:            "API-Service",
			logs:            join(deployment(first, genuineDigest, testSite)),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// A scale-out starts a second instance of the same image; the window
			// can end before that instance is ready while the first is serving.
			name: "a scale-out start of the same digest keeps the fingerprint",
			logs: join(
				deployment(first, genuineDigest, testSite),
				[]string{
					second + ":33.104Z INFO  - Starting container for site",
					second + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + testSite + "_1_9c1d22e0 -e WEBSITE_SITE_NAME=" + testSite + " reg.azurecr.io/team/app@sha256:" + genuineDigest + "  ",
				},
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// The ready container's start precedes the window, so its digest is
			// unknown and the new start could be a replacement.
			name: "a start after a ready report of unknown digest yields nothing",
			logs: join(
				readyBlock(first, testSite),
				startBlock(second, genuineDigest, testSite),
			),
		},
		{
			name: "a replacement start of another digest after a ready report yields nothing",
			logs: join(
				deployment(first, genuineDigest, testSite),
				startBlock(second, spoofedDigest, testSite),
			),
		},
		{
			name: "a later start of another site does not replace this site's",
			logs: join(
				deployment(first, genuineDigest, testSite),
				startBlock(second, spoofedDigest, "other-site"),
				readyBlock(second, "other-site"),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// Over bufio.Scanner's token cap; a scanner would fail the whole snapshot.
			name: "a container line longer than 64KiB does not fail the app",
			logs: join(
				pullBlock(first, genuineDigest),
				[]string{containerLine(first, strings.Repeat("x", 70*1024))},
				startBlock(first, genuineDigest, testSite),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// A configured startup command follows the image on the run line, so
			// the image is not always its last argument.
			name: "a startup command after the image does not supply the digest",
			logs: join(
				startBlockWithCommand(first, genuineDigest, testSite, "node server.js --seed other/repo@sha256:"+spoofedDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a run line pinning another repository does not supply the digest",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlockWithImage(first, testSite, "reg.azurecr.io/team/other@sha256:"+spoofedDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a run line token longer than a digest is not a digest",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlockWithImage(first, testSite, "reg.azurecr.io/team/app@sha256:"+spoofedDigest+"a"),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// Two tokens pin the configured repository and disagree, so the run
			// line is not trusted and the pulled digest is used.
			name: "a startup command naming the configured repository under another digest disables the run line",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlockWithCommand(first, spoofedDigest, testSite, "node server.js reg.azurecr.io/team/app@sha256:"+genuineDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// App settings are passed as -e KEY=VALUE before the image, and a
			// value containing whitespace becomes extra tokens.
			name: "an app setting carrying the configured repository under another digest disables the run line",
			logs: join(
				pullBlock(first, genuineDigest),
				[]string{
					first + ":33.104Z INFO  - Starting container for site",
					first + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + testSite + "_0_5b07493a -e WEBSITE_SITE_NAME=" + testSite +
						" -e FOO=x reg.azurecr.io/team/app@sha256:" + spoofedDigest + " reg.azurecr.io/team/app@sha256:" + genuineDigest + "  ",
				},
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// The tag token must count as a competing claim, or the injected digest
			// stands unopposed.
			name: "an app setting carrying a digest does not supply it when the image ran by tag",
			logs: join(
				pullBlock(first, genuineDigest),
				[]string{
					first + ":33.104Z INFO  - Starting container for site",
					first + ":33.104Z INFO  - docker run -d -p 6693:3000 --name " + testSite + "_0_5b07493a -e WEBSITE_SITE_NAME=" + testSite +
						" -e FOO=x reg.azurecr.io/team/app@sha256:" + spoofedDigest + " " + testImage + "  ",
				},
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a startup command repeating the image under the same digest keeps the run line",
			logs: join(
				startBlockWithCommand(first, genuineDigest, testSite, "node server.js reg.azurecr.io/team/app@sha256:"+genuineDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name:  "a docker hub short form matches an expanded run line",
			image: "nginx:latest",
			logs: join(
				startBlockWithImage(first, testSite, "docker.io/library/nginx@sha256:"+genuineDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name:  "a docker hub short form matches an unexpanded run line",
			image: "nginx:latest",
			logs: join(
				startBlockWithImage(first, testSite, "nginx@sha256:"+genuineDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name:  "registry host case does not matter",
			image: "MyReg.azurecr.io/team/app:v1",
			logs: join(
				startBlockWithImage(first, testSite, "myreg.azurecr.io/team/app@sha256:"+genuineDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name:  "an unparsable configured image falls back to the pulled digest",
			image: "not a reference",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlockWithImage(first, testSite, "reg.azurecr.io/team/app@sha256:"+spoofedDigest),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// The prefix checks only the timestamp's shape.
			name: "a ready line whose timestamp is not an instant is skipped",
			logs: join(
				startBlock(first, genuineDigest, testSite),
				[]string{"2023-13-45T25:61:61.000Z INFO  - Container x for site " + testSite + " initialized successfully and is ready to serve requests."},
			),
			wantFingerprint: genuineDigest,
		},
		{
			// The run line is both start proof and digest, so nothing earlier can
			// stand in for it.
			name: "the digest the platform ran wins over a platform-shaped line before the start",
			logs: join(
				pullBlock(first, genuineDigest),
				[]string{first + ":31.900Z INFO  -  Digest: sha256:" + spoofedDigest},
				startBlock(first, genuineDigest, testSite),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a restart with no pull in the window yields the digest the platform ran",
			logs: join(
				startBlock(first, genuineDigest, testSite),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a run line naming a tag falls back to the pulled digest",
			logs: join(
				pullBlock(first, genuineDigest),
				startBlockByTag(first, testSite),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a run line naming a tag with no pull in the window yields nothing",
			logs: join(
				startBlockByTag(first, testSite),
				readyBlock(first, testSite),
			),
		},
		{
			name: "an up to date pull yields its digest",
			logs: join(
				upToDatePullBlock(first, genuineDigest),
				startBlockByTag(first, testSite),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			// The anti-spoof property is the timestamp shape, not the level.
			name: "a platform line at another level still counts",
			logs: join(
				[]string{first + ":31.201Z WARN  -  Digest: sha256:" + genuineDigest},
				startBlockByTag(first, testSite),
				readyBlock(first, testSite),
			),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name:            "windows line endings are accepted",
			logs:            []byte(strings.ReplaceAll(string(join(deployment(first, genuineDigest, testSite))), "\n", "\r\n")),
			wantFingerprint: genuineDigest,
			wantStartedAt:   readyAt(t, first),
		},
		{
			name: "a short digest line does not panic",
			logs: join(
				[]string{first + ":31.201Z INFO  -  Digest: sha256:abc"},
				startBlockByTag(first, testSite),
				readyBlock(first, testSite),
			),
		},
		{
			name: "empty logs yield nothing",
			logs: []byte{},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			image := tc.image
			if image == "" {
				image = testImage
			}
			site := tc.site
			if site == "" {
				site = testSite
			}
			fingerprint, startedAt := extractImageFingerprintAndStartedTimestampFromLogs(tc.logs, site, image, logger.NewLogger(io.Discard, io.Discard, false))
			require.Equal(t, tc.wantFingerprint, fingerprint)
			require.Equal(t, tc.wantStartedAt, startedAt)
		})
	}
}

// A run line the site check rejects is the symptom of a platform log shape the
// parser does not know, so it must be visible under --debug.
func TestExtractImageFingerprintLogsARejectedRunLine(t *testing.T) {
	var errOut bytes.Buffer
	log := logger.NewLogger(io.Discard, &errOut, true)

	fingerprint, _ := extractImageFingerprintAndStartedTimestampFromLogs(
		join(deployment("2023-09-28T12:27", genuineDigest, "other-site")), testSite, testImage, log)

	require.Empty(t, fingerprint)
	require.Contains(t, errOut.String(), "[debug]")
	require.Contains(t, errOut.String(), "does not name site "+testSite)
	require.Contains(t, errOut.String(), "docker run")
}
