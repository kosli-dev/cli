---
title: 'Getting started with K8S and GitHub actions'
weight: 1
---
# Getting started with Kubernetes and GitHub actions

In this guide you will learn how to track changes in a Kubernetes environment using GitHub Actions.

## Prerequisites

To follow the "Getting Started" guide you'll need to set up a few things:

1. [Kosli account](https://app.kosli.com/signup)
2. Fork [github-k8s-demo repository](https://github.com/kosli-dev/github-k8s-demo)
3. Your own Kubernetes cluster where you'll deploy the demo application
4. hub.docker.com account


### GitHub secrets

Create following Actions Secrets in your forked repository on GitHub:
* **KOSLI_API_TOKEN** - you can find the Kosli API Token under your profile at https://app.kosli.com/ (click your avatar in the right top corner of the window and select 'Profile')
* **GCP_K8S_CREDENTIALS** - service account credentials (.json file), with Kubernetes access permissions
* **DOCKERHUB_TOKEN** - your DockerHub Access Token (you can create one at hub.docker.com, under *Account Settings* > *Security*)



#### GitHub workflow

This is the GitHub action workflow for reporting changes to your Kubernetes namespace to Kosli.
```
name: Report environment

on:
  # Report every 5 minutes
  schedule:
    - cron: '*/5 * * * *'
  # You might also want to run it manually, so we added workflow_dispatch
  workflow_dispatch:

env:
  K8S_CLUSTER_NAME: INSERT_YOUR_VALUE_HERE
  K8S_GCP_ZONE: INSERT_YOUR_VALUE_HERE
  NAMESPACE: github-k8s-demo
  MERKELY_OWNER: INSERT_YOUR_VALUE_HERE
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

    - name: Authenticate to Google
      uses: google-github-actions/auth@v0.4.0
      with:
        credentials_json: ${{ secrets.GCP_K8S_CREDENTIALS }}

    - name: Get Google Kubernetes Engine credentials into kube config
      uses: google-github-actions/get-gke-credentials@main
      with:
        cluster_name: ${{ env.K8S_CLUSTER_NAME }}
        location: ${{ env.K8S_GCP_ZONE }}

    - name: Report what is running in your Kubernetes namespace to Kosli
      env:
        MERKELY_API_TOKEN: ${{ secrets.MERKELY_API_TOKEN }}
      run:
        ./kosli environment report k8s \
          --kubeconfig ${{ env.KUBECONFIG }} \
          --namespace ${{ env.NAMESPACE }} \
          ${{ env.MERKELY_ENVIRONMENT }}
```

xxxxxxxxx We got to here in the review xxxxxxxxxx

There is a few things you'll need to edit in the workflow below:

* `K8S_CLUSTER_NAME` and `K8S_GCP_ZONE` should refer to your cluster setup
* `NAMESPACE` should refer to the namespace you will deploy your application to
* `MERKELY_OWNER` is your Kosli username (which will be the same as the GitHub account you used to log into Kosli)

With these ready you can try to run the following workflow:


Once the workflow runs successfully, and there is already something running in your cluster, you will see the information about it in **github-k8s-test** environment in Kosli (You'll find it under **Environments** section).  
If there is nothing running in your cluster we'll build and deploy an artifact in the next step.

If you had something running in given namespace, here is what you should see in your **github-k8s-test environment** in Kosli if the pipeline succeedes (triggered either by cron or - if you don't want to wait - manually). The name of the artifact will likely be a different one:

![Incompliant environment, artifact with no provenance](/images/env-no-provenance.png)

Reporting an **environment** is an easy way to get the answer to a question like: "What is running in production?". 
So, naturally, the next thing you may want to figure out is: "Is it verified?".

For now, whatever is running there, will be incompliant since we don't know anything else about the artifact just yet, but reporting your artifacts to Kosli will take us one step further.





## Reporting K8S namespace changes

The first thing we will configure is an [**environment**](/concepts/environments/) in Kosli. In this example we will be tracking a K8S namespace called **github-k8s-test**.

You need to give your **environment** a name - it doesn't have to be the same name you use for the actual environment, but it certainly helps to identify it in the future. In this guide we'll use **github-k8s-test** as the name of the **environment**.
You also need to provide the description of the environment. You'll find this helpful as the number of your environments increases.


### Report an environment

Time to implement an actual reporting of what's running in your k8s cluster - which means we need to reach out to the cluster and check which docker images were used to run the containers that are currently up in given namespace. 

#### CLI

You report the environment using [Kosli CLI tool](https://github.com/kosli-dev/cli/releases).  
You need to download a correct package depending on the architecture of the machine you use to run the CLI. 

You can run the [command](https://docs.merkely.com/client_reference/merkely_environment_report_k8s/) manually on any machine that can access your k8s cluster, but it is much better to automate the reporting from the start, and we'll use GitHub Actions for that.


Move this section to environment reporting section: 
In our example we use Google Cloud to host k8s cluster and we rely on `google-github-actions/get-gke-credentials` action to authenticate to GKE cluster via a `kubeconfig` file. If you're hosting your k8s cluster somewhere else you need to use a different action.
