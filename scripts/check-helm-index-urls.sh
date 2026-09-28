#!/usr/bin/env bash
# Fails if any chart download URL in a Helm repo index is not on charts.kosli.com.
#
# `helm repo index --merge` copies old entries forward unchanged, so a stale URL
# pointing at a host we no longer own would be republished on every release
# (kosli-dev/server#7024). Every step fails closed: an unreadable index or an
# index with no URLs is an error, not a pass.
set -euo pipefail

index="${1:?usage: $0 <index.yaml>}"

urls=$(yq '.entries[][].urls[]' "$index") || {
  echo "failed to read chart URLs from $index" >&2
  exit 1
}
if [ -z "$urls" ]; then
  echo "no chart URLs found in $index" >&2
  exit 1
fi

bad=$(printf '%s\n' "$urls" | grep -v '^https://charts\.kosli\.com/' || true)
if [ -n "$bad" ]; then
  echo "chart URLs not on https://charts.kosli.com/:" >&2
  printf '%s\n' "$bad" >&2
  exit 1
fi

echo "all $(printf '%s\n' "$urls" | wc -l | tr -d ' ') chart URLs are on https://charts.kosli.com/"
