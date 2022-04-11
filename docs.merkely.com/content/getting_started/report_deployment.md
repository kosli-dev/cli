---
title: 'Report Deployment'
weight: 30
---

# Report Deployment

In previous sections we covered reporting environment - so you know what's running in your cluster, and reporting artifact - so you know what you're building (and - in the future, once you add more controls, you'll know if it's verified).

The missing piece is figuring out how your artifact ended up in production, and that's why we report deployment. We'll extend the workflow from previos step with two steps to add the reporting:

```
- name: Download Merkely cli client
      id: download-merkely-cli
      run: |
        wget https://github.com/merkely-development/cli/releases/download/v${{ env.MERKELY_CLI_VERSION }}/merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz
        tar -xf merkely_${{ env.MERKELY_CLI_VERSION }}_linux_amd64.tar.gz merkely 

    - name: Report deployment
      env:
        MERKELY_API_TOKEN: ${{ secrets.MERKELY_API_TOKEN }}
      run: 
        ./merkely pipeline deployment report ${{ needs.build-report.outputs.tagged-image }}
          --sha256 ${{ needs.build-report.outputs.image-digest }} 
```

In our example deployment is part of the same workflow as build so we'll extend the `deploy` job with the reporting part. In real life you may want to deploy in a seperate pipeline, especially if you're deploying to your production environment. Once you learn how to use Merkely using this example it should be easier to add required steps to your existing workflows, wherever you need them. 


Once the pipeline runs succesfully you should see new entry in your **github-k8s-demo pipeline** in Merkely, this time with a **Deployment** listed:

![Compliant artifact with no deployments](/images/artifact-list-2.png)

Before we check the environment we need to - again - wait for the environment reporting workflow to kick in (or run it manually) and when it succeeds we can check the status of the environment.

This time it should be compliant:

![Compliant environment](/images/env-compliant.png)
