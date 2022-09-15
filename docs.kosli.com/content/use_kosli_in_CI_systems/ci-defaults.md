---
title: "Defaulted Kosli command flags from CI variables"
weight: 2
aliases:
    - /ci-defaults
---

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

{{< tab "Teamcity" >}}
| Flag | Default |
| :--- | :--- |
| --git-commit | ${BUILD_VCS_NUMBER} |
{{< /tab >}}

{{< /tabs >}}