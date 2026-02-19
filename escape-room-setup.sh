#!/usr/bin/env bash
# =============================================================================
# Knowledge Transfer Game Day — Snapshot Escape Room
# Setup & Cleanup Scripts for Docker Desktop Kubernetes
# =============================================================================
#
# PREREQUISITES:
#   - Docker Desktop with Kubernetes enabled
#   - kubectl, helm CLI installed
#   - The kosli-dev/cli repo cloned locally
#
# USAGE:
#   ./escape-room-setup.sh setup-all     # Set up all 4 rooms
#   ./escape-room-setup.sh setup-room N  # Set up room N (1-4)
#   ./escape-room-setup.sh teardown      # Clean up everything
#   ./escape-room-setup.sh verify        # Verify prerequisites
#
# =============================================================================

set -euo pipefail

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
KUBE_CONTEXT="docker-desktop"
CHART_PATH="./charts/k8s-reporter"  # Relative to the kosli-dev/cli repo root
DUMMY_API_TOKEN="gameday-dummy-token-not-real"
KOSLI_ORG="gameday-acme"
KOSLI_ENV="gameday-prod"

# Namespaces used across rooms
NS_APP_TEAM="app-team"
NS_TEAM_ALPHA="team-alpha"
NS_TEAM_BETA="team-beta"
NS_DEFAULT="default"
NS_ROOM3="room3-trust-no-tag"
NS_ROOM4="room4-reporters"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
info()    { echo -e "${CYAN}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[OK]${NC}   $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC} $*"; }
err()     { echo -e "${RED}[ERR]${NC}  $*"; }

switch_context() {
  info "Switching kubectl context to ${KUBE_CONTEXT}..."
  kubectl config use-context "${KUBE_CONTEXT}" >/dev/null 2>&1 || {
    err "Could not switch to context '${KUBE_CONTEXT}'."
    err "Make sure Docker Desktop is running and Kubernetes is enabled."
    exit 1
  }
  success "Context set to ${KUBE_CONTEXT}"
}

ensure_namespace() {
  local ns="$1"
  if ! kubectl get namespace "${ns}" >/dev/null 2>&1; then
    kubectl create namespace "${ns}" >/dev/null
    info "Created namespace: ${ns}"
  fi
}

create_kosli_secret() {
  local ns="$1"
  if ! kubectl get secret kosli-api-token -n "${ns}" >/dev/null 2>&1; then
    kubectl create secret generic kosli-api-token \
      --from-literal=key="${DUMMY_API_TOKEN}" \
      -n "${ns}" >/dev/null
    info "Created kosli-api-token secret in ${ns}"
  fi
}

deploy_dummy_pods() {
  local ns="$1"
  local count="${2:-3}"
  info "Deploying ${count} dummy pods in namespace ${ns}..."
  for i in $(seq 1 "${count}"); do
    kubectl run "demo-app-${i}" \
      --image=nginx:1.25-alpine \
      --namespace="${ns}" \
      --labels="app=demo,room=escape-room" \
      --restart=Never \
      --overrides='{"spec":{"terminationGracePeriodSeconds":1}}' \
      >/dev/null 2>&1 || true
  done
  # Wait for pods to be running
  kubectl wait --for=condition=Ready pod -l app=demo -n "${ns}" --timeout=60s >/dev/null 2>&1 || true
  success "Dummy pods running in ${ns}"
}

