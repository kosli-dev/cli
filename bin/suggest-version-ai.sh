#!/usr/bin/env bash
# Suggest next semver and changelog by sending the git diff to Claude.
# Does not rely on commit messages. Changelog is suitable for GoReleaser --release-notes.
#
# Auth (first non-empty wins):
#   - ANTHROPIC_API_KEY: call Claude directly.
#   - OP_ANTHROPIC_API_KEY_REF: 1Password reference (default below; override if your item path differs).
#
# Optional: CLAUDE_MODEL (default: claude-sonnet-4-6) — e.g. claude-opus-4-6.
#
# Requires: curl, jq; for 1Password: op CLI
# Usage: bin/suggest-version-ai.sh [base_ref] [-o release_notes.md]
#   base_ref defaults to the latest git tag.
#   -o FILE  write changelog markdown to FILE (default: dist/release_notes.md)
#
# Output: bump (major|minor|patch), next_version (e.g. v1.3.0), and changelog file.

set -euo pipefail

BASE_REF=""
RELEASE_NOTES_FILE="dist/release_notes.md"
while [[ $# -gt 0 ]]; do
  case "$1" in
    -o) RELEASE_NOTES_FILE="$2"; shift 2 ;;
    *)  BASE_REF="$1"; shift ;;
  esac
done
BASE_REF="${BASE_REF:-$(git describe --tags --abbrev=0 2>/dev/null)}"
SUGGESTED_VERSION_FILE="$(dirname "$RELEASE_NOTES_FILE")/suggested_version"

if [ -z "$BASE_REF" ]; then
  echo "ERROR: No base ref. Pass a tag or branch, or create a tag first." >&2
  exit 1
fi

# Cap diff size to stay within context
MAX_DIFF_CHARS=50000
DIFF="$(git diff "$BASE_REF"..HEAD 2>/dev/null | head -c "$MAX_DIFF_CHARS")"

# Get API key from 1Password if not set (default ref; override with OP_ANTHROPIC_API_KEY_REF)
OP_ANTHROPIC_API_KEY_REF="${OP_ANTHROPIC_API_KEY_REF:-op://Shared/Anthropic API Key/credential}"
if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
  if command -v op >/dev/null 2>&1; then
    ANTHROPIC_API_KEY=$(op read "$OP_ANTHROPIC_API_KEY_REF" 2>/dev/null) || true
  fi
fi
if [ -z "${ANTHROPIC_API_KEY:-}" ]; then
  echo "ERROR: Set ANTHROPIC_API_KEY or OP_ANTHROPIC_API_KEY_REF (1Password)." >&2
  exit 1
fi

# Remove stale outputs from a previous run so a failure partway through doesn't mislead the next invocation.
trap 'rm -f "$RELEASE_NOTES_FILE" "$SUGGESTED_VERSION_FILE"' EXIT

