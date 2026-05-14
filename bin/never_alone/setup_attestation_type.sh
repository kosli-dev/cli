#!/usr/bin/env bash
set -euo pipefail

# Use first arg, existing ENV or set a default
KOSLI_ORG="${1:-${KOSLI_ORG:-kosli-public}}"

# One-time setup: create custom attestation types for never-alone.
# Run this after any schema change. Types cannot be updated in place;
# delete via the Kosli UI/API first if re-creating an existing type.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [[ -z "${KOSLI_API_TOKEN:-}" ]]; then
  echo "ERROR: KOSLI_API_TOKEN is not set" >&2
  exit 1
fi

echo "Creating four-eyes-result attestation type (release-level policy evaluation result)..."
kosli create attestation-type four-eyes-result \
  --description "Four-eyes policy evaluation result for a release commit range (never-alone)" \
  --schema "${SCRIPT_DIR}/four-eyes-result-schema.json" \
  --jq ".allow == true" \
  --org "${KOSLI_ORG}"

echo "Done — four-eyes-result attestation type ready."