# ---------------------------------------------------------------------------
# Verify Prerequisites
# ---------------------------------------------------------------------------
verify_prereqs() {
  info "Verifying prerequisites..."
  local missing=0

  for cmd in kubectl helm docker; do
    if command -v "${cmd}" >/dev/null 2>&1; then
      success "${cmd} found: $(command -v "${cmd}")"
    else
      err "${cmd} not found — please install it."
      missing=1
    fi
  done

  # Check Docker Desktop K8S
  if kubectl config get-contexts "${KUBE_CONTEXT}" >/dev/null 2>&1; then
    success "Kubernetes context '${KUBE_CONTEXT}' exists"
  else
    err "Kubernetes context '${KUBE_CONTEXT}' not found."
    err "Enable Kubernetes in Docker Desktop → Settings → Kubernetes → Enable."
    missing=1
  fi

  # Check chart path
  if [ -f "${CHART_PATH}/Chart.yaml" ]; then
    success "Helm chart found at ${CHART_PATH}"
  else
    err "Chart not found at ${CHART_PATH}"
    err "Run this script from the root of the kosli-dev/cli repo."
    missing=1
  fi

  if [ "${missing}" -eq 1 ]; then
    err "Some prerequisites are missing. Fix the above and retry."
    exit 1
  fi
  echo
  success "All prerequisites met! Ready to set up escape rooms."
}

# =============================================================================
# ROOM 1 — "The Invisible Pods"
# =============================================================================
setup_room1() {
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  info "ROOM 1: The Invisible Pods 👻"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  ensure_namespace "${NS_APP_TEAM}"
  deploy_dummy_pods "${NS_APP_TEAM}" 3
  create_kosli_secret "${NS_DEFAULT}"

  info "Installing reporter in default namespace..."
  helm upgrade --install kosli-reporter-room1 "${CHART_PATH}" \
    --namespace "${NS_DEFAULT}" \
    --set reporterConfig.kosliOrg="${KOSLI_ORG}" \
    --set reporterConfig.kosliEnvironmentName="${KOSLI_ENV}" \
    --set reporterConfig.namespaces="${NS_APP_TEAM}" \
    --set reporterConfig.dryRun=true \
    --set serviceAccount.permissionScope=namespace \
    >/dev/null 2>&1

info "Starting reporter..."
kubectl create job -n default room1-manual-run --from=cronjob/kosli-reporter-room1-k8s-reporter >/dev/null 2>&1

info "Waiting for reporter to complete..."
kubectl wait --for=condition=complete --timeout=120s job/$(kubectl get jobs -n default --sort-by=.metadata.creationTimestamp -o jsonpath='{.items[-1:].metadata.name}') -n default >/dev/null 2>&1
success "Reporter completed!"

  success "Room 1 ready!"
  echo
  echo -e "  ${YELLOW}SCENARIO:${NC} Kosli shows 0 pods. The namespace '${NS_APP_TEAM}' definitely has running pods."
  echo -e "  ${YELLOW}INVESTIGATE:${NC}"
  echo -e "    kubectl get cronjob -n default"
  echo -e "    kubectl get pods -n ${NS_APP_TEAM}"
  echo -e "    kubectl get role,rolebinding -n default"
  echo -e '    kubectl logs -n default job/$(kubectl get jobs -n default --sort-by=.metadata.creationTimestamp -o jsonpath='"'"'{.items[-1:].metadata.name}'"'"') --tail=50'
  echo
}

# =============================================================================
# ROOM 2 — "The Regex Trap"
# =============================================================================
setup_room2() {
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  info "ROOM 2: The Regex Trap 🪤"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  ensure_namespace "${NS_TEAM_ALPHA}"
  ensure_namespace "${NS_TEAM_BETA}"
  deploy_dummy_pods "${NS_TEAM_ALPHA}" 2
  deploy_dummy_pods "${NS_TEAM_BETA}" 2
  create_kosli_secret "${NS_DEFAULT}"

  success "Room 2 ready!"
  echo
  echo -e "  ${YELLOW}SCENARIO:${NC} You want to monitor all namespaces matching team-.*"
  echo -e "  ${YELLOW}TRY RUNNING:${NC}"
  echo
  echo -e "  ${CYAN}helm upgrade --install kosli-reporter-room2 ${CHART_PATH} \\\\${NC}"
  echo -e "  ${CYAN}  --namespace default \\\\${NC}"
  echo -e "  ${CYAN}  --set reporterConfig.kosliOrg=${KOSLI_ORG} \\\\${NC}"
  echo -e "  ${CYAN}  --set reporterConfig.kosliEnvironmentName=${KOSLI_ENV} \\\\${NC}"
  echo -e "  ${CYAN}  --set reporterConfig.namespacesRegex=\"team-.*\" \\\\${NC}"
  echo -e "  ${CYAN}  --set reporterConfig.dryRun=true \\\\${NC}"
  echo -e "  ${CYAN}  --set serviceAccount.permissionScope=namespace${NC}"
  echo
  echo -e "  ${YELLOW}QUESTION:${NC} Why does the install fail? How would you fix it?"
  echo
}

