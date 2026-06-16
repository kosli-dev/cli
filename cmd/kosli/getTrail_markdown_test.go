package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPrintTrailAsMarkdown is a server-free unit test for the markdown renderer.
// It feeds a representative trail payload through printTrailAsMarkdown and
// compares the result against the golden file used by the integration test, so
// global opts and the flow name mirror the integration suite's values.
func TestPrintTrailAsMarkdown(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true") // makes formattedTimestamp use a fixed time
	global = &GlobalOpts{
		Host: "http://localhost:8001",
		Org:  "docs-cmd-test-user-shared",
	}
	o := &getTrailOptions{flowName: "get-trail"}

	// Mirrors the get-trail integration fixture: a freshly-begun trail from a
	// template with a trail attestation "bar" and an artifact "cli" with
	// attestation "foo", none reported yet (all MISSING).
	raw := `{
		"name": "cli-build-1",
		"description": "test trail",
		"compliance_state": "INCOMPLETE",
		"last_modified_at": 1452902400,
		"compliance_status": {
			"status": "INCOMPLETE",
			"is_compliant": false,
			"attestations_statuses": [
				{"attestation_name": "bar", "attestation_type": "generic", "attestation_id": null, "overridden_attestation_id": null, "status": "MISSING", "is_compliant": null, "unexpected": false}
			],
			"artifacts_statuses": {
				"cli": {
					"artifact_fingerprint": null,
					"artifact_id": null,
					"status": "MISSING",
					"is_compliant": null,
					"attestations_statuses": [
						{"attestation_name": "foo", "attestation_type": "generic", "attestation_id": null, "overridden_attestation_id": null, "status": "MISSING", "is_compliant": null, "unexpected": false}
					],
					"unexpected": false
				}
			}
		},
		"events": [
			{"type": "trail_reported", "timestamp": 1452902400}
		]
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	want, err := os.ReadFile(goldenPath("output/get/get-trail-markdown.txt"))
	require.NoError(t, err)

	require.Equal(t, string(want), buf.String())
}

// TestPrintTrailAsMarkdownGitCommitAndEscaping covers the git-commit block and
// mdCell escaping: pipes, CR/CRLF, and angle brackets (e.g. author emails) must
// not break the table layout or be swallowed as HTML, and missing (null) fields
// must render as empty cells, not "<nil>".
func TestPrintTrailAsMarkdownGitCommitAndEscaping(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-2",
		"description": null,
		"compliance_state": "COMPLIANT",
		"last_modified_at": 1452902400,
		"origin_url": "https://github.com/kosli-dev/cli/actions/runs/123",
		"git_commit_info": {
			"sha1": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
			"author": "Jane | Doe <jane@kosli.com>",
			"timestamp": 1452902400,
			"url": "https://github.com/kosli-dev/cli/commit/1a2b3c4",
			"message": "fix: a | in a title\r\nsecond line\rthird line\nfourth line"
		},
		"events": []
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()

	require.Contains(t, got, "## Trail: [cli-build-2](https://app.kosli.com/my-org/flows/my-flow/trails/cli-build-2)")
	require.NotContains(t, got, "<nil>", "null fields must render as empty cells")
	require.NotContains(t, got, "\r", "carriage returns must not survive into the markdown")
	require.Contains(t, got, "| Description |  |")
	require.Contains(t, got, "| Compliance | ✅ COMPLIANT |")
	require.Contains(t, got, "| Origin | https://github.com/kosli-dev/cli/actions/runs/123 |")
	require.Contains(t, got, "| Author | Jane \\| Doe &lt;jane@kosli.com&gt; |")
	require.Contains(t, got, "| Sha1 | [1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b](https://github.com/kosli-dev/cli/commit/1a2b3c4) |")
	require.Contains(t, got, "| Message | fix: a \\| in a title |",
		"only the first line of the commit message is shown")
	require.Contains(t, got, "### Git commit")
	require.Contains(t, got, "_No events._")
}

// TestPrintTrailAsMarkdownEventLinksAndCompliance covers event rows: the commit
// column links to the commit URL when available, and compliance gets a
// glanceable emoji prefix.
func TestPrintTrailAsMarkdownEventLinksAndCompliance(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-3",
		"description": "events trail",
		"compliance_state": "NON_COMPLIANT",
		"last_modified_at": 1452902400,
		"events": [
			{
				"type": "trail_attestation_reported",
				"timestamp": 1452902400,
				"attestation_type": "junit",
				"template_reference_name": "unit-tests",
				"is_compliant": true,
				"git_commit_info": {
					"sha1": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b",
					"url": "https://github.com/kosli-dev/cli/commit/1a2b3c4d"
				}
			},
			{
				"type": "trail_attestation_reported",
				"timestamp": 1452902400,
				"attestation_type": "snyk",
				"template_reference_name": "snyk-scan",
				"is_compliant": false,
				"git_commit_info": {
					"sha1": "9f8e7d6c5b4a39281706f5e4d3c2b1a098765432"
				}
			}
		]
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()

	require.Contains(t, got, "| Compliance | ❌ NON_COMPLIANT |")
	require.Contains(t, got, "| [1a2b3c4](https://github.com/kosli-dev/cli/commit/1a2b3c4d) |",
		"commit column links to the commit URL when available")
	require.Contains(t, got, "| 9f8e7d6 |", "commit without a URL renders as plain short sha")
	require.Contains(t, got, "| ✅ compliant |")
	require.Contains(t, got, "| ❌ non-compliant |")
}

// TestPrintTrailAsMarkdownAttestationStatuses covers the attestation-statuses
// table: each attestation name links to the attestation in the app, the
// compliance column uses emoji, and every compliance status is handled
// (COMPLETE+compliant, COMPLETE+non-compliant, MISSING, and the unexpected flag),
// grouped by trail and by artifact.
func TestPrintTrailAsMarkdownAttestationStatuses(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-6",
		"description": "attestation statuses trail",
		"compliance_state": "NON_COMPLIANT",
		"last_modified_at": 1452902400,
		"compliance_status": {
			"status": "NON-COMPLIANT",
			"is_compliant": false,
			"attestations_statuses": [
				{"attestation_name": "bar", "attestation_type": "generic", "attestation_id": "aaa-bar", "status": "COMPLETE", "is_compliant": true, "unexpected": false}
			],
			"artifacts_statuses": {
				"cli": {
					"status": "NON-COMPLIANT",
					"is_compliant": false,
					"attestations_statuses": [
						{"attestation_name": "foo", "attestation_type": "generic", "attestation_id": "aaa-foo", "status": "COMPLETE", "is_compliant": true, "unexpected": false},
						{"attestation_name": "baz", "attestation_type": "snyk", "attestation_id": "aaa-baz", "status": "COMPLETE", "is_compliant": false, "unexpected": false},
						{"attestation_name": "qux", "attestation_type": "junit", "attestation_id": null, "status": "MISSING", "is_compliant": null, "unexpected": false},
						{"attestation_name": "extra", "attestation_type": "generic", "attestation_id": "aaa-extra", "status": "COMPLETE", "is_compliant": true, "unexpected": true}
					]
				}
			}
		},
		"events": []
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()
	trailURL := "https://app.kosli.com/my-org/flows/my-flow/trails/cli-build-6"

	require.Contains(t, got, "### Attestations")
	require.Contains(t, got, "**Trail**")
	require.Contains(t, got, "**cli** — ❌ NON-COMPLIANT")
	// trail-level attestation, linked, compliant
	require.Contains(t, got, fmt.Sprintf("| [bar](%s?attestation_id=aaa-bar) | ✅ compliant |", trailURL))
	// artifact attestations: compliant, non-compliant, missing (unlinked), unexpected
	require.Contains(t, got, fmt.Sprintf("| [foo](%s?attestation_id=aaa-foo) | ✅ compliant |", trailURL))
	require.Contains(t, got, fmt.Sprintf("| [baz](%s?attestation_id=aaa-baz) | ❌ non-compliant |", trailURL))
	require.Contains(t, got, "| qux | ⏳ missing |")
	require.Contains(t, got, fmt.Sprintf("| [extra](%s?attestation_id=aaa-extra) | ✅ compliant (+) |", trailURL))
}

// TestPrintTrailAsMarkdownSkipsEmptyArtifactAttestations covers an artifact in
// artifacts_statuses that has no attestations of its own: it must not emit a
// header followed by an empty table when another artifact (or the trail) does
// have attestations, so the section-level total > 0 guard still passes.
func TestPrintTrailAsMarkdownSkipsEmptyArtifactAttestations(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-7",
		"description": "mixed artifacts trail",
		"compliance_state": "COMPLIANT",
		"last_modified_at": 1452902400,
		"compliance_status": {
			"status": "COMPLIANT",
			"is_compliant": true,
			"attestations_statuses": [],
			"artifacts_statuses": {
				"has-atts": {
					"status": "COMPLIANT",
					"is_compliant": true,
					"attestations_statuses": [
						{"attestation_name": "foo", "attestation_type": "generic", "attestation_id": "aaa-foo", "status": "COMPLETE", "is_compliant": true, "unexpected": false}
					]
				},
				"no-atts": {
					"status": "COMPLIANT",
					"is_compliant": true,
					"attestations_statuses": []
				}
			}
		},
		"events": []
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()

	require.Contains(t, got, "**has-atts** — ✅ COMPLIANT",
		"an artifact with attestations is still rendered")
	require.NotContains(t, got, "no-atts",
		"an artifact with zero attestations is skipped entirely, not rendered as an empty table")
}

// TestPrintTrailAsMarkdownShortCommitSha covers a malformed commit sha shorter
// than 7 characters on an event: the short-sha slice must be length-guarded so
// it can't panic with a slice-bounds error.
func TestPrintTrailAsMarkdownShortCommitSha(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-8",
		"description": "short sha trail",
		"compliance_state": "COMPLIANT",
		"last_modified_at": 1452902400,
		"events": [
			{
				"type": "trail_reported",
				"timestamp": 1452902400,
				"git_commit_info": {"sha1": "abc"}
			}
		]
	}`

	var buf bytes.Buffer
	require.NotPanics(t, func() {
		err := o.printTrailAsMarkdown(raw, &buf, 0)
		require.NoError(t, err)
	})
}

