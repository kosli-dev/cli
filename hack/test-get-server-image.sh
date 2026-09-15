#!/usr/bin/env bash
# Exercises every path through get-server-image.sh with a stubbed kosli on PATH.
#
# The failure paths sleep between attempts, so they are only testable because ATTEMPTS and
# INTERVAL are overridable. Without that this file could not exist: ten real attempts at
# thirty seconds is four and a half minutes per case.
set -uo pipefail

HERE=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
SCRIPT="${HERE}/get-server-image.sh"
WORK=$(mktemp -d)
trap 'rm -rf "${WORK}"' EXIT
mkdir -p "${WORK}/bin"

pass=0; fail=0
stub() { printf '#!/usr/bin/env bash\n%s\n' "$1" > "${WORK}/bin/kosli"; chmod +x "${WORK}/bin/kosli"; }
run() {
  ATTEMPTS=3 INTERVAL=0 KOSLI_API_TOKEN_PROD=dummy \
    PATH="${WORK}/bin:${PATH}" "${SCRIPT}" "${WORK}/out.txt" >"${WORK}/stdout" 2>"${WORK}/stderr"
}
check() {
  if [ "$1" = "$2" ]; then pass=$((pass+1)); echo "  ok    $3"
  else fail=$((fail+1)); echo "  FAIL  $3"; echo "        wanted: $2"; echo "        got:    $1"; fi
}

ONE='cat <<J
{"artifacts":[{"name":"reg/merkely:a","fingerprint":"aaa","annotation":{"type":"unchanged"}}]}
J'
TWO='cat <<J
{"artifacts":[{"name":"reg/merkely:a","fingerprint":"aaa","annotation":{"type":"unchanged"}},{"name":"reg/merkely:b","fingerprint":"bbb","annotation":{"type":"unchanged"}}]}
J'
SAME='cat <<J
{"artifacts":[{"name":"reg/merkely:a","fingerprint":"aaa","annotation":{"type":"unchanged"}},{"name":"reg/merkely:a","fingerprint":"aaa","annotation":{"type":"unchanged"}}]}
J'

echo "one running server"
rm -f "${WORK}/out.txt"; stub "$ONE"; run; rc=$?
check "$rc" "0" "exits zero"
check "$(cat "${WORK}/out.txt")" "reg/merkely@sha256:aaa" "writes the one image"
check "$(wc -l < "${WORK}/out.txt" | tr -d ' ')" "1" "one line, with a trailing newline"

echo "one running server, with a warning on stderr"
rm -f "${WORK}/out.txt"; stub 'echo "[warning] benign" >&2
'"$ONE"; run; rc=$?
check "$rc" "0" "a warning does not corrupt the response"
check "$(cat "${WORK}/out.txt")" "reg/merkely@sha256:aaa" "still writes the one image"

echo "two records naming the same image"
rm -f "${WORK}/out.txt"; stub "$SAME"; run; rc=$?
check "$rc" "0" "counts as settled, not as two"

echo "two different servers, a deploy"
echo "stale" > "${WORK}/out.txt"; stub "$TWO"; run; rc=$?
check "$rc" "1" "exits non-zero"
check "$(test -f "${WORK}/out.txt" && echo present || echo absent)" "absent" "removes a stale output file"
check "$(grep -c 'reported 2 running' "${WORK}/stderr")" "1" "names two running servers"

echo "no running server"
rm -f "${WORK}/out.txt"; stub 'echo "{\"artifacts\":[]}"'; run; rc=$?
check "$rc" "1" "exits non-zero"
check "$(grep -c 'reported 0 running' "${WORK}/stderr")" "1" "names zero running servers"

echo "the query fails"
rm -f "${WORK}/out.txt"; stub 'echo "503 service unavailable" >&2; exit 1'; run; rc=$?
check "$rc" "1" "exits non-zero"
check "$(grep -c 'the snapshot query failed' "${WORK}/stderr")" "1" "blames the query, not a deploy"

echo "the response is not json"
rm -f "${WORK}/out.txt"; stub 'echo "not json at all"'; run; rc=$?
check "$rc" "1" "exits non-zero"
check "$(grep -c 'could not be read' "${WORK}/stderr")" "1" "blames the response, not a deploy"

echo
echo "${pass} passed, ${fail} failed"
[ "${fail}" -eq 0 ]
