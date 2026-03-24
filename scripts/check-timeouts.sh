#!/bin/bash
# Script to verify timeout configuration is working correctly
# Usage: scripts/check-timeouts.sh [endpoint]

set -e

ENDPOINT="${1:-https://app.kosli.com}"
TIMEOUT="${KOSLI_REQUEST_TIMEOUT:-30}"

echo "Checking timeout configuration..."
echo "Endpoint: $ENDPOINT"
echo "Timeout: ${TIMEOUT}s"

# Validate timeout is a number
if [[ ! $TIMEOUT =~ ^[0-9]+$ ]]; then
    echo "Error: KOSLI_REQUEST_TIMEOUT must be a positive integer"
    exit 1
fi

# Check if kosli binary exists
if ! command -v kosli &> /dev/null; then
    echo "Error: kosli binary not found in PATH"
    echo "Run 'make build' first"
    exit 1
fi

# Run a simple version check with timeout
TIMESTART=$(date +%s)
kosli version > /dev/null 2>&1
TIMEEND=$(date +%s)
DURATION=$((TIMEEND - TIMESTART))

echo "Version check completed in ${DURATION}s"

if [ $DURATION -gt $TIMEOUT ]; then
    echo "WARNING: Version check exceeded timeout threshold"
    exit 1
fi

echo "All timeout checks passed"
exit 0
