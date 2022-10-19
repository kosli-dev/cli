---
title: 'Report Environment'
weight: 10
---

# Report Environment

## Create an environment in Kosli

The first thing we need to do is creating an **environment** in [Kosli](https://app.kosli.com). 
Kosli **Environments** is where you'll be reporting the state of your actual environments, like *staging* or *production*. 
You can either create an environment with [Kosli CLI](/installation)) or via the web UI. We will use the CLI in this guide.

You need a name for your **environment** - it doesn't have to be the same name you use for the actual environment, but it certainly helps to identify it in the future. In this guide we'll use **github-k8s-test** as the name of the **environment**.
You also need to provide the description of the environment. You'll find this helpful as the number of your environments increases.

```shell {.command}
kosli env declare --name github-k8s-test --environment-type K8S --description "k8s and github actions demo" --api-token <your-api-token> --owner <your-github-username>
```

## Report an environment

Time to implement an actual reporting of what's running in your k8s cluster - which means we need to reach out to the cluster and check which docker images were used to run the containers that are currently up in a given namespace. 
You can run the [k8s environment report command](https://docs.kosli.com/client_reference/kosli_environment_report_k8s/) manually on any machine that can access your k8s cluster, but it is much better to automate the reporting from the start, and we'll use GitHub Actions for that.

### GitHub workflow

There are a few things you'll need to adjust in the workflow below, so it can work for you:

* `K8S_CLUSTER_NAME` and `K8S_GCP_ZONE` should refer to your cluster setup and `NAMESPACE` should refer to a namespace you will to deploy your application to
* `KOSLI_OWNER` is your Kosli username (which will be the same as the GitHub account you used to log into Kosli)

With these ready you can try to run the following workflow:

```
name: Report environment

# You can choose to run the reporting on schedule but for 
# the purpose of setting it up you may also want to be able
# to run it manually, so we added 'workflow_dispatch'
on:
  schedule:
    - cron: '*/5 * * * *'
  workflow_dispatch:

env:
  K8S_CLUSTER_NAME: kosli-dev
  K8S_GCP_ZONE: europe-west1
  NAMESPACE: github-k8s-demo
  KOSLI_OWNER: demo
  KOSLI_ENVIRONMENT: github-k8s-test
  KOSLI_CLI_VERSION: "2.0.0"

jobs:
  report-env:
    runs-on: ubuntu-20.04

    steps:

    - name: setup-kosli-cli
      uses: kosli-dev/setup-cli-action@v1
      with:
        version:
          ${{ env.KOSLI_CLI_VERSION }}

    - name: auth
      uses: google-github-actions/auth@v0.4.0
      with:
        credentials_json: ${{ secrets.GCP_K8S_CREDENTIALS }}

    - name: connect-to-k8s
      uses: google-github-actions/get-gke-credentials@main
      with:
        cluster_name: ${{ env.K8S_CLUSTER_NAME }}
        location: ${{ env.K8S_GCP_ZONE }}

    - name: report to Kosli
      env:
        KOSLI_API_TOKEN: ${{ secrets.KOSLI_API_TOKEN }}
      run:
        ./kosli environment report k8s --kubeconfig ${{ env.KUBECONFIG }} -n ${{ env.NAMESPACE }} ${{ env.KOSLI_ENVIRONMENT }}
```

Once the workflow runs successfully, and there is already something running in your cluster, you will see the information about it in **github-k8s-test** environment in Kosli (You'll find it under **Environments** section).  
If there is nothing running in your cluster we'll build and deploy an artifact in the next step.

If you had something running in the given namespace, here is what you should see in your **github-k8s-test environment** in Kosli if the pipeline succeeds (triggered either by cron or - if you don't want to wait - manually). The name of the artifact will likely be a different one:

![Incompliant environment, artifact with no provenance](/images/env-no-provenance.png)

Reporting an **environment** is an easy way to get the answer to a question like: "What is running in production?". 
So, naturally, the next thing you may want to figure out is: "Is it verified?".

For now, whatever is running there, will be incompliant since we don't know anything else about the artifact just yet, but reporting your artifacts to Kosli will take us one step further.

