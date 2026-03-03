#!/bin/bash

# Note: set -e is intentionally omitted here to allow manual exit code checking.

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

# Run the install script
run_install

# Capture the exit code of the install script
EXIT_CODE=$?

# Check if the exit code is 1 (expected failure due to existing brew installation)
if [ $EXIT_CODE -eq 1 ]; then
    echo "Success: install-cli.sh detected existing Homebrew installation and exited with 1."
    exit 0
else
    echo "Failure: install-cli.sh did not exit with 1. Actual exit code: $EXIT_CODE"
    exit 1
fi
