#!/bin/bash

# Attempt to run the install script
./install-cli.sh

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