// TestPrintTrailAsMarkdownAttestationLinks covers linking attestation events to
// the attestation in the Kosli app:
// {host}/{org}/flows/{flow}/trails/{trail}?attestation_id={id}.
func TestPrintTrailAsMarkdownAttestationLinks(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-5",
		"description": "attestation trail",
		"compliance_state": "COMPLIANT",
		"last_modified_at": 1452902400,
		"events": [
			{
				"type": "trail_attestation_for_artifact_reported",
				"timestamp": 1452902400,
				"attestation_type": "snyk",
				"target_artifact": "artifact",
				"template_reference_name": "snyk-code-test",
				"attestation_id": "b8366cb0-249f-419e-b68a-d7046b7b",
				"is_compliant": true
			},
			{
				"type": "trail_attestation_reported",
				"timestamp": 1452902400,
				"attestation_type": "junit",
				"template_reference_name": "unit-tests",
				"attestation_id": "59d541bb-136c-49d7-a086-6c4dc5eb",
				"is_compliant": true
			},
			{
				"type": "trail_attestation_reported",
				"timestamp": 1452902400,
				"attestation_type": "generic",
				"template_reference_name": "no-id-attestation",
				"is_compliant": true
			}
		]
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()

	require.Contains(t, got,
		"'snyk' attestation reported for [artifact.snyk-code-test](https://app.kosli.com/my-org/flows/my-flow/trails/cli-build-5?attestation_id=b8366cb0-249f-419e-b68a-d7046b7b)")
	require.Contains(t, got,
		"'junit' attestation reported for [unit-tests](https://app.kosli.com/my-org/flows/my-flow/trails/cli-build-5?attestation_id=59d541bb-136c-49d7-a086-6c4dc5eb) on the trail")
	require.Contains(t, got,
		"'generic' attestation reported for no-id-attestation on the trail",
		"an attestation event without an id stays unlinked")
}

