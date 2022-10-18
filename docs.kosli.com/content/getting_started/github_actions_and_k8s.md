---
title: Report with GitHub Actions and k8s
bookCollapseSection: false
weight: 30
---

# Report with GitHub Actions and k8s

In this guide, you will learn how to start using Kosli for one of our most common use cases: **k8s** as the environment where you deploy your application, **GitHub** as the repository, and **GitHub Actions** as the CI.

To get started, follow the sections below!

## Prerequisites

To follow the "Getting Started" guide you'll need to set up a few things:

1. Kosli account
2. GitHub repository where you'll store your code (you can fork our demo repository) 
3. Your own k8s cluster where you'll deploy your application
4. hub.docker.com account

### GitHub
If you want to use our workflow examples, there are a few things you need to configure in your GitHub repository - you can fork the [github-k8s-demo repository](https://github.com/kosli-dev/github-k8s-demo), create your own from scratch, or use an already existing project. 

Workflows in [github-k8s-demo repository](https://github.com/kosli-dev/github-k8s-demo) are the complete version of the workflows we're developing in this guide.

In our example we use Google Cloud to host k8s cluster and we rely on `google-github-actions/get-gke-credentials` action to authenticate to GKE cluster via a `kubeconfig` file. If you're hosting your k8s cluster elsewhere you need to use a different action.

#### Secrets

Create the following Actions Secrets in your repository on GitHub:
* **KOSLI_API_TOKEN** - you can find the Kosli API Token under your profile at https://app.kosli.com/ (click on your avatar in the right top corner of the window and select 'Profile')
* **GCP_K8S_CREDENTIALS** - service account credentials (.json file), with k8s access permissions
* **DOCKERHUB_TOKEN** - your DockerHub Access Token (you can create one at hub.docker.com, under *Account Settings* > *Security*)



Once these are in place you're ready to go!

## Report Environment

### Create an environment in Kosli

The first thing we need to do is creating an **environment** in [Kosli](https://app.kosli.com). 
Kosli **Environments** is where you'll be reporting the state of your actual environments, like *staging* or *production*. 
You can either create an environment with [Kosli CLI](../../installation/_index.md)) or via the web UI. We will use the CLI in this guide.

You need a name for your **environment** - it doesn't have to be the same name you use for the actual environment, but it certainly helps to identify it in the future. In this guide we'll use **github-k8s-test** as the name of the **environment**.
You also need to provide the description of the environment. You'll find this helpful as the number of your environments increases.

```shell {.command}
kosli env declare --name github-k8s-test --environment-type K8S --description "k8s and github actions demo" --api-token <your-api-token> --owner <your-github-username>
```

### Report an environment

Time to implement an actual reporting of what's running in your k8s cluster - which means we need to reach out to the cluster and check which docker images were used to run the containers that are currently up in a given namespace. 
You can run the [k8s environment report command](https://docs.kosli.com/client_reference/kosli_environment_report_k8s/) manually on any machine that can access your k8s cluster, but it is much better to automate the reporting from the start, and we'll use GitHub Actions for that.

#### GitHub workflow

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

## Report Artifact

Every time you build an **artifact** - in our case a Docker image - you can store (and easily access) all the information you have about it in Kosli. We call it *reporting an **artifact***.

Artifacts in Kosli are reported to Kosli **Pipelines**. You can find the **Pipelines** section just below **Environments**.

### Create a pipeline

To report an **artifact** from your GitHub workflow you need to create a Kosli **pipeline** first. Every time your workflow builds a new version of Docker image it will be reported to the same Kosli **pipeline**.
Kosli **pipeline** has to exist before you can start reporting **artifacts** to it, and you can make the creation of a **pipeline** a part of the build workflow. (It's safe - rerunning **pipeline** creation command won't erase existing entries.)
In this guide we're creating a Kosli **pipeline** called **github-k8s-demo** and that's the name you'll see in the code.

As it was in the case of reporting environment, we need to download Kosli CLI in the workflow, to be able to run the commands.

### Report an artifact

Here is a complete workflow that takes care of CLI download, **pipeline** creation and docker image build and reports it to the Kosli **pipeline**.

Remember:
* `K8S_CLUSTER_NAME`, `K8S_GCP_ZONE` and `NAMESPACE` should be the same you used in **Report Environment** step
* `IMAGE` should contain your dockerhub username (instead of our colleague's Ewelina username). You also need to use the correct username in *Login to hub.docker.com* step
* `KOSLI_OWNER` should be the same your Kosli username.


In the workflow you'll find comments about specific parts of it.

```
name: Build and Deploy

on:
  push:


env:
  # gke k8s cluster variables
  K8S_CLUSTER_NAME: kosli-dev
  K8S_GCP_ZONE: europe-west1
  NAMESPACE: github-k8s-demo
  # name of the docker image to build, replace with the name
  # that will contain your dockerhub id
  IMAGE: ewelinawilkosz/github-k8s-demo
  # kosli variables - will be picked up by commands
  KOSLI_OWNER: demo
  KOSLI_PIPELINE: github-k8s-demo
  KOSLI_ENVIRONMENT: github-k8s-test
  KOSLI_CLI_VERSION: "2.0.0"

jobs:
  build-report:
    runs-on: ubuntu-20.04

    # outputs to be passed on to 'deploy' job below
    outputs:
      tag: ${{ steps.prep.outputs.tag }}
      tagged-image: ${{ steps.prep.outputs.tagged-image }}
      image-digest: ${{ steps.digest-prep.outputs.image-digest }}

    steps:
    # checkout code
    - uses: actions/checkout@v2
      with:
        fetch-depth: 1

    - name: Prepare
      id: prep
      run: |
        TAG=$(echo $GITHUB_SHA | head -c7)
        TAGGED_IMAGE=${{ env.IMAGE }}:${TAG}
        echo "TAG=${TAG}" >> ${GITHUB_ENV}
        echo "TAGGED_IMAGE=${TAGGED_IMAGE}" >> ${GITHUB_ENV}
        echo ::set-output name=tag::${TAG}
        echo ::set-output name=tagged-image::${TAGGED_IMAGE}

    # This is the a separate action that sets up buildx (buildkit) runner
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v1

    # use your own username and configured token to log into dockerhub
    - name: Login to hub.docker.com
      uses: docker/login-action@v1
      with:
        username: ewelinawilkosz
        password: ${{ secrets.DOCKERHUB_TOKEN }}

    - name: Build and push Docker image
      id: docker_build
      uses: docker/build-push-action@v2
      with:
        push: true
        tags: ${{ env.TAGGED_IMAGE }}
        no-cache: true

    # the digest will be passed to kosli commands using 'sha256' flags
    - name: Make the image digest available for following steps
      id: digest-prep
      run: |
        ARTIFACT_SHA=$( echo ${{ steps.docker_build.outputs.digest }} | sed 's/.*://')
        echo "DIGEST=$ARTIFACT_SHA" >> ${GITHUB_ENV}
        echo ::set-output name=image-digest::${ARTIFACT_SHA}

    - name: setup-kosli-cli
      uses: kosli-dev/setup-cli-action@v1
      with:
        version:
          ${{ env.KOSLI_CLI_VERSION }}

    - name: Declare pipeline in Kosli
      env:
        KOSLI_API_TOKEN: ${{ secrets.KOSLI_API_TOKEN }}
      run:
        ./kosli pipeline declare
          --description "Kosli server"
          --pipeline ${{ env.KOSLI_PIPELINE }}
          --template "artifact"

    - name: Report Docker image in Kosli
      env:
        KOSLI_API_TOKEN: ${{ secrets.KOSLI_API_TOKEN }}
      run:
        ./kosli pipeline artifact report creation ${{ env.TAGGED_IMAGE }}
          --sha256 ${{ env.DIGEST }}

  # deploy review environment
  deploy:
    needs: [build-report]
    runs-on: ubuntu-20.04

    steps:
    - uses: actions/checkout@v2
      with:
        fetch-depth: 1

    - id: auth
      uses: google-github-actions/auth@v0.4.0
      with:
        credentials_json: ${{ secrets.GCP_K8S_CREDENTIALS }}

    - id: connect-to-k8s
      uses: google-github-actions/get-gke-credentials@main
      with:
        cluster_name: ${{ env.K8S_CLUSTER_NAME }}
        location: ${{ env.K8S_GCP_ZONE }}

    - uses: azure/setup-kubectl@v1
      id: install-kubectl

    # The KUBECONFIG env var is automatically exported and picked up by kubectl.
    - name: Ensure review env namespace
      run: |
        kubectl get namespace ${{ env.NAMESPACE }} || kubectl create namespace ${{ env.NAMESPACE }}

    - name: Deploy
      run: |
        sed -i 's/TAG/${{ needs.build-report.outputs.tag }}/g' k8s/deployment.yaml
        kubectl apply -f k8s/deployment.yaml -n ${{ env.NAMESPACE }}
```

Once the workflow runs successfully, you should see it reported in Kosli **github-k8s-demo pipeline**:

![Compliant artifact with no deployments](/images/artifact-list.png)

With more details once you click on it:

![Compliant artifact with no deployments](/images/artifact-no-deployment.png)

You will also notice a change in the state of your **github-k8s-test** environment (if the environment reporting workflow ran successfully): it is still incompliant, but now the artifact running there has provenance (you can see the name of Kosli **pipeline: github-k8s-demo** that the artifact was reported to, in a grey, pill shaped field) so we can check how it was built:

![Incompliant environment, artifact with provenance](/images/env-provenance.png)


Now that your artifact reporting works the only thing missing is reporting the deployment.

## Report Deployment

In previous sections we covered reporting environment - so you know what's running in your cluster, and reporting artifact - so you know what you're building (and - in the future, if you add more controls, you'll know if it's verified).

The missing piece is figuring out how your artifact ended up in the environment, and that's why, when our workflow deploys to an environment, we report the deployment to that environment to Kosli.  

We'll extend the workflow from previous section with two steps, to add the reporting at the end the `deploy` job:

``` 
    - name: setup-kosli-cli
      uses: kosli-dev/setup-cli-action@v1
      with:
        version:
          ${{ env.KOSLI_CLI_VERSION }} 

    - name: Report deployment
      env:
        KOSLI_API_TOKEN: ${{ secrets.KOSLI_API_TOKEN }}
      run: 
        ./kosli pipeline deployment report ${{ needs.build-report.outputs.tagged-image }}
          --sha256 ${{ needs.build-report.outputs.image-digest }} 
```

The [main.yml](https://github.com/kosli-dev/github-k8s-demo/blob/main/.github/workflows/main.yml) workflow in the [github-k8s-demo](https://github.com/kosli-dev/github-k8s-demo) repository is a complete workflow for reporting an artifact and deployment to Kosli.

Once the pipeline runs successfully you should see new entry in your **github-k8s-demo pipeline** in Kosli, this time with **deployment** linked in the last column:

![Compliant artifact with no deployments](/images/artifact-list-2.png)

Before we check the environment we need to - again - wait for the environment reporting workflow to kick in (or run it manually) and when it succeeds we can check the status of the environment.

This time it should be compliant - which means we know where the artifact is coming from and how it ended up in the environment:

![Compliant environment](/images/env-compliant.png)

In our example, *deployment* is part of the same workflow as *build*. In real life you may want to deploy in a separate pipeline, especially if you're deploying to your production environment. Once you learn how to use Kosli with this example it should be easier to add required steps to your existing workflows, wherever you need them. 

Visit [Kosli Commands](../../client_reference) section to learn more about available Kosli CLI commands.


