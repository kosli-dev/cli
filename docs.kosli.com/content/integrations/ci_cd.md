---
title: CI/CD
bookCollapseSection: false
weight: 310
aliases:
    - /ci-defaults  # To keep short URL in docs and help in the CLI
---
# Use Kosli in CI Systems

This section provides how-to guides showing you how to use Kosli to report changes from
different CI systems.

## Defaulted Kosli command flags from CI variables

The following flags are defaulted as follows in the CI list below:

{{< tabs "ci-defaults" "col-no-wrap" >}}

{{< tab "Bitbucket Cloud" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | https://bitbucket&#46;org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER} |
| --commit-url | https://bitbucket&#46;org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT} |
| --commit | ${BITBUCKET_COMMIT} |
| --git-commit | ${BITBUCKET_COMMIT} |
| --repository | ${BITBUCKET_REPO_SLUG} |
| --bitbucket-workspace |  ${BITBUCKET_WORKSPACE} |
{{< /tab >}}

{{< tab "Github" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID} |
| --commit-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA} |
| --commit | ${GITHUB_SHA} |
| --git-commit | ${GITHUB_SHA} |
| --repository | ${GITHUB_REPOSITORY} |
| --github-org | ${GITHUB_REPOSITORY_OWNER} |
{{< /tab >}}

{{< tab "Gitlab" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${CI_JOB_URL} |
| --commit-url | ${CI_PROJECT_URL}/-/commit/${CI_COMMIT_SHA} |
| --commit | ${CI_COMMIT_SHA} |
| --git-commit | ${CI_COMMIT_SHA} |
| --repository | ${CI_PROJECT_NAME} |
| --gitlab-org | ${CI_PROJECT_NAMESPACE} |
{{< /tab >}}

{{< tab "Teamcity" >}}
| Flag | Default |
| :--- | :--- |
| --git-commit | ${BUILD_VCS_NUMBER} |
{{< /tab >}}

{{< /tabs >}}

## Github Actions

To use Kosli in [Github Actions](https://docs.github.com/en/actions) workflows, you can use the kosli [CLI setup action](https://github.com/marketplace/actions/setup-kosli-cli) to install the CLI on your Github Actions Runner.
Then, you can use all the [CLI commands](/client_reference) in your workflows.

Here is an example Github Actions workflow snippet using the `kosli create flow` command:

```yaml
jobs:
  example:
    runs-on: ubuntu-latest
    env:
      KOSLI_API_TOKEN: ${{ secrets.MY_KOSLI_API_TOKEN }}
      KOSLI_OWNER: my-org
    steps:
      - name: setup kosli
        uses: kosli-dev/setup-cli-action@v1
      - name: create flow
        run: kosli create flow my-flow --template pull-request,artifact,test
```

{{< hint info >}}
Note that all CLI command flags can be set as environment variables by adding the the `KOSLI_` prefix and capitalizing them. 
In the example above, both `--api-token` and `--owner` flags were set from environment variables.
{{< /hint >}}

### Defaulted CLI flags in Github Actions

The following commands flags are defaulted when the Kosli CLI is run inside a Github Actions workflow:

| Flag | Default |
| :--- | :--- |
| --build-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID} |
| --commit-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA} |
| --commit | ${GITHUB_SHA} |
| --git-commit | ${GITHUB_SHA} |
| --repository | ${GITHUB_REPOSITORY} |
| --github-org | ${GITHUB_REPOSITORY_OWNER} |