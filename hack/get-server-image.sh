#!/bin/bash
set -uo pipefail

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

# Get snapshot JSON from kosli
json=$(kosli get snapshot staging-aws --org kosli -a ${KOSLI_API_TOKEN_PROD} --output json)

# Extract and format the desired artifact
echo "$json" | jq -r '
  .artifacts[]
  | select(.name | test("merkely:"))
  | select(.annotation.type != "exited")
  | "\(.name | sub(":.*"; ""))@sha256:\(.fingerprint)"
'