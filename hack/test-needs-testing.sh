#!/usr/bin/env bash
# Exercises needs-testing.sh with lists of changed paths.
set -uo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
SCRIPT="${HERE}/needs-testing.sh"

pass=0; fail=0
check() {
  if [ "$1" = "$2" ]; then pass=$((pass+1)); echo "  ok    $3"
  else fail=$((fail+1)); echo "  FAIL  $3"; echo "        wanted: $2"; echo "        got:    $1"; fi
}
classify() { printf '%s' "$1" | "${SCRIPT}"; }

echo "changes that never reach a build or a test"
check "$(classify $'README.md\nCONTRIBUTING.md\nLICENSE\n')" "false" "top-level docs"
check "$(classify $'docs/adr/0001.md\ndesign-docs/x.md\n')" "false" "docs folders"
check "$(classify $'.claude/skills/new-command/SKILL.md\n')" "false" "Claude config"
check "$(classify $'charts/k8s-reporter/values.yaml\n')" "false" "the Helm chart, which has its own workflow"
check "$(classify $'.github/ISSUE_TEMPLATE/bug_report.yml\n.github/labeler.yaml\n')" "false" "GitHub config"
check "$(classify $'.github/workflows/release.yml\n.github/workflows/pr-checks.yml\n')" "false" "workflows outside the main pipeline"
check "$(classify $'hack/empty-flag-audit/docs/2026-08-13-empty-value-decision.md\n')" "false" "audit notes"

echo "changes that need the tests"
check "$(classify $'cmd/kosli/root.go\n')" "true" "Go code"
check "$(classify $'README.md\ngo.mod\n')" "true" "docs alongside a code change"
check "$(classify $'cmd/kosli/testdata/policies/bundle/README.md\n')" "true" "markdown a test reads"
check "$(classify $'internal/docs/x.go\n')" "true" "a docs folder below the root"
check "$(classify $'.github/workflows/main.yml\n')" "true" "the main pipeline itself"
check "$(classify $'.github/workflows/test.yml\n')" "true" "a workflow the main pipeline runs"
check "$(classify $'hack/needs-testing.sh\n')" "true" "this classifier"
check "$(classify '')" "true" "no paths at all, so nothing is known"

echo "a list longer than a pipe holds"
out=$(seq 100000 | sed 's|^|cmd/kosli/f|' | "${SCRIPT}"); rc=$?
check "${rc}" "0" "reads every path, so the command feeding it is not cut off"
check "${out}" "true" "still answers true"

echo
echo "${pass} passed, ${fail} failed"
[ "${fail}" -eq 0 ]
