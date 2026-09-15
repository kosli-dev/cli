#!/bin/bash
set -uo pipefail

if [ $# -lt 1 ]; then
  echo "Output result file is missing" >&2
  exit 1
fi

OUTPUT_FILE=$1; shift

# Check that jq is installed
if ! command -v jq &> /dev/null; then
  echo "❌ Error: 'jq' is not installed. Please install it:" >&2
  echo "    macOS: brew install jq" >&2
  echo "    Debian/Ubuntu: sudo apt install jq" >&2
  exit 1
fi

# Check if KOSLI_API_TOKEN_PROD is set, if not prompt for it
if [[ -z "${KOSLI_API_TOKEN_PROD:-}" ]]; then
  printf "Enter KOSLI_API_TOKEN_PROD: " >&2
  read -s KOSLI_API_TOKEN_PROD
  echo "" >&2  # new line after silent read
  export KOSLI_API_TOKEN_PROD
fi

# Now that we have the token, we can set -e
set -e

# The environment reports one running server most of the time, but during a deploy it
# briefly reports the old and the new one together. Picking either is guesswork: the
# artifact's creationTimestamp is a list of per-instance start times, so it says when
# instances started, not which artifact supersedes which. Writing both is what produced
# "invalid reference format" from docker on run 34970128696. So wait for it to settle,
# and say so plainly if it does not.
ATTEMPTS=${ATTEMPTS:-10}
INTERVAL=${INTERVAL:-30}

REASONS=""

note_reason() {
  case "${REASONS}" in
    *"$1"*) ;;
    *) REASONS="${REASONS}  - $1"$'\n' ;;
  esac
}

# A snapshot query can fail transiently, and a deploy is exactly when that is likely, so a
# failed query has to retry rather than abort. This is called as an if-condition, which
# suppresses errexit for the whole call, so every step checks its own status and records a
# reason with note_reason. The final message then names what actually went wrong.
attempt_lookup() {
  local json refs rc errfile
  # stderr goes to its own file rather than into the value. The CLI writes warnings there
  # on successful calls, and merging them would corrupt the JSON on a call that worked.
  errfile=$(mktemp)
  json=$(kosli get snapshot staging-aws --org kosli -a "${KOSLI_API_TOKEN_PROD}" --output json 2>"${errfile}")
  rc=$?
  if [ "$rc" -ne 0 ]; then
    note_reason "the snapshot query failed"
    echo "query failed: $(cat "${errfile}")" >&2
    rm -f "${errfile}"
    return 1
  fi
  rm -f "${errfile}"

  refs=$(printf '%s' "$json" | jq -r '
    .artifacts[]
    | select(.name | test("merkely:"))
    | select(.annotation.type != "exited")
    | "\(.name | sub(":.*"; ""))@sha256:\(.fingerprint)"
  ' | sort -u)
  rc=$?
  if [ "$rc" -ne 0 ]; then
    note_reason "the snapshot response could not be read"
    echo "could not read the snapshot response; jq exited ${rc}" >&2
    return 1
  fi

  # Trailing newline kept so the file matches what the previous jq redirect produced and
  # is a well-formed text file. The Makefile reads it with $(cat ...), which strips it
  # either way, but other readers may not.
  printf '%s\n' "$refs"
}

# Written to a scratch file and moved into place only on success, so a failed run cannot
# leave a value behind for the next one to pick up. The Makefile guards on the file
# existing, not on it being usable.
SCRATCH=$(mktemp)
trap 'rm -f "${SCRATCH}"' EXIT

for attempt in $(seq 1 "${ATTEMPTS}"); do
  if attempt_lookup > "${SCRATCH}"; then
    count=$(grep -c . "${SCRATCH}" || true)
    if [ "${count}" -eq 1 ]; then
      mv "${SCRATCH}" "${OUTPUT_FILE}"
      trap - EXIT
      exit 0
    fi
    note_reason "the environment reported ${count} running server images, expected 1"
    echo "attempt ${attempt}/${ATTEMPTS}: the environment reports ${count} running server images, expected 1" >&2
    sed 's/^/    /' "${SCRATCH}" >&2
  fi
  if [ "${attempt}" -lt "${ATTEMPTS}" ]; then
    sleep "${INTERVAL}"
  fi
done

rm -f "${OUTPUT_FILE}"
cat >&2 <<MSG

Could not determine which server image to test against.

Over ${ATTEMPTS} attempts ${INTERVAL} seconds apart, what happened was:

${REASONS}
Two running servers usually means a deploy is in progress, in which case re-run
the job once it has finished. Anything else above is not a deploy and wants
looking at.

Nothing was written, so a later run will look again rather than reuse a bad value.

MSG
exit 1
