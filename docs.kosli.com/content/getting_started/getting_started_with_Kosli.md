---
title: Getting started with Kosli
bookCollapseSection: false
weight: 10
# aliases:
#     - /getting_started
---
# Getting started with Kosli

A typical use of Kosli follows a scenario:

1. Record the state of your runtime environments
1. Connect the data from you pipelines 
1. Search in Kosli to get the whole picture

We'll be using Kosli CLI to run all the commands. You can find installation instruction [here](/getting_started/installation).

In the examples below we report k8s type of [environment](/introducing_kosli/environments) but the general approach would work for any supported type.


### Example repository
If you want to see our workflow examples you can find them at [github-k8s-demo repository](https://github.com/kosli-dev/github-k8s-demo)

## Record Environment

### Create an environment in Kosli

The first thing we need to do is to create an **[environment](/introducing_kosli/environments)** in [Kosli](https://app.kosli.com). 
Kosli **environment** is where you'll be reporting the state of your actual runtime environments (k8s cluster, docker host, AWS lambda, ...), like *staging* or *production*. 
You can either create an environment with [Kosli CLI](/introducing_kosli/cli/) or via the web UI. 

You need a name for your **environment** - it doesn't have to be the same name you use for the actual environment, but it certainly helps to identify it in the future. In this guide we use **github-k8s-test** as the name of the **environment**.
You also need to provide the description of the environment. You'll find this helpful as the number of your environments increases.

```shell {.command}
kosli env declare \
    --name github-k8s-test \
    --environment-type K8S \
    --description "k8s and github actions demo" \
    --api-token <your-api-token> \
    --owner <your-github-username>
```

### Report an environment

To record what's running in your k8s cluster you can run the [k8s environment report command](/client_reference/kosli_environment_report_k8s/) - manually on any machine that can access your k8s cluster or you can automate the reporting with cron job or using CI tool, e.g. you can use [GitHub Actions](https://github.com/kosli-dev/github-k8s-demo/blob/main/.github/workflows/report.yml) for that.

```shell {.command}
kosli environment report k8s github-k8s-test \
    --kubeconfig <path to kubeconfig> \
    --namespace <namespace to report> \
    --api-token <your-api-token> \
    --owner <your-github-username>
```

Once the workflow runs successfully, and there is already something running in your environment, you will see the information about it in **github-k8s-test** environment in Kosli (You'll find it under **Environments** section).  

## Connect the pipeline

Every time you build an **artifact** you can store (and easily access) all the information you have about it in Kosli.

Artifacts in Kosli are reported to Kosli **[pipelines](/introducing_kosli/pipelines)**. You can find the **Pipelines** section just below **Environments** in Kosli.

### Create a pipeline

To report an **artifact** you need to create a Kosli **pipeline** first. Every time your workflow builds a new version of Docker image it will be reported to the same Kosli **pipeline**.

```shell {.command}
kosli pipeline declare --description "Kosli server" \
    --pipeline github-k8s-demo \
    --template "artifact" \
    --api-token <your-api-token> \
    --owner <your-github-username>
```

### Report an artifact

To report an artifact use [kosli pipeline artifact report creation](/client_reference/kosli_pipeline_artifact_report_creation/) command:

```shell {.command}
kosli pipeline artifact report creation <your artifact name> 
    --artifact-type <docker, file or dir> \
    --api-token <your-api-token> \
    --owner <your-github-username> \
    --pipeline github-k8s-demo
```

Typically you'd run that command in your CI system as part of your pipeline, and if you use GitHub Actions or Bitbucket Pipelines a number of required flag would be [automatically taken care of](/ci-defaults). 

If you want to run the command manually you'd need to provide required flags on your own, so your artifact reporting command would have to be extended by:

```shell {.command}
--build-url <will accept any string> --commit-url <will accept any string> --git-commit <your commit sha>
```

Once the workflow runs successfully, you should see it reported in Kosli **github-k8s-demo pipeline**

### Report Deployment

The missing piece is making sure you know how your artifact ended up in the environment, and that's why, when your workflow deploys to an environment, it should report the deployment to that environment to Kosli.  

```shell {.command}
kosli pipeline deployment report <your artifact name> \
    --sha256 <your artifact fingerprint> \
    --build-url <deployment build url> \
    --api-token <your-api-token> \
    --owner <your-github-username> \
    --environment <kosli environment to report to>
```

The [main.yml](https://github.com/kosli-dev/github-k8s-demo/blob/main/.github/workflows/main.yml) workflow in the [github-k8s-demo](https://github.com/kosli-dev/github-k8s-demo) repository is a complete workflow for reporting an artifact and deployment to Kosli.


Visit [Kosli Commands](/client_reference) section to learn more about available Kosli CLI commands.


## GitHub workflow notes

In order to reuse our [demo repository](https://github.com/kosli-dev/github-k8s-demo) you need to have following secrets configured in your CI:

* **KOSLI_API_TOKEN** - you can find the Kosli API Token under your profile at https://app.kosli.com/ (click on your avatar in the right top corner of the window and select 'Profile')
* **GCP_K8S_CREDENTIALS** - service account credentials (.json file), with k8s access permissions
* **DOCKERHUB_TOKEN** - your DockerHub Access Token (you can create one at hub.docker.com, under *Account Settings* > *Security*)

There are also a few things you'll need to adjust in the workflows, so it can work for you:

* `K8S_CLUSTER_NAME` and `K8S_GCP_ZONE` should refer to your cluster setup and `NAMESPACE` should refer to a namespace you will to deploy your application to
* `KOSLI_OWNER` is your Kosli username (which will be the same as the GitHub account you used to log into Kosli)
