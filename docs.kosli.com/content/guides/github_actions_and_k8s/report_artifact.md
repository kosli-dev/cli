---
title: 'Report Artifact'
weight: 20
---

# Report Artifact

Every time you build an **artifact** - in our case a Docker image - you can store (and easily access) all the information you have about it in Kosli. We call it *reporting an **artifact***.

Artifacts in Kosli are reported to Kosli **Pipelines**. You can find the **Pipelines** section just below **Environments**.

## Create a pipeline

To report an **artifact** from your GitHub workflow you need to create a Kosli **pipeline** first. Every time your workflow builds a new version of Docker image it will be reported to the same Kosli **pipeline**.
Kosli **pipeline** has to exist before you can start reporting **artifacts** to it, and you can make the creation of a **pipeline** a part of the build workflow. (It's safe - rerunning **pipeline** creation command won't erase existing entries.)
In this guide we're creating a Kosli **pipeline** called **github-k8s-demo** and that's the name you'll see in the code.

As it was in the case of reporting environment, we need to download Kosli CLI in the workflow, to be able to run the commands.

## Report an artifact

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
    - uses: actions/checkout@v3
      with:
        fetch-depth: 1

    - name: Prepare
      id: prep
      run: |
        TAG=$(echo $GITHUB_SHA | head -c7)
        TAGGED_IMAGE=${{ env.IMAGE }}:${TAG}
        echo "TAG=${TAG}" >> ${GITHUB_ENV}
        echo "TAGGED_IMAGE=${TAGGED_IMAGE}" >> ${GITHUB_ENV}
        echo "tag=$TAG" >> $GITHUB_OUTPUT
        echo "tagged-image=$TAGGED_IMAGE" >> $GITHUB_OUTPUT

    # This is the a separate action that sets up buildx (buildkit) runner
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2

    # use your own username and configured token to log into dockerhub
    - name: Login to hub.docker.com
      uses: docker/login-action@v1
      with:
        username: ewelinawilkosz
        password: ${{ secrets.DOCKERHUB_TOKEN }}

    - name: Build and push Docker image
      id: docker_build
      uses: docker/build-push-action@v3
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
        echo "image-digest=$ARTIFACT_SHA" >> $GITHUB_OUTPUT


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
    - uses: actions/checkout@v3
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