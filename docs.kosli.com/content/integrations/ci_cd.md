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

{{< hint info >}}
Note that **all** CLI command flags can be set as environment variables by adding the the `KOSLI_` prefix and capitalizing them. 
In the GitHub workflow example [further down](/integrations/ci_cd/#use-kosli-in-github-actions), both `--api-token` and `--org` flags were set from environment variables.
{{< /hint >}}

## Defaulted Kosli command flags from CI variables

The following flags are **defaulted** (which means you don't need to provide the flags, they'll be automatically set to values listed below) as follows in the CI systems below:

{{< tabs "ci-defaults" "col-no-wrap" >}}

{{< tab "Azure DevOps" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${SYSTEM_COLLECTIONURI}/${SYSTEM_TEAMPROJECT}/_build/results?buildId=${BUILD_BUILDID} |
| --commit-url | ${SYSTEM_COLLECTIONURI}/${SYSTEM_TEAMPROJECT}/_git/${BUILD_REPOSITORY_NAME}/commit/${BUILD_SOURCEVERSION} |
| --commit | ${BUILD_SOURCEVERSION} |
| --git-commit | ${BUILD_SOURCEVERSION} |
| --repository | ${BUILD_REPOSITORY_NAME} |
| --project | ${SYSTEM_TEAMPROJECT} |
| --azure-org-url | ${SYSTEM_COLLECTIONURI} |
{{< /tab >}}

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

{{< tab "CircleCI" >}}
| Flag | Default |
| :--- | :--- |
| --build-url | ${CIRCLE_BUILD_URL} |
| --commit-url | ${CIRCLE_REPOSITORY_URL}(converted to https url)/commit(s)/${CIRCLE_SHA1} |
| --git-commit | ${CIRCLE_SHA1} |
{{< /tab >}}

{{< tab "Teamcity" >}}
| Flag | Default |
| :--- | :--- |
| --git-commit | ${BUILD_VCS_NUMBER} |
{{< /tab >}}

{{< /tabs >}}


## Use Kosli in Github Actions

To use Kosli in [Github Actions](https://docs.github.com/en/actions) workflows, you can use the kosli [CLI setup action](https://github.com/marketplace/actions/setup-kosli-cli) to install the CLI on your Github Actions Runner.
Then, you can use all the [CLI commands](/client_reference) in your workflows.

### GitHub Secrets 

Keep in mind that secrets in Github actions are not automatically exported as environment variables. You need to add required secrets to your GITHUB environment explicitly. E.g. to make kosli_api_token secret available for all cli commands as an environment variable use following:

```yaml
env:
  KOSLI_API_TOKEN: ${{ secrets.kosli_api_token }}
```

### Example

Here is an example Github Actions workflow snippet using `kosli-dev/setup-cli-action` running `kosli create flow` command:

```yaml
jobs:
  example:
    runs-on: ubuntu-latest
    env:
      KOSLI_API_TOKEN: ${{ secrets.MY_KOSLI_API_TOKEN }}
      KOSLI_ORG: my-org
    steps:
      - name: setup kosli
        uses: kosli-dev/setup-cli-action@v2
      - name: create flow
        run: kosli create flow my-flow --template pull-request,artifact,test
```
