---
title: 'Report Artifact'
weight: 20
---

# Report Artifact

Every time you build an **artifact** - in our case a docker image - you can keep and easily access all the information you have about it in Merkely. We call it *reporting an **artifact***.

Artifacts in Merkely are reported to a Merkely **Pipelines**. You can find the **Pipelines** section just below **Environments**.

## Create a pipeline

In order to report an **aritfact** from your GitHub workflow you need to create a Merkely **pipeline** first. Every time our workflow builds a docker image we will be reporting it to the same Merkely **pipeline**.  
Merkely **pipeline** has to exists before you start reporting **artifacts** to it, and you can make the creation a part of the build workflow. (It's safe, since rerunning **pipeline** creation command won't erase existing entries.)  
In this guide we're creating a Merkely **pipeline** called **test-pipeline** and that's the name you'll see in the code.

As it was in the case of reporting environment, we need to download Merkely CLI in the workflow, to be able to run the commands. 

## Report an artifact

Here is a complete workflow that takes care of CLI download, **pipeline** creation and docker image build we can report it to the Merkely **pipeline**. Remember to replace *MERKELY_OWNER* variable value with your Merkely username.

Below you'll find comments about Merkely specific parts of the workflow.


```
name: Build and Report

on:
  push:


env: 
  MERKELY_OWNER: compliancedb
  MERKELY_PIPELINE: test-pipeline
  MERKELY_CLI_VERSION: "1.5.0"

jobs:
  build-test-report:
    runs-on: ubuntu-20.04

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
        IMAGE=ghcr.io/merkely-development/merkely
        TAGGED_IMAGE=${IMAGE}:${TAG}
        echo "TAG=${TAG}" >> ${GITHUB_ENV}
        echo "IMAGE=${IMAGE}" >> ${GITHUB_ENV}
        echo "IMAGE_URI=${IMAGE}" >> ${GITHUB_ENV}
        echo "TAGGED_IMAGE=${TAGGED_IMAGE}" >> ${GITHUB_ENV}
        echo ::set-output name=tag::${TAG}
        echo ::set-output name=tagged-image::${TAGGED_IMAGE}
  
    # This is the a separate action that sets up buildx (buildkit) runner
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v1

    - name: Login to GitHub Container Registry
      uses: docker/login-action@v1
      with:
        registry: ghcr.io
        username: ${{ github.repository_owner }}
        password: ${{ secrets.GITHUB_TOKEN }}

    - name: Build and push Docker image
      id: docker_build
      uses: docker/build-push-action@v2
      with:
        push: true
        tags: ${{ env.TAGGED_IMAGE }}
        cache-from: type=registry,ref=${{ env.IMAGE }}:buildcache
        cache-to: type=registry,ref=${{ env.IMAGE }}:buildcache,mode=max
        build-args: |
          COMMIT_SHA=${{ github.sha }}

    - name: Tag image as latest and push it
      if: ${{ github.ref == 'refs/heads/master' || github.ref == 'refs/heads/main'  }}
      run: |
        docker pull ${{ env.TAGGED_IMAGE }}
        docker tag ${{ env.TAGGED_IMAGE }} ${{ env.IMAGE }}:latest
        docker push ${{ env.IMAGE }}:latest    

    - name: Download Merkely cli client
      id: download-merkely-cli
      run: |
        wget https://github.com/merkely-development/cli/releases/download/v${{ env.MERKELY_CLI_VERSION }}/merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz
        tar -xf merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz merkely

    - name: Declare pipeline in Merkely
      run: 
        ./merkely pipeline declare 
          --description "Merkely test docker image" 
          --pipeline ${{ env.MERKELY_PIPELINE }} 
          --template "artifact"

    - name: Report Docker image in Merkely (production)
      env:
        MERKELY_API_TOKEN: ${{ secrets.MERKELY_API_TOKEN }}
      run: 
        ./merkely pipeline artifact report creation ${{ env.TAGGED_IMAGE }}
          --artifact-type docker 

```

