#!/usr/bin/env bash
# Renders the k8s-reporter chart with different values and asserts on the CronJob.
# Uses only scalar yq paths, which mikefarah yq (CI) and python yq read alike.
set -euo pipefail

chart="$(cd "$(dirname "$0")/.." && pwd)/charts/k8s-reporter"
pod='.spec.jobTemplate.spec.template.spec'
container="$pod.containers[0]"
status=0

render() {
  helm template test "$chart" \
    --show-only templates/cronjob.yaml \
    --set reporterConfig.kosliOrg=placeholder \
    --set 'reporterConfig.environments[0].name=placeholder' \
    "$@"
}

# assert <description> <yq path> <expected> [helm args...]
assert() {
  local description="$1" path="$2" expected="$3" actual
  shift 3
  actual=$(render "$@" | yq -r "$path") || actual="<render failed>"
  if [ "$actual" = "$expected" ]; then
    echo "ok   $description"
  else
    echo "FAIL $description: $path is '$actual', expected '$expected'" >&2
    status=1
  fi
}

assert "default container context forbids privilege escalation" \
  "$container.securityContext.allowPrivilegeEscalation" false
assert "default container context runs as non root" \
  "$container.securityContext.runAsNonRoot" true
assert "default container context runs as user 1000" \
  "$container.securityContext.runAsUser" 1000
assert "default pod has no pod-level security context" \
  "$pod.securityContext" null

assert "container context passes capabilities through" \
  "$container.securityContext.capabilities.drop[0]" ALL \
  --set 'reporterConfig.securityContext.capabilities.drop[0]=ALL'
assert "container context passes seccompProfile through" \
  "$container.securityContext.seccompProfile.type" RuntimeDefault \
  --set reporterConfig.securityContext.seccompProfile.type=RuntimeDefault
assert "container context passes readOnlyRootFilesystem through" \
  "$container.securityContext.readOnlyRootFilesystem" true \
  --set reporterConfig.securityContext.readOnlyRootFilesystem=true
assert "null runAsUser drops the field" \
  "$container.securityContext.runAsUser" null \
  --set reporterConfig.securityContext.runAsUser=null
assert "null runAsUser keeps the other defaults" \
  "$container.securityContext.runAsNonRoot" true \
  --set reporterConfig.securityContext.runAsUser=null
assert "null securityContext drops the container context" \
  "$container.securityContext" null \
  --set reporterConfig.securityContext=null

assert "podSecurityContext renders at pod level" \
  "$pod.securityContext.fsGroup" 2000 \
  --set podSecurityContext.fsGroup=2000
assert "podSecurityContext passes seccompProfile through" \
  "$pod.securityContext.seccompProfile.type" RuntimeDefault \
  --set podSecurityContext.seccompProfile.type=RuntimeDefault
assert "fsGroup under reporterConfig.securityContext moves to pod level" \
  "$pod.securityContext.fsGroup" 3000 \
  --set reporterConfig.securityContext.fsGroup=3000
assert "fsGroup under reporterConfig.securityContext stays out of the container" \
  "$container.securityContext.fsGroup" null \
  --set reporterConfig.securityContext.fsGroup=3000
assert "podSecurityContext.fsGroup wins over reporterConfig.securityContext.fsGroup" \
  "$pod.securityContext.fsGroup" 2000 \
  --set podSecurityContext.fsGroup=2000 \
  --set reporterConfig.securityContext.fsGroup=3000

exit $status
