#!/usr/bin/env bash
# Suggest next semver and changelog by sending the git diff to Claude.
# Does not rely on commit messages. Changelog is suitable for GoReleaser --release-notes.
#
# Auth (first non-empty wins):
#   - ANTHROPIC_API_KEY: call Claude directly.
#   - OP_ANTHROPIC_API_KEY_REF: 1Password reference (default below; override if your item path differs).
#
# Requires: curl, jq; for 1Password: op CLI
# Usage: bin/suggest-version-ai.sh [base_ref] [-o release_notes.md]
#   base_ref defaults to the latest git tag.
#   -o FILE  write changelog markdown to FILE (default: dist/release_notes.md)
#
# Output: bump (major|minor|patch), next_version (e.g. v1.3.0), and changelog file.

set -e

BASE_REF=""
RELEASE_NOTES_FILE="dist/release_notes.md"
while [[ $# -gt 0 ]]; do
  case "$1" in
    -o) RELEASE_NOTES_FILE="$2"; shift 2 ;;
    *)  BASE_REF="$1"; shift ;;
  esac
done
BASE_REF="${BASE_REF:-$(git describe --tags --abbrev=0 2>/dev/null)}"

if [ -z "$BASE_REF" ]; then
  echo "ERROR: No base ref. Pass a tag or branch, or create a tag first." >&2
  exit 1
fi

# Cap diff size to stay within context
MAX_DIFF_CHARS=50000
DIFF="$(git diff "$BASE_REF"..HEAD 2>/dev/null | head -c "$MAX_DIFF_CHARS")"

# Get API key from 1Password if not set (default ref; override with OP_ANTHROPIC_API_KEY_REF)
OP_ANTHROPIC_API_KEY_REF="${OP_ANTHROPIC_API_KEY_REF:-op://Shared/Anthropic API Key/credential}"
if [ -z "$ANTHROPIC_API_KEY" ]; then
  if command -v op >/dev/null 2>&1; then
    ANTHROPIC_API_KEY=$(op read "$OP_ANTHROPIC_API_KEY_REF" 2>/dev/null) || true
  fi
fi
if [ -z "$ANTHROPIC_API_KEY" ]; then
  echo "ERROR: Set ANTHROPIC_API_KEY or OP_ANTHROPIC_API_KEY_REF (1Password). See scripts/README-release-suggest.md." >&2
  exit 1
fi

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
    echo "$NEXT" > "$(dirname "$RELEASE_NOTES_FILE")/suggested_version"
    echo "$NEXT"
  fi
  exit 0
fi

PROMPT="You are a release engineer. Given the following git diff for a CLI application (Kosli CLI), do two things.

1) Suggest the semantic version bump:
   - major: Breaking changes (removed/renamed commands or flags, changed default behavior).
   - minor: New commands, flags, subcommands, or features.
   - patch: Bug fixes, docs, refactors, internal or dependency changes.

2) Write a short changelog in markdown for the GitHub release body. Use bullet points; be concise; no preamble.

Reply in this exact format (no other text before or after):
BUMP: major|minor|patch
---CHANGELOG---
<markdown changelog here>"

BODY=$(jq -n \
  --arg prompt "$PROMPT" \
  --rawfile diff - \
  '{model: "claude-sonnet-4-20250514", max_tokens: 1024, messages: [{role: "user", content: ($prompt + "\n\n--- diff ---\n\n" + $diff)}]}' \
  < <(printf '%s' "$DIFF"))

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

if echo "$CONTENT" | grep -q '---CHANGELOG---'; then
  CHANGELOG=$(echo "$CONTENT" | sed -n '/---CHANGELOG---/,$ p' | tail -n +2)
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

[ -n "$NEXT" ] && echo "$NEXT" > "$(dirname "$RELEASE_NOTES_FILE")/suggested_version"
echo "Suggested bump: $BUMP (from diff since $BASE_REF)" >&2
echo "Next version:  $NEXT" >&2
echo "Changelog:     $RELEASE_NOTES_FILE" >&2
echo "$BUMP"
echo "$NEXT"