// TestPrintTrailAsMarkdownEnvironmentLinks covers linking the environment name
// in started/stopped running events to the environment snapshot in the Kosli
// app: {host}/{org}/environments/{env}/{snapshot-index}, falling back to the
// environment page when no snapshot index is present.
func TestPrintTrailAsMarkdownEnvironmentLinks(t *testing.T) {
	t.Setenv("KOSLI_TESTS", "true")
	global = &GlobalOpts{
		Host: "https://app.kosli.com",
		Org:  "my-org",
	}
	o := &getTrailOptions{flowName: "my-flow"}

	raw := `{
		"name": "cli-build-4",
		"description": "env trail",
		"compliance_state": "COMPLIANT",
		"last_modified_at": 1452902400,
		"events": [
			{
				"type": "artifact_started_running",
				"timestamp": 1452902400,
				"template_reference_name": "artifact",
				"environment_name": "staging-aws",
				"snapshot_index": 15144
			},
			{
				"type": "artifact_stopped_running",
				"timestamp": 1452902400,
				"template_reference_name": "artifact",
				"environment_name": "prod-aws"
			}
		]
	}`

	var buf bytes.Buffer
	err := o.printTrailAsMarkdown(raw, &buf, 0)
	require.NoError(t, err)

	got := buf.String()

	require.Contains(t, got,
		"artifact 'artifact' started running in ['staging-aws'](https://app.kosli.com/my-org/environments/staging-aws/15144)")
	require.Contains(t, got,
		"artifact 'artifact' stopped running in ['prod-aws'](https://app.kosli.com/my-org/environments/prod-aws/)",
		"without a snapshot index the link falls back to the environment page")
}
