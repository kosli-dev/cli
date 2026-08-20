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

# Updates the named entry's outcome in $RESULTS and flushes to $RESULTS_FILE.
write_result() {
  local name="$1"
  local outcome="$2"

  RESULTS="$(jq --arg name "$name" --arg outcome "$outcome" \
    'map(if .name == $name then .outcome = $outcome else . end)' <<< "$RESULTS")" \
    || { echo "jq failed" >&2; exit 1; }
  printf '%s\n' "$RESULTS" > "$RESULTS_FILE" || { echo "failed to write $RESULTS_FILE" >&2; exit 1; }
}

# Runs a smoke test case and records its outcome, keyed by name. The CI
# workflow reports a single aggregate attestation, compliant only if every
# recorded case succeeded, with the results file attached. Every case is
# seeded as "not-run" before any case executes (see below), so a run that
# aborts part-way leaves a results file that reads as incomplete rather than
# as a clean pass.
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

  write_result "$name" "$outcome"
}

# --- Smoke test cases -------------------------------------------------
# Add a new smoke test by writing a test_* function below and adding one
# entry to the CASES array further down — no CI workflow changes needed.

test_list_environments() {
  docker run --rm \
    -e KOSLI_API_TOKEN=any-token-will-do \
    -e KOSLI_ORG=cyber-dojo \
    "${IMAGE}:${TAG}" \
    list environments
}

test_attest_artifact_dir() {
  local commit_sha
  commit_sha="$(git -C "$REPO_ROOT" rev-parse HEAD)" || return 1

  # --api-token is a required flag, validated before --dry-run's own
  # short-circuit is ever reached — omitting it fails the command, not the
  # request. And under --dry-run, main.go turns ANY command error into a
  # logged warning plus exit 0 ("Encountered an error but --dry-run is
  # enabled"), so a missing/invalid token here would silently report this
  # case as a pass without exercising fingerprinting or git resolution at
  # all. Guard against that below rather than trusting the exit code alone.
  local output
  output="$(docker run --rm \
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
    --debug 2>&1)"
  local status=$?
  echo "$output"

  [ "$status" -eq 0 ] && ! grep -q "Encountered an error but --dry-run is enabled" <<< "$output"
}

# --- Run all cases ------------------------------------------------------
# Add a case by adding one entry here alongside its test_* function above.

CASES=(
  "list-environments:test_list_environments"
  "attest-artifact-dir:test_attest_artifact_dir"
)

# Seed every case as not-run and flush before running any of them, so an
# abort part-way through (crash, timeout, hung docker run) leaves a results
# file that visibly distinguishes "didn't run" from "passed".
for entry in "${CASES[@]}"; do
  RESULTS="$(jq --arg name "${entry%%:*}" '. + [{name: $name, outcome: "not-run"}]' <<< "$RESULTS")" \
    || { echo "jq failed" >&2; exit 1; }
done
printf '%s\n' "$RESULTS" > "$RESULTS_FILE" || { echo "failed to write $RESULTS_FILE" >&2; exit 1; }

for entry in "${CASES[@]}"; do
  run_case "${entry%%:*}" "${entry##*:}"
done

exit $EXIT_CODE