# =============================================================================
# ROOM 3 — "Trust No Tag"
# =============================================================================
setup_room3() {
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  info "ROOM 3: Trust No Tag 🏷️"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  ensure_namespace "${NS_ROOM3}"
  create_kosli_secret "${NS_ROOM3}"

  # --- PUZZLE 1: Tag mutation ---
  # Build first image and deploy pod-v2-1 before tag is moved
  info "Building gameday-myapp:v2 (original)..."
  docker build -t gameday-myapp:v2 -f- . <<'DOCKERFILE' >/dev/null 2>&1
FROM alpine:3.19
LABEL org.opencontainers.image.version="v2-original"
CMD ["sleep", "3600"]
DOCKERFILE

  info "Deploying app-v2-1 with current gameday-myapp:v2..."
  cat <<EOF | kubectl apply -n "${NS_ROOM3}" -f - >/dev/null 2>&1
apiVersion: v1
kind: Pod
metadata:
  name: app-v2-1
  labels:
    app: room3
    role: healthy
spec:
  terminationGracePeriodSeconds: 1
  containers:
  - name: app
    image: gameday-myapp:v2
    imagePullPolicy: Never
    command: ["sleep", "3600"]
EOF

  # Silently move the tag to a different image — the mutation participants must discover
  info "Rebuilding gameday-myapp:v2 with different content (tag mutation)..."
  docker build -t gameday-myapp:v2 -f- . <<'DOCKERFILE' >/dev/null 2>&1
FROM busybox:1.36
LABEL org.opencontainers.image.version="v2-patched"
CMD ["sleep", "3600"]
DOCKERFILE

  info "Deploying app-v2-2 with (silently mutated) gameday-myapp:v2..."
  cat <<EOF | kubectl apply -n "${NS_ROOM3}" -f - >/dev/null 2>&1
apiVersion: v1
kind: Pod
metadata:
  name: app-v2-2
  labels:
    app: room3
    role: healthy
spec:
  terminationGracePeriodSeconds: 1
  containers:
  - name: app
    image: gameday-myapp:v2
    imagePullPolicy: Never
    command: ["sleep", "3600"]
EOF

  # --- PUZZLE 2: CrashLoopBackOff pod included in snapshot ---
  # Pod phase stays Running even while crashing; imageID is populated after first start.
  # The reporter includes it — participants must explain why.
  info "Deploying crashing pod..."
  cat <<EOF | kubectl apply -n "${NS_ROOM3}" -f - >/dev/null 2>&1
apiVersion: v1
kind: Pod
metadata:
  name: app-crasher
  labels:
    app: room3
    role: crasher
spec:
  restartPolicy: Always
  terminationGracePeriodSeconds: 1
  containers:
  - name: app
    image: alpine:3.19
    command: ["sh", "-c", "exit 1"]
EOF

  info "Waiting for healthy pods to be ready..."
  kubectl wait --for=condition=Ready pod -l role=healthy \
    -n "${NS_ROOM3}" --timeout=60s >/dev/null 2>&1 || true

  info "Waiting for crasher to populate imageID (needs at least 1 restart)..."
  local deadline=$((SECONDS + 60))
  while [ $SECONDS -lt $deadline ]; do
    local restarts
    restarts=$(kubectl get pod app-crasher -n "${NS_ROOM3}" \
      -o jsonpath='{.status.containerStatuses[0].restartCount}' 2>/dev/null || echo "")
    if [ -n "${restarts}" ] && [ "${restarts}" -ge 1 ]; then
      break
    fi
    sleep 3
  done

  info "Installing reporter..."
  helm upgrade --install kosli-reporter-room3 "${CHART_PATH}" \
    --namespace "${NS_ROOM3}" \
    --set reporterConfig.kosliOrg="${KOSLI_ORG}" \
    --set reporterConfig.kosliEnvironmentName="${KOSLI_ENV}" \
    --set reporterConfig.namespaces="${NS_ROOM3}" \
    --set reporterConfig.dryRun=true \
    --set serviceAccount.permissionScope=namespace \
    >/dev/null 2>&1

  info "Running reporter job..."
  kubectl delete job room3-manual-run -n "${NS_ROOM3}" --ignore-not-found >/dev/null 2>&1
  kubectl create job room3-manual-run -n "${NS_ROOM3}" \
    --from=cronjob/kosli-reporter-room3-k8s-reporter >/dev/null 2>&1
  kubectl wait --for=condition=complete job/room3-manual-run \
    -n "${NS_ROOM3}" --timeout=120s >/dev/null 2>&1 || true

  success "Room 3 ready!"
  echo
  echo -e "  ${YELLOW}SCENARIO:${NC} Three pods in ${NS_ROOM3}. The snapshot holds two surprises."
  echo
  echo -e "  ${YELLOW}PUZZLE 1 — The Tag That Lies:${NC}"
  echo -e "    kubectl get pods -n ${NS_ROOM3} -o custom-columns='NAME:.metadata.name,IMAGE:.spec.containers[0].image'"
  echo -e "    kubectl get pod app-v2-1 -n ${NS_ROOM3} -o jsonpath='{.status.containerStatuses[0].imageID}'"
  echo -e "    kubectl get pod app-v2-2 -n ${NS_ROOM3} -o jsonpath='{.status.containerStatuses[0].imageID}'"
  echo -e "  ${YELLOW}QUESTION:${NC} Both pods claim to run gameday-myapp:v2. Are they really the same image?"
  echo
  echo -e "  ${YELLOW}PUZZLE 2 — The CrashLoop Witness:${NC}"
  echo -e "    kubectl get pods -n ${NS_ROOM3}"
  echo -e "    kubectl logs -n ${NS_ROOM3} job/room3-manual-run --tail=80"
  echo -e "  ${YELLOW}QUESTION:${NC} app-crasher is in CrashLoopBackOff. Does Kosli include it in the snapshot? Why?"
  echo
}

