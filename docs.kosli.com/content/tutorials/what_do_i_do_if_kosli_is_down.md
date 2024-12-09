---
title: "What do I do if Kosli is down?"
bookCollapseSection: false
weight: 507
---

# What do I do if Kosli is down?

Customers use Kosli to attest evidence of their business and software processes.
If Kosli is down, these attestations will fail.
This will break CI workflow pipelines, blocking artifacts from being deployed.
In this situation there is a built-in mechanism to instantly turn Kosli off and keep the pipeline flowing.
When Kosli is back up, you can instantly turn Kosli back on.

## Turning Kosli CLI calls on and off instantly

If the `KOSLI_DRY_RUN` environment variable is set to `true` then all Kosli CLI commands will:
* Not communicate with Kosli at all
* Print the payload they would have sent
* Exit with a zero status code

We recommend creating an Org-level KOSLI_DRY_RUN variable in your CI system and, in all CI workflows,
ensuring there is an environment variable set from it. 

For example, in a [Github Action workflow](https://github.com/cyber-dojo/differ/blob/main/.github/workflows/main.yml):

```yaml
name: Main
...
env:
  KOSLI_DRY_RUN: ${{ vars.KOSLI_DRY_RUN }}           # true iff Kosli is down
```


## Turning Kosli API calls on and off instantly

If you are using the Kosli API in your workflows (e.g. using `curl`), we recommend using the same Org-level `KOSLI_DRY_RUN` 
environment variable and guarding the `curl` call with a simple if statement. For example:

```shell
#!/usr/bin/env bash

kosli_curl()
{
  local URL="${1}"
  local JSON_PAYLOAD="${2}"

  if [ "${KOSLI_DRY_RUN:-}" == "true" ]; then
    echo KOSLI_DRY_RUN is set to true. This is the payload that would have been sent
    echo "${JSON_PAYLOAD}" | jq .
  else
    curl ... --data="${JSON_PAYLOAD}" "${URL}"
  fi
}
```




