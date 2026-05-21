#!/usr/bin/env bash
# Walk through the explainable-evaluation demos. Pauses between each one.
# Run from the repo root:  ./demos/explainable-evaluation/run.sh

set -u
DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT="$(cd "$DIR/../.." && pwd)"
KOSLI="$ROOT/kosli"

if [[ ! -x "$KOSLI" ]]; then
  echo "Build the CLI first:  make build" >&2
  exit 1
fi

pause() {
  echo
  read -rp "[enter] next demo, [q] quit > " key
  if [[ "$key" == "q" ]]; then exit 0; fi
  echo
}

heading() {
  printf '\n\033[1;36m── %s ──\033[0m\n\n' "$1"
}

run_decision() {
  local label="$1"; shift
  local input="$1"; shift
  local policy="$1"; shift
  heading "$label"
  echo "$ kosli evaluate input --decision $* \\"
  echo "    --input-file ${input#$ROOT/} \\"
  echo "    --policy    ${policy#$ROOT/}"
  echo
  "$KOSLI" evaluate input --decision "$@" --input-file "$input" --policy "$policy" || true
}

run_decision "Demo 1 — bakery, passing" \
  "$DIR/inputs/bakery-pass.json" "$DIR/policies/bakery.rego"
pause

run_decision "Demo 2 — bakery, failing (short-circuit)" \
  "$DIR/inputs/bakery-fail.json" "$DIR/policies/bakery.rego" --no-assert
pause

run_decision "Demo 3 — bakery, parameterised" \
  "$DIR/inputs/bakery-pass.json" "$DIR/policies/bakery-params.rego" \
  --params '{"min_temp_c": 175, "max_temp_c": 200, "min_minutes": 25, "max_minutes": 40}'
pause

run_decision "Demo 4 — iteration over input.batches" \
  "$DIR/inputs/batches-mixed.json" "$DIR/policies/batches.rego" --no-assert
pause

run_decision "Demo 5 — alternatives, passing" \
  "$DIR/inputs/pr-human-approved.json" "$DIR/policies/pr-approval.rego"
pause

run_decision "Demo 6 — alternatives, all failing" \
  "$DIR/inputs/pr-no-approval.json" "$DIR/policies/pr-approval.rego" --no-assert
pause

run_decision "Demo 7 — Kosli-shaped: iterating trails with multi-def per-trail check" \
  "$DIR/inputs/trails.json" "$DIR/policies/scr-trails.rego" --no-assert

echo
echo "Done."
