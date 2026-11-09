#!/usr/bin/env bash
# Reads changed paths, one per line, and prints "false" when every one of them
# never reaches a build or a test, otherwise "true". No paths prints "true",
# since then nothing is known about the change.
set -euo pipefail

skippable=unknown
while IFS= read -r path; do
  [ -z "${path}" ] && continue
  case "${path}" in
    # The workflows the main pipeline runs, so a change to one is tested by running it.
    .github/workflows/main.yml | .github/workflows/test.yml | \
      .github/workflows/docker.yml | .github/workflows/init_kosli.yml)
      skippable=no ;;
    README.md | CONTRIBUTING.md | CODE_OF_CONDUCT.md | SECURITY.md | CLAUDE.md | \
      AUTHORS | LICENSE | TODO.md | dev-guide.md | release-guide.md | \
      docs/* | design-docs/* | .claude/* | charts/* | .github/* | \
      hack/empty-flag-audit/docs/*)
      [ "${skippable}" = unknown ] && skippable=yes ;;
    *)
      skippable=no ;;
  esac
done

if [ "${skippable}" = yes ]; then echo false; else echo true; fi
