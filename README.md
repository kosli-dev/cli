[![codecov](https://codecov.io/gh/kosli-dev/cli/branch/main/graph/badge.svg?token=Z4Y53XIOKJ)](https://codecov.io/gh/kosli-dev/cli)
[![Static Badge](https://img.shields.io/badge/provenance-blue?style=plastic&link=https%3A%2F%2Fapp.kosli.com%2Fkosli-public%2Fflows%2Fcli-release%2Ftrails%2F)](https://app.kosli.com/kosli-public/flows/cli-release/trails/)

# Kosli Reporter

This CLI is used to report DevOps change events to Kosli and query them.

## Status

Kosli is still in beta

## Usage 

See the [docs](https://docs.kosli.com/client_reference/)

## Linting the code

`make lint`

## Building the code (Mac/Linux)

`make build`

Then to run Kosli commands:  
`./kosli [COMMAND]`

## Building the code (Windows)

Windows will not allow building using the makefile, so we need to run the commands directly in the terminal.

`set GOFLAGS=""`  
`go mod download`  
`go mod tidy`  
`go vet ./...`  
`go build -o kosli.exe -ldflags '-extldflags "-static"' ./cmd/kosli/`

Then to run Kosli commands:  
`./kosli.exe [COMMAND]` or `.\kosli.exe [COMMAND]`

## Building the documentation

`make hugo-local`

## Running the tests

To run only the tests in a single test suite, eg TestAttestJunitCommandTestSuite
`make test_integration_single TARGET=TestAttestJunitCommandTestSuite`