# =============================================================================
# ROOM 4 — "The Double Report"
# =============================================================================
setup_room4() {
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  info "ROOM 4: The Double Report 👯"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  ensure_namespace "${NS_ROOM4}"
  create_kosli_secret "${NS_ROOM4}"
  deploy_dummy_pods "${NS_ROOM4}" 2

  info "Installing reporter releases..."
  helm upgrade --install kosli-reporter "${CHART_PATH}" \
    --namespace "${NS_ROOM4}" \
    --set reporterConfig.kosliOrg="${KOSLI_ORG}" \
    --set reporterConfig.kosliEnvironmentName="${KOSLI_ENV}" \
    --set reporterConfig.namespaces="${NS_ROOM4}" \
    --set reporterConfig.dryRun=true \
    --set serviceAccount.permissionScope=namespace \
    >/dev/null 2>&1

  helm upgrade --install kosli-reporter-duplicate "${CHART_PATH}" \
    --namespace "${NS_ROOM4}" \
    --set reporterConfig.kosliOrg="${KOSLI_ORG}" \
    --set reporterConfig.kosliEnvironmentName="${KOSLI_ENV}" \
    --set reporterConfig.namespaces="${NS_ROOM4}" \
    --set reporterConfig.dryRun=true \
    --set serviceAccount.permissionScope=namespace \
    --set fullnameOverride=kosli-reporter-dup \
    >/dev/null 2>&1

  success "Room 4 ready!"
  echo
  echo -e "  ${YELLOW}SCENARIO:${NC} Duplicate snapshots are being sent. What's causing this?"
  echo -e "  ${YELLOW}INVESTIGATE:${NC}"
  echo -e "    helm list -n ${NS_ROOM4}"
  echo -e "    kubectl get cronjob -n ${NS_ROOM4}"
  echo -e "  ${YELLOW}QUESTION:${NC} What concurrency policy does the CronJob use, and why isn't it preventing duplicates?"
  echo
}

