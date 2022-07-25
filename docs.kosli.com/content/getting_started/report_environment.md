---
title: 'Report Environment'
weight: 10
---

# Report Environment

## Create an environment in Kosli

The first thing we need to configure is an **environment** in [Kosli](https://app.merkely.com).  
Kosli **Environments** is where you'll be reporting the state of your actual environments, like *staging* or *production*. 

When you log in to Kosli the **Environments** page is the first thing you see. If you clicked around before reading this guide you'll find a link to **Environments** on the left side of the window in Kosli.

Click the "Add environment" button to create a new Kosli **environment**. On the next page you'll have to select the type - for the purpose of this guide we'll use 'Kubernetes cluster'.

Next, you need to give your **environment** a name - it doesn't have to be the same name you use for the actual environment, but it certainly helps to identify it in the future. In this guide we'll use **github-k8s-test** as the name of the **environment**.
You also need to provide the description of the environment. You'll find this helpful as the number of your environments increases.

Click "Save Environment" and you're ready to move on to the next step.

## Report an environment

Time to implement an actual reporting of what's running in your k8s cluster - which means we need to reach out to the cluster and check which docker images were used to run the containers that are currently up in given namespace. 

### CLI

You report the environment using [Kosli CLI tool](https://github.com/kosli-dev/cli/releases).  
You need to download a correct package depending on the architecture of the machine you use to run the CLI. 

You can run the [command](https://docs.kosli.com/client_reference/merkely_environment_report_k8s/) manually on any machine that can access your k8s cluster, but it is much better to automate the reporting from the start, and we'll use GitHub Actions for that.

### GitHub workflow

There is a few things you'll need to adjust in the workflow below, so it can work for you:

* `K8S_CLUSTER_NAME` and `K8S_GCP_ZONE` should refer to your cluster setup and `NAMESPACE` should refer to a namespace you will to deploy your application to
* `MERKELY_OWNER` is your Kosli username (wchich will be the same as the GitHub account you used to log into Kosli)

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
  K8S_CLUSTER_NAME: merkely-dev
  K8S_GCP_ZONE: europe-west1
  NAMESPACE: github-k8s-demo
  MERKELY_OWNER: demo
  MERKELY_ENVIRONMENT: github-k8s-test
  MERKELY_CLI_VERSION: "1.5.0"

jobs:
  report-env:
    runs-on: ubuntu-20.04

    steps:

    - name: Download Kosli cli client
      id: download-merkely-cli
      run: |
        wget https://github.com/kosli-dev/cli/releases/download/v${{ env.MERKELY_CLI_VERSION }}/merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz
        tar -xf merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz kosli

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
        MERKELY_API_TOKEN: ${{ secrets.MERKELY_API_TOKEN }}
      run:
        ./merkely environment report k8s --kubeconfig ${{ env.KUBECONFIG }} -n ${{ env.NAMESPACE }} ${{ env.MERKELY_ENVIRONMENT }}
```

Once the workflow runs successfully, and there is already something running in your cluster, you will see the information about it in **github-k8s-test** environment in Kosli (You'll find it under **Environments** section).  
If there is nothing running in your cluster we'll build and deploy an artifact in the next step.

If you had something running in given namespace, here is what you should see in your **github-k8s-test environment** in Kosli if the pipeline succeedes (triggered either by cron or - if you don't want to wait - manually). The name of the artifact will likely be a different one:

![Incompliant environment, artifact with no provenance](/images/env-no-provenance.png)

Reporting an **environment** is an easy way to get the answer to a question like: "What is running in production?". 
So, naturally, the next thing you may want to figure out is: "Is it verified?".

For now, whatever is running there, will be incompliant since we don't know anything else about the artifact just yet, but reporting your artifacts to Kosli will take us one step further.

