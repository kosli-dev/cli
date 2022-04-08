---
title: 'Report Artifact'
weight: 20
---

# Report Artifact

Every time you build an **artifact** - in our case a Docker image - you can store and easily access all the information you have about it in Merkely. We call it *reporting an **artifact***.

Artifacts in Merkely are reported to Merkely **Pipelines**. You can find the **Pipelines** section just below **Environments**.

## Create a pipeline

To report an **artifact** from your GitHub workflow you need to create a Merkely **pipeline** first. Every time your workflow builds a Docker image it will be reported to the same Merkely **pipeline**.  
Merkely **pipeline** has to exist before you can start reporting **artifacts** to it, and you can make the creation a part of the build workflow. (It's safe - rerunning **pipeline** creation command won't erase existing entries.)  
In this guide we're creating a Merkely **pipeline** called **test-pipeline** and that's the name you'll see in the code.

As it was in the case of reporting environment, we need to download Merkely CLI in the workflow, to be able to run the commands. 

## Report an artifact

Here is a complete workflow that takes care of CLI download, **pipeline** creation and docker image build and reports it to the Merkely **pipeline**. Remember to replace *MERKELY_OWNER* variable value with your Merkely username.

Below you'll find comments about Merkely specific parts of the workflow.


```
name: Build and Deploy

on:
  push:


env: 
  K8S_CLUSTER_NAME: merkely-dev
  K8S_GCP_ZONE: europe-west1
  NAMESPACE: github-k8s-demo
  IMAGE: ewelinawilkosz/github-k8s-demo
  MERKELY_OWNER: demo
  MERKELY_PIPELINE: github-k8s-demo
  MERKELY_CLI_VERSION: "1.5.0"

jobs:
  build-report:
    runs-on: ubuntu-20.04

    outputs:
      tag: ${{ steps.prep.outputs.tag }}
      tagged-image: ${{ steps.prep.outputs.tagged-image }}
      #image-digest: ${{ steps.digest-prep.outputs.image-digest }}

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

    - name: Login to GitHub Container Registry
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
        cache-from: type=registry,ref=${{ env.IMAGE }}:buildcache
        cache-to: type=registry,ref=${{ env.IMAGE }}:buildcache,mode=max

    # - name: Make the image digest available for following steps
    #   id: digest-prep
    #   run: |
    #     ARTIFACT_SHA=$( echo ${{ steps.docker_build.outputs.digest }} | sed 's/.*://')
    #     echo "DIGEST=$ARTIFACT_SHA" >> ${GITHUB_ENV}
    #     echo ::set-output name=image-digest::${ARTIFACT_SHA}

    # - name: Download Merkely cli client
    #   id: download-merkely-cli
    #   run: |
    #     wget https://github.com/merkely-development/cli/releases/download/v${{ env.MERKELY_CLI_VERSION }}/merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz
    #     tar -xf merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz merkely

    # - name: Declare pipeline in Merkely (staging)
    #   env:
    #     MERKELY_API_TOKEN: ${{ secrets.MERKELY_API_TOKEN }}
    #   run: 
    #     ./merkely pipeline declare 
    #       --description "Merkely server" 
    #       --pipeline ${{ env.MERKELY_PIPELINE }} 
    #       --template "artifact"


    # - name: Report Docker image in Merkely (production)
    #   env:
    #     MERKELY_API_TOKEN: ${{ secrets.MERKELY_API_TOKEN }}
    #   run: 
    #     ./merkely pipeline artifact report creation ${{ env.TAGGED_IMAGE }}
    #       --sha256 ${{ env.DIGEST }} 




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

