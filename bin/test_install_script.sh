#!/bin/bash
set -e

# --- Configuration ---
TOKEN=""

# Parse arguments for the test script itself, primarily to pass --token
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --token) 
		    if [[ -n "$2" && "$2" != --* ]]; then
		      TOKEN="$2"
			  shift
			else
			  echo "Error: --token requires a value"
			  exit 1
			fi
			;;
        *) echo "Unknown parameter: $1"; exit 1 ;;
    esac
    shift
done

# Helper to construct command
run_install() {
    local cmd="./install-cli.sh"
    if [ -n "$TOKEN" ]; then
        cmd="$cmd --token $TOKEN"
    fi
    # Add other arguments passed to function
    cmd="$cmd $@"
    echo "Running: $cmd"
    $cmd
}

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO] $1${NC}"
}

log_error() {
    echo -e "${RED}[ERROR] $1${NC}"
}


# Test 1: Install specific version
SPECIFIC_VERSION="v2.11.40"
log_info "Test 1: Installing specific version ${SPECIFIC_VERSION}..."

# We pass --version as requested by the user
run_install --version "${SPECIFIC_VERSION}" --debug

if ! command -v kosli &> /dev/null; then
    log_error "Kosli CLI not found after installation"
    exit 1
fi

INSTALLED_VERSION=$(kosli version | grep -o "v[0-9]\+\.[0-9]\+\.[0-9]\+")
log_info "Installed version: ${INSTALLED_VERSION}"

if [[ "${INSTALLED_VERSION}" == "${SPECIFIC_VERSION}" ]]; then
    log_info "✅ Specific version installed successfully"
else
    log_info "Expected ${SPECIFIC_VERSION}, got ${INSTALLED_VERSION}"
    log_error "❌ Version mismatch"
    exit 1
fi

# Test 2: Upgrade to latest version
log_info "Test 2: Upgrading to latest version..."
run_install --debug

LATEST_INSTALLED_VERSION=$(kosli version | grep -o "v[0-9]\+\.[0-9]\+\.[0-9]\+")
log_info "Installed version after update: ${LATEST_INSTALLED_VERSION}"

# Simple check to ensure version changed (assuming latest > specific)
if [[ "${LATEST_INSTALLED_VERSION}" != "${SPECIFIC_VERSION}" ]]; then
    log_info "✅ Version updated successfully (from ${SPECIFIC_VERSION} to ${LATEST_INSTALLED_VERSION})"
else
    log_error "❌ Version did not update"
    exit 1
fi

log_info "🎉 All installation tests passed!"
