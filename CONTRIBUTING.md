# Contributions to the Kosli CLI

## Running tests

To run all tests, including the kubernetes tests, which take a few minutes:
```
make test_integration_full
```

To run tests and ignore tests that take longer to run:
```
make test_integration
```

To run a single test suite:
```
make test_integration_single TARGET=<suiteName>
```

Some tests will be skipped if the following environment variables are not set:
```
KOSLI_GITHUB_TOKEN
KOSLI_GITLAB_TOKEN
KOSLI_BITBUCKET_ACCESS_TOKEN
KOSLI_AZURE_TOKEN
KOSLI_SONAR_API_TOKEN
```

Additionally authentication is necessary to run some tests. See https://github.com/kosli-dev/knowledge-base.

## Timeout Configuration

The CLI supports configurable timeouts for HTTP requests. See `internal/requests/timeout_config.go` for the implementation. Key environment variables:

- `KOSLI_REQUEST_TIMEOUT` - Default timeout in seconds (default: 30)
- `KOSLI_UPLOAD_TIMEOUT` - Upload timeout in seconds (default: 300)

## Code Style

We follow standard Go conventions. Please ensure:
- All exported functions have godoc comments
- Error messages start with lowercase (per Go convention)
- Tests use table-driven patterns where appropriate
- New dependencies are justified in the PR description

## Releases

The version number is not generated automatically and must be decided manually.
We are using semantic versioning (ie: 2.3.2).
```
make release tag=v<version_number>
```