if [ -z "$DIFF" ]; then
  echo "No changes since $BASE_REF. Bump: patch (no change)." >&2
  CHANGELOG="No code changes since $BASE_REF."
  mkdir -p "$(dirname "$RELEASE_NOTES_FILE")"
  echo "$CHANGELOG" > "$RELEASE_NOTES_FILE"
  echo "patch"
  CURRENT="${BASE_REF#v}"
  if [[ "$CURRENT" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
    MAJOR="${CURRENT%%.*}"; REST="${CURRENT#*.}"; MINOR="${REST%%.*}"; PATCH="${REST#*.}"; PATCH="${PATCH%%[-+]*}"
    NEXT="v${MAJOR}.${MINOR}.$((PATCH+1))"
    if [ -n "$(git tag -l "$NEXT")" ]; then
      echo "ERROR: tag $NEXT already exists. Push it or use: make release tag=$NEXT" >&2
      exit 1
    fi
    echo "$NEXT" > "$SUGGESTED_VERSION_FILE"
    echo "$NEXT"
  fi
  trap - EXIT
  exit 0
fi

PROMPT="You are a release engineer. Given the following git diff for a CLI application (Kosli CLI), do two things.

Scope: Consider ONLY changes to the CLI itself—i.e. code under cmd/ and internal/ that affects user-facing commands, flags, and behavior. IGNORE all other changes when deciding the version and when writing the changelog:
- Ignore: documentation (docs*, *.md), Helm charts (charts/), CI/workflows (.github/), scripts (bin/, scripts/), tests (*_test.go, testdata/), Makefile, config files, and any other non-CLI code.
- If the diff contains only ignored changes, recommend a patch bump and write a single short line for the changelog (e.g. \"No user-facing CLI changes.\").

1) Suggest the semantic version bump (based only on CLI changes):
   - major: Breaking changes (removed/renamed commands or flags, changed default behavior).
   - minor: New commands, flags, subcommands, or features.
   - patch: Bug fixes, refactors, internal or dependency updates; or no user-facing CLI changes.

2) Write a short changelog in markdown for the GitHub release body. Include only user-facing CLI changes. Use bullet points; be concise; no preamble.
   - Structure the changelog with section headers (e.g. \"# Breaking changes\", \"# New features\", \"# Bug fixes\" or \"# Improvements\") and list items under each header. Use only headers that have at least one change—omit any section that would be empty.
   - Do not write placeholder lines under any header (no \"No other changes\", \"No user-facing CLI changes in this release\", or similar). If there are no CLI changes at all, output a single short line only (no headers).

Reply in this exact format (no other text before or after):
BUMP: major|minor|patch
---CHANGELOG---
<markdown changelog here>"

CLAUDE_MODEL="${CLAUDE_MODEL:-claude-sonnet-4-6}"
DIFF_FILE=$(mktemp)
trap 'rm -f "$DIFF_FILE" "$RELEASE_NOTES_FILE" "$SUGGESTED_VERSION_FILE"' EXIT
printf '%s' "$DIFF" > "$DIFF_FILE"
BODY=$(jq -n \
  --arg model "$CLAUDE_MODEL" \
  --arg prompt "$PROMPT" \
  --rawfile diff "$DIFF_FILE" \
  '{model: $model, max_tokens: 1024, messages: [{role: "user", content: ($prompt + "\n\n--- diff ---\n\n" + $diff)}]}')

RESPONSE=$(curl -s -S -X POST "https://api.anthropic.com/v1/messages" \
  -H "x-api-key: $ANTHROPIC_API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "Content-Type: application/json" \
  -d "$BODY")

CONTENT=$(echo "$RESPONSE" | jq -r '.content[0].text // empty')
if [ -z "$CONTENT" ]; then
  echo "ERROR: Anthropic API failed or returned no content. Response:" >&2
  echo "$RESPONSE" | jq . >&2
  exit 1
fi

BUMP=$(echo "$CONTENT" | tr '[:upper:]' '[:lower:]' | grep -oE 'major|minor|patch' | head -1)
case "$BUMP" in
  major|minor|patch) ;;
  *)
    echo "WARN: Could not parse bump (got: $CONTENT). Defaulting to patch." >&2
    BUMP=patch
    ;;
esac

CHANGELOG_MARKER='---CHANGELOG---'
if echo "$CONTENT" | grep -qF -- "$CHANGELOG_MARKER"; then
  CHANGELOG=$(echo "$CONTENT" | sed -n "/${CHANGELOG_MARKER}/,\$ p" | tail -n +2)
else
  CHANGELOG=$(echo "$CONTENT" | sed -n '2,$ p')
fi
mkdir -p "$(dirname "$RELEASE_NOTES_FILE")"
echo "$CHANGELOG" > "$RELEASE_NOTES_FILE"

# Compute next version from current tag
CURRENT="${BASE_REF#v}"
if [[ "$CURRENT" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
  MAJOR="${CURRENT%%.*}"; REST="${CURRENT#*.}"; MINOR="${REST%%.*}"; PATCH="${REST#*.}"; PATCH="${PATCH%%[-+]*}"
  case "$BUMP" in
    major) NEXT="v$((MAJOR+1)).0.0" ;;
    minor) NEXT="v${MAJOR}.$((MINOR+1)).0" ;;
    patch) NEXT="v${MAJOR}.${MINOR}.$((PATCH+1))" ;;
  esac
else
  NEXT=""
fi

if [ -n "$NEXT" ]; then
  if [ -n "$(git tag -l "$NEXT")" ]; then
    echo "ERROR: tag $NEXT already exists. Push it or use: make release tag=$NEXT" >&2
    exit 1
  fi
  echo "$NEXT" > "$SUGGESTED_VERSION_FILE"
fi
echo "Suggested bump: $BUMP (from diff since $BASE_REF)" >&2
echo "Next version:  $NEXT" >&2
echo "Changelog:     $RELEASE_NOTES_FILE" >&2
trap - EXIT
echo "$BUMP"
echo "$NEXT"