# =============================================================================
# Teardown — Clean up everything
# =============================================================================
teardown() {
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  info "TEARDOWN: Cleaning up all escape room resources"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

  switch_context

  # Uninstall all helm releases
  for release in kosli-reporter-room1 kosli-reporter-room2 kosli-reporter-room3 \
                 kosli-reporter kosli-reporter-duplicate; do
    for ns in "${NS_DEFAULT}" "${NS_ROOM3}" "${NS_ROOM4}"; do
      helm uninstall "${release}" -n "${ns}" >/dev/null 2>&1 && \
        info "Uninstalled ${release} from ${ns}" || true
    done
  done

  # Delete game day namespaces
  for ns in "${NS_APP_TEAM}" "${NS_TEAM_ALPHA}" "${NS_TEAM_BETA}" \
            "${NS_ROOM3}" "${NS_ROOM4}"; do
    kubectl delete namespace "${ns}" --ignore-not-found >/dev/null 2>&1 && \
      info "Deleted namespace: ${ns}" || true
  done

  # Clean up pods/secrets in default namespace
  kubectl delete pod -l room=escape-room -n default --ignore-not-found >/dev/null 2>&1 || true
  kubectl delete secret kosli-api-token -n default --ignore-not-found >/dev/null 2>&1 || true

  # Remove local docker images
  docker rmi gameday-myapp:v2 >/dev/null 2>&1 && \
    info "Removed local docker image gameday-myapp:v2" || true

  echo
  success "All escape room resources cleaned up!"
}

# =============================================================================
# Setup All
# =============================================================================
setup_all() {
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  info "SETTING UP ALL 4 ESCAPE ROOMS"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo

  switch_context
  verify_prereqs

  echo
  setup_room1
  setup_room2
  setup_room3
  setup_room4

  echo
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  success "ALL 4 ROOMS ARE READY! 🎮"
  info "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
  echo
  echo -e "  Room 1: ${CYAN}The Invisible Pods${NC}"
  echo -e "  Room 2: ${CYAN}The Regex Trap${NC}"
  echo -e "  Room 3: ${CYAN}Trust No Tag${NC}"
  echo -e "  Room 4: ${CYAN}The Double Report${NC}"
  echo
  echo -e "  To tear down after the game: ${YELLOW}$0 teardown${NC}"
  echo
}

# =============================================================================
# Entrypoint
# =============================================================================
case "${1:-help}" in
  setup-all)
    setup_all
    ;;
  setup-room)
    switch_context
    case "${2:-}" in
      1) setup_room1 ;;
      2) setup_room2 ;;
      3) setup_room3 ;;
      4) setup_room4 ;;
      *) err "Usage: $0 setup-room [1|2|3|4]"; exit 1 ;;
    esac
    ;;
  teardown)
    teardown
    ;;
  verify)
    switch_context
    verify_prereqs
    ;;
  *)
    echo "Usage: $0 {setup-all|setup-room N|teardown|verify}"
    echo
    echo "  setup-all      Set up all 4 escape rooms"
    echo "  setup-room N   Set up room N (1-4)"
    echo "  teardown       Clean up all escape room resources"
    echo "  verify         Check prerequisites"
    exit 0
    ;;
esac