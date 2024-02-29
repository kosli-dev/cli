---
title: "Attesting Snyk scans"
bookCollapseSection: false
weight: 507
---

# Attesting Snyk Scans

Snyk scans analyze your source code, docker images and IaC source for security issues and vulnerabilities. Reporting these results to Kosli is beneficial for: 
- Tracking whether the snyk scan happened on a given artifact or trail or not.
- Keeping a record of the findings.  

In this tutorial, we will see how you can run and attest different types of Snyk scans to Kosli. We will run the scans on the [Kosli CLI git repo](https://github.com/kosli-dev/cli).

{{<hint info>}}
While snyk attestations can be bound to a trail or an artifact in a trail, this tutorial
demonstrates it only on trails for simplicity.
{{</hint>}}

## Getting ready

To follow the steps in this tutorial, you need to:
* [Setup Snyk on your machine](https://docs.snyk.io/snyk-cli/getting-started-with-the-snyk-cli#install-the-snyk-cli-and-authenticate-your-machine).
* [Install Helm](https://helm.sh/docs/intro/install/) if you want to try Snyk IaC attestations, otherwise skip.
* [Install Docker](https://docs.docker.com/engine/install/) if you want to try Snyk container attestations, otherwise skip.
* [Create a Kosli account](https://app.kosli.com/) (Skip if you already have one).
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to your personal org name and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  $ export KOSLI_ORG=<your-personal-kosli-org-name>
  $ export KOSLI_API_TOKEN=<your-api-token>
  ```
* Clone the Kosli CLI git repo
  ```shell {.command}
  $ git clone https://github.com/kosli-dev/cli.git 
  $ cd cli
  ```

## Creating a Flow and Trail

We will start by creating a flow in Kosli to contain Trails and Artifacts for this demo.

```shell {.command}
$ kosli create flow snyk-demo --use-empty-template
```

{{<hint info>}}
`--use-empty-template` indicates that this flow does not have a predefined set of required attestations.
{{</hint>}}

Then, we can start a trail to bind our snyk attestations to.

```shell {.command}
$ kosli begin trail test-1 --flow snyk-demo
```

Now we can start running Snyk scans and attest them to this trail.

{{<hint info>}}
After each attestation in the sections below, you can navigate to:
**https://app.kosli.com/\<your-personal-org-name\>/flows/snyk-demo/trails/test-1** to view the status of the trail in Kosli.
{{</hint>}}

## Snyk Open source scan

[Snyk Open Source](https://docs.snyk.io/scan-using-snyk/snyk-open-source) allows you to find and fix vulnerabilities in the open-source libraries used by your applications. 

You can run a snyk opens source scan and report it to Kosli as follows:
```shell {.command}
$ snyk test --sarif-file-output=os.json

$ kosli attest snyk --flow snyk-demo --trail test-1 --name open-source-scan --scan-results os.json --commit HEAD
```

{{<hint info>}}
`--commit` allows you to relate the attestation to a specific git commit.
{{</hint>}}


## Snyk Code scan

[Snyk Code](https://docs.snyk.io/scan-using-snyk/snyk-code) lets you scan your source code for security issues. 

You can run a snyk code scan and report it to Kosli as follows:
```shell {.command}
$ snyk code test --sarif-file-output=code.json

$ kosli attest snyk --flow snyk-demo --trail test-1 --name code-scan --scan-results code.json --commit HEAD
```

## Snyk Container scan

[Snyk Container](https://docs.snyk.io/scan-using-snyk/snyk-container) lets you scan your container images for security issues. 

You can run a snyk container scan and report it to Kosli as follows:
```shell {.command}
# pull the cli docker image before scanning it
$ docker pull ghcr.io/kosli-dev/cli:v2.8.3
$ snyk container test ghcr.io/kosli-dev/cli:v2.8.3  --file=Dockerfile --sarif-file-output=container.json

$ kosli attest snyk --flow snyk-demo --trail test-1 --name container-scan --scan-results container.json --commit HEAD
```

## Snyk IaC scan

[Snyk IaC](https://docs.snyk.io/scan-using-snyk/snyk-iac) lets you scan various types of IaC configuration files (e.g. Terraform, Kubernetes, Helm) for security issues. 

We can run a snyk IaC scan on the K8S reporter Helm chart and report it to Kosli as follows:
```shell {.command}
$ helm template ./charts/k8s-reporter --output-dir helm \
   --set kosliApiToken.secretName=secret \
   --set reporterConfig.kosliEnvironmentName=foo \
   --set reporterConfig.kosliOrg=bar

$ snyk iac test helm  --sarif-file-output=helm.json

$ kosli attest snyk --flow snyk-demo --trail test-1 --name helm-scan --scan-results helm.json --commit HEAD
```

You can refer to the [Snyk docs](https://docs.snyk.io/snyk-cli/scan-and-maintain-projects-using-the-cli/snyk-cli-for-iac/test-your-iac-files) for more information on supported IaC configuration formats and how you can run snyk scans on them.

For more details about the `kosli attest snyk` command, please refer to [its CLI reference](/client_reference/kosli_attest_snyk/).