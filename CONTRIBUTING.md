# Contributions to the Kosli CLI

## Prerequisites

You need to have the followin environment variable:

`export KOSLI_API_TOKEN_PROD=YourKeyHere`

As well as access to the KOSLI AWS accounts.

If you have not gotten that, and you are a KOSLI employee, please [read this](https://github.com/kosli-dev/knowledge-base/blob/master/aws_vault.md).

## Tools setup

You can compile and run the project via [DevBox from Jetlify](https://www.jetify.com/docs/devbox/installing_devbox/)

After the installation, run `devbox shell` and all relevant tools will be avaliable to you.

## Running tests

To run all tests, including the kubernetes tests, which take a few minutes:

```bash
make test_integration_full
```

To run tests and ignore tests that take longer to run:

```bash
make test_integration
```

To run a single test suite:

```bash
make test_integration_single TARGET=<suiteName>
```

Some tests will be skipped if the following environment variables are not set:

```bash
KOSLI_GITHUB_TOKEN
KOSLI_GITLAB_TOKEN
KOSLI_BITBUCKET_ACCESS_TOKEN
KOSLI_AZURE_TOKEN
KOSLI_SONAR_API_TOKEN
```

Additionally authentication is necessary to run some tests. See <https://github.com/kosli-dev/knowledge-base>.

## Releases

The version number is not generated automatically and must be decided manually.
We are using semantic versioning (ie: 2.3.2).

```bash
make release tag=v<version_number>
```
