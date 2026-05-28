#!/bin/bash
# Run a passing trail
#../../kosli evaluate input --policy four-eyes.rego -i trails/v2.11.42-pass.json --decision
# Run a failing trail
../../kosli evaluate input --policy four-eyes.rego -i trails/v2.11.44-fail.json --decision