#!/usr/bin/env bash
# Fails if any packaged chart's default image is not under ghcr.io/kosli-dev/.
#
# `image.repository` in a chart's values.yaml names the registry account whose
# code every install of that chart runs. It must be ours. A typo, or a chart
# copied from an old release, could point it at an account we don't control.
# Every step fails closed: no charts, an unreadable chart, a chart with no
# values.yaml, or a values.yaml whose image.repository is missing or not a single
# ghcr.io/kosli-dev/ reference is an error, not a pass.
set -euo pipefail

dir="${1:?usage: $0 <directory of packaged charts>}"
# Whole value, one token: a multi-document or block-scalar values.yaml makes yq
# print several lines, and a per-line prefix match would pass if any one matched.
allowed='^ghcr\.io/kosli-dev/[^[:space:]]+$'

shopt -s nullglob
charts=("$dir"/*.tgz)
if [ "${#charts[@]}" -eq 0 ]; then
  echo "no packaged charts (*.tgz) in $dir" >&2
  exit 1
fi

status=0
for chart in "${charts[@]}"; do
  listing=$(tar -tzf "$chart") || {
    echo "$chart: cannot read the chart archive" >&2
    status=1
    continue
  }
  # helm package puts values.yaml one level down, under the chart's name.
  member=$(printf '%s\n' "$listing" | grep -E '^[^/]+/values\.yaml$' || true)
  if [ -z "$member" ]; then
    echo "$chart: no values.yaml in the chart" >&2
    status=1
    continue
  fi
  repo=$(tar -xzOf "$chart" "$member" | yq '.image.repository') || {
    echo "$chart: failed to read image.repository from $member" >&2
    status=1
    continue
  }
  if ! [[ $repo =~ $allowed ]]; then
    echo "$chart: image.repository is $repo, must be one image reference under ghcr.io/kosli-dev/" >&2
    status=1
    continue
  fi
  echo "$chart: image.repository is $repo"
done
exit "$status"
