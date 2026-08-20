#!/bin/bash
set -uo pipefail

IMAGE="${IMAGE:?IMAGE is required}"
TAG="${TAG:?TAG is required}"
RESULTS_FILE="${RESULTS_FILE:?RESULTS_FILE is required}"

REPO_ROOT="$(git rev-parse --show-toplevel)"
EXIT_CODE=0
RESULTS="[]"

# Runs a smoke test case and records its outcome in $RESULTS_FILE, keyed by
# name, so the CI workflow can report a Kosli attestation for it.
# Usage: run_case <attestation-name> <test-function>
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
    '. + [{name: $name, outcome: $outcome}]' <<< "$RESULTS")"
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
  docker run --rm \
    -v "${REPO_ROOT}":/workspace:ro \
    -w /workspace \
    -e KOSLI_API_TOKEN=DRY_RUN \
    -e KOSLI_ORG=test-org \
    "${IMAGE}:${TAG}" \
    attest artifact /workspace/internal/utils \
    --artifact-type dir \
    --flow test-flow \
    --trail test-trail \
    --name test-artifact \
    --build-url https://example.com/build/1 \
    --commit-url https://github.com/kosli-dev/cli/commit/HEAD \
    --repo-root /workspace \
    --dry-run \
    --debug
}

# --- Run all cases ------------------------------------------------------

run_case "smoke-test" test_list_environments
run_case "smoke-test-attest-artifact-dir" test_attest_artifact_dir

echo "$RESULTS" > "$RESULTS_FILE"
exit $EXIT_CODE
