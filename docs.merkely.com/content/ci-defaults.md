---
title: "Defaulted command flags from CI"
---

## Defaulted command flags from CI

The following flags are defaulted as follows in the CI list below:

{{< tabs "uniqueid" >}}

{{< tab "Bitbucket Cloud" >}}
```bash
--build-url : https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/addon/pipelines/home#!/results/${BITBUCKET_BUILD_NUMBER}

--commit-url : https://bitbucket.org/${BITBUCKET_WORKSPACE}/${BITBUCKET_REPO_SLUG}/commits/${BITBUCKET_COMMIT}

--git-commit : ${BITBUCKET_COMMIT}
```
{{< /tab >}}

{{< tab "Github" >}}
```bash
--build-url : ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}

--commit-url : ${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/commit/${GITHUB_SHA}

--git-commit : ${GITHUB_SHA}
```
{{< /tab >}}

{{< tab "Teamcity" >}}
```bash
--git-commit : ${BUILD_VCS_NUMBER}
```
{{< /tab >}}

{{< /tabs >}}