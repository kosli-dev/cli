---
title: "How to report Kubernetes Clusters"
bookCollapseSection: false
weight: 507
---

# How to report Kubernetes Clusters to Kosli

Kosli environments allow you to track changes in your physical/virtual runtime environments. Such changes must be reported from the runtime environment to Kosli.

This tutorial shows you how to set up reporting of running artifacts from a Kubernetes cluster to Kosli.


## Different ways for reporting

There are 3 different ways to report what's running in a Kubernetes cluster:

- Using Kosli CLI (suitable for testing only)
- Using a Kubernetes cronjob configured with a helm chart (recommended for production use).
- Using an externally scheduled cron process (e.g. a scheduled CI workflow)

We describe how to use the different options below and you can choose what suites your needs.

## Prerequisites

To follow this tutorial, you will need to:

- Have access to a Kubernetes cluster.
- [Create a Kosli account](https://app.kosli.com/sign-up) if you have not got one already.
- [Create a Kubernetes Kosli environment](/getting_started/environments/#create-an-environment) named `k8s-tutorial` 
- [Get a Kosli API token](/getting_started/service-accounts/)
- [Install Kosli CLI](/getting_started/install/) (only needed if you will report using CLI)
- [Install Helm](https://helm.sh/docs/intro/install/) (only needed if you will use the Kosli helm chart)

## Report snapshots using Kosli CLI

This option is **only suitable for testing purposes**. 

> All the commands below will use the default `kubecontext` in "$HOME/.kube/config". You can change it with `--kubeconfig` 

To report the **artifacts running in an entire cluster**, you can run the following command:

```shell {.command}
$ kosli snapshot k8s k8s-tutorial \
    --api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```

To report **artifacts running in one or more namespaces**, you can run the following command:

```shell {.command}
$ kosli snapshot k8s k8s-tutorial \
    --namespaces namespace1,namespace2 \
    --api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```

To report **artifacts running in the entire cluster except from some namespaces**, you can run the following command:

```shell {.command}
$ kosli snapshot k8s k8s-tutorial \
    --exclude-namespaces namespace1,namespace2 \
    --api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```

## Report snapshots using the Kosli K8S reporter helm chart

The recommended way to regularly report artifacts running in a cluster to Kosli is to use the [K8S reporter helm chart](/helm).

The chart creates a cronjob that will run the Kosli CLI inside a pod to report the artifacts running in the cluster.

1. Create a K8S secret to contain your Kosli API token.

```shell {.command}
$ kubectl create secret generic kosli-api-token --from-literal=apikey=<your-kosli-api-token>
```

> Make sure the secret value does not contain any trailing whitespace.

2. Prepare the settings for the helm chart

To customize how the helm chart creates the cronjob, you can create your own values file by copying and modifying the [default values file](https://github.com/kosli-dev/cli/blob/main/charts/k8s-reporter/values.yaml).

We will use this file (named `tutorial-values.yaml`):

```yaml {.command}
# -- the cron schedule at which the reporter is triggered to report to kosli  
cronSchedule: "*/5 * * * *"

kosliApiToken:
  # -- the name of the secret containing the kosli API token
  secretName: "kosli-api-token"
  # -- the name of the key in the secret data which contains the kosli API token
  secretKey: "apikey"

reporterConfig:
  # -- the name of the kosli org
  kosliOrg: "<your-kosli-org-name>"
  # -- the name of kosli environment that the k8s cluster/namespace correlates to
  kosliEnvironmentName: "k8s-tutorial"
  # -- the namespaces which represent the environment.
  # It is a comma separated list of namespace name regex patterns.
  # e.g. `^prod$,^dev-*` reports for the `prod` namespace and any namespace that starts with `dev-`
  # leave this unset if you want to report what is running in the entire cluster
  namespaces: ""
```

3. Install the Kosli helm chart

```shell {.command}
$ helm repo add kosli https://charts.kosli.com/
$ helm repo update
$ helm install kosli-reporter kosli/k8s-reporter -f tutorial-values.yaml
```

4. Confirm the cronjob is created in the cluster:

```shell {.command}
$ kubectl get cronjobs
```

Now, the cronjob will run every 5 minutes and report what is running in the entire cluster to Kosli.


## Report snapshots using externally scheduled cronjobs

If you do not wish to run the Kosli reporter inside the cluster, you can run it from outside the cluster. This requires opening access to the cluster from the place you will run the CLI regularly. 

One option to send reports regularly from outside the cluster is to use Github Actions scheduled workflows. Here is an example workflow definition:

> Note that the workflow below needs secrets to be added in Github actions.

```yaml {.command}
name: Regular Kubernetes reports to Kosli

on:
  workflow_dispatch: 
  schedule: 
    - cron: '0 * * * *' # every one hour

jobs:
  k8s-report:
    runs-on: ubuntu-latest
    permissions:
      id-token: write
      contents: write
    env:
      KOSLI_API_TOKEN: ${{ secrets.MY_KOSLI_API_TOKEN }}

    steps:
      - name: install kosli
        uses: kosli-dev/setup-cli-action@v2
      
      # connect to your cluster
      # if not using GKE, replace this step with one that connects to your cluster
      - name: Connect to GKE
        uses: 'Swibi/connect-to-gke'
        with:
          GCP_SA_KEY: ${{ secrets.GKE_SA_KEY }}
          GCP_PROJECT_ID: ${{ secrets.GKE_PROJECT }}
          GKE_CLUSTER: <your-cluster-name>
          GKE_ZONE: <your-cluster-zone>
      
      - name: Scan artifacts and send K8S report to Kosli
        run: 
          kosli snapshot k8s k8s-tutorial --org <your-kosli-org-name>
          
      # send slack notifications on failure to report
      - name: Slack Notification on Failure
        if: ${{ failure() }}
        uses: rtCamp/action-slack-notify@v2
        env:
          SLACK_CHANNEL: kosli-reports-failure
          SLACK_COLOR: ${{ job.status }}
          SLACK_TITLE: Reporting K8S artifacts to Kosli has failed
          SLACK_USERNAME: GithubActions
          SLACK_WEBHOOK: ${{ secrets.SLACK_CI_FAILURES_WEBHOOK }}
          SLACK_MESSAGE: "Reporting K8S artifacts to Kosli has failed. Please check the logs for more details."
```
