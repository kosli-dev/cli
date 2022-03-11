---
title: "Defaulted command flags from CI"
---

## Defaulted command flags from CI

The following flags are defaulted as follows in the CI list below:

{{< tabs "uniqueid" >}}

{{< tab "Bitbucket Cloud" >}}
| Flag | Description |
| :--- | :--- |
| --build-url | https://bitbucket&#46;org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER} |
| --commit-url | https://bitbucket&#46;org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT} |
| --commit | ${BITBUCKET_COMMIT} |
| --git-commit | ${BITBUCKET_COMMIT} |
| --repository | ${BITBUCKET_REPO_SLUG} |
| --bitbucket-workspace |  ${BITBUCKET_WORKSPACE} |
{{< /tab >}}

{{< tab "Github" >}}
| Flag | Description |
| :--- | :--- |
| --build-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID} |
| --commit-url | ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA} |
| --commit | ${GITHUB_SHA} |
| --git-commit | ${GITHUB_SHA} |
| --repository | ${GITHUB_REPOSITORY} |
| --github-org | ${GITHUB_REPOSITORY_OWNER} |

{{< /tab >}}

{{< tab "Teamcity" >}}
| Flag | Description |
| :--- | :--- |
| --git-commit | ${BUILD_VCS_NUMBER} |
{{< /tab >}}

{{< /tabs >}}