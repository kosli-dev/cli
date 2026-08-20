#!/bin/bash
# Runs smoke tests against the built docker image and records each case's
# outcome to $RESULTS_FILE for the CI workflow to report as a Kosli
# attestation.
#
# Usage: IMAGE=... TAG=... RESULTS_FILE=... ./scripts/docker-smoke-tests.sh
set -uo pipefail

IMAGE="${IMAGE:?IMAGE is required}"
TAG="${TAG:?TAG is required}"
RESULTS_FILE="${RESULTS_FILE:?RESULTS_FILE is required}"

REPO_ROOT="${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}"
if [ -z "$REPO_ROOT" ]; then
  echo "could not determine repo root" >&2
  exit 1
fi

EXIT_CODE=0
RESULTS="[]"

# Runs a smoke test case and records its outcome in $RESULTS_FILE, keyed by
# name. The CI workflow reports a single aggregate attestation, compliant only
# if every recorded case succeeded, with the results file attached.
# Usage: run_case <case-name> <test-function>
run_case() {
  local name="$1"
  local test_fn="$2"

  echo "::group::Smoke test: ${name}"
  local outcome="success"
  if ! "$test_fn"; then
    outcome="failure"
    EXIT_CODE=1
  fi
  echo "::endgroup::"
  echo "Smoke test ${name}: ${outcome}"

  RESULTS="$(jq --arg name "$name" --arg outcome "$outcome" \
    '. + [{name: $name, outcome: $outcome}]' <<< "$RESULTS")" || { echo "jq failed" >&2; exit 1; }
  printf '%s\n' "$RESULTS" > "$RESULTS_FILE" || { echo "failed to write $RESULTS_FILE" >&2; exit 1; }
}

# --- Smoke test cases -------------------------------------------------
# Add a new smoke test by writing a test_* function below and adding one
# run_case call for it — no CI workflow changes needed.

test_list_environments() {
  docker run --rm \
    -e KOSLI_API_TOKEN=any-token-will-do \
    -e KOSLI_ORG=cyber-dojo \
    "${IMAGE}:${TAG}" \
    list environments
}

test_attest_artifact_dir() {
  local commit_sha
  commit_sha="$(git -C "$REPO_ROOT" rev-parse HEAD)"

  docker run --rm \
    -v "${REPO_ROOT}":/workspace:ro \
    -w /workspace \
    -e KOSLI_API_TOKEN=any-token-will-do \
    -e KOSLI_ORG=test-org \
    "${IMAGE}:${TAG}" \
    attest artifact /workspace/internal/utils \
    --artifact-type dir \
    --flow test-flow \
    --trail test-trail \
    --name test-artifact \
    --build-url https://example.com/build/1 \
    --commit-url "https://github.com/kosli-dev/cli/commit/${commit_sha}" \
    --repo-root /workspace \
    --dry-run \
    --debug
}

# --- Run all cases ------------------------------------------------------

run_case "list-environments" test_list_environments
run_case "attest-artifact-dir" test_attest_artifact_dir

exit $EXIT_CODE
