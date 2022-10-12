---
title: Github Actions
weight: 1
---

## Github Actions

To use Kosli in [Github Actions](https://docs.github.com/en/actions) workflows, you can use the kosli [CLI setup action](https://github.com/marketplace/actions/setup-kosli-cli) to install the CLI on your Github Actions Runner.
Then, you can use all the [CLI commands](/client_references) in your workflows.

Here is an example Github Actions workflow snippet using the Kosli declare pipeline command:

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
      - name: declare pipeline
        run: kosli pipeline declare --pipeline my-pipeline -t pull-request,artifact,test
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