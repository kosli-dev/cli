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
KOSLI_BITBUCKET_USERNAME
KOSLI_BITBUCKET_PASSWORD
KOSLI_AZURE_TOKEN
```

Additionally authentication is necessary to run some tests. See https://github.com/kosli-dev/knowledge-base.

## Releases

The version number is not generated automatically and must be decided manually.
We are using semantic versioning (ie: 2.3.2).
```
make release tag=v<version_number>
```