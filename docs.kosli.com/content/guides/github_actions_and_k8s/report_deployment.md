---
title: 'Report Deployment'
weight: 30
---

# Report Deployment

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

Visit [Kosli Commands](/client_reference) section to learn more about available Kosli CLI commands.


