---
title: Sonar
bookCollapseSection: false
weight: 340
---
# Record Sonar scan results in Kosli

The results of SonarCloud and SonarQube scans can be tracked in [Kosli trails](/getting_started/trails/). <br>
This integration involves setting up a Sonar webhook in Kosli and a corresponding webhook in SonarCloud or SonarQube. When you run a scan of your SonarCloud/SonarQube project, the webhook is triggered and the results of the scan are sent to Kosli.<br>
Some parameters must be passed to the Sonar scanner when it is run (e.g. the name of the Flow corresponding to the project, and the name of the trail the results should be attested to); these are sent with the scan results, and allow Kosli to determine the compliance status of the results and attest them to the correct trail/artifact.

## Setting up in Kosli

To set up the integration, navigate to the Sonar integration page for your org in the [Kosli app](https://app.kosli.com/).

After switching on the integration, you will be provided with a webhook and a secret.

## Setting up Sonar Webhooks

You're now just a few steps away from connecting SonarCloud/SonarQube to Kosli.

Both SonarCloud and SonarQube provide two types of webhooks: global (which are triggered when any project in your organization is scanned) and project-specific (which are triggered by a scan for that project only). Kosli supports both types of webhooks.

In [SonarCloud](https://sonarcloud.io/) or [SonarQube](https://sonarqube.org):

### To create a global webhook:

- In SonarCloud: Go to your Organization, then Administration > Webhooks
- In SonarQube: Go to Administration > Configuration > Webhooks
- Create a new Webhook
- Add the Kosli webhook URL and secret provided
- Click Create

![SonarCloud Global Webhook page](/images/sonarcloud_integration_global.png)
![SonarQube Global Webhook page](/images/sonarqube_integration_global.png)

### To create a project-specific webhook:

- Go to the project you want to create a webhook for
- Click on Administration (SonarCloud) or Project Settings (SonarQube) and go to Webhooks in the dropdown menu
- Create a new Webhook
- Add the Kosli webhook URL and secret provided
- Click Create

![SonarCloud Project Webhook page](/images/sonarcloud_integration_project.png)
![SonarQube Project Webhook page](/images/sonarqube_integration_project.png)

## Setting up the SonarScanner

In order for Kosli to know where the scan results should be attested, certain parameters can be passed to the SonarScanner. Note that parameters cannot be passed with SonarCloud's Automatic Analysis - in this case, Kosli determines the relevant Flow and Trail as described below.

These parameters can be passed to the scanner in three ways:
- As part of the sonar-project.properties file used in CI analysis
- As arguments to the scanner in your CI pipeline's YML file
```shell
    - name: SonarCloud Scan
        uses: sonarsource/sonarcloud-github-action@master
        with:
          args: >
            -Dsonar.analysis.kosli_flow=<YourFlowName>
            -Dsonar.analysis.kosli_trail=<YourTrailName>
```
- As arguments to the CLI scanner
```shell
$ sonar scanner \
  -Dsonar.analysis.kosli_flow=<YourFlowName> \
  -Dsonar.analysis.kosli_trail=<YourTrailName> 
```


### Scanner parameters:
- `sonar.analysis.kosli_flow=<YourFlowName>`
    - The name of the Flow relevant to your project. If a Flow does not already exist with the given name, it is created. If no Flow name is provided, the project key of your project in SonarCloud/SonarQube is used as the name (with any invalid symbols replaced by '-').
- `sonar.analysis.kosli_trail=<YourTrailName>`
    - The name of the Trail to attest the scan results. If a Trail does not already exist with the given name it is created. If no Trail name is provided, the revision ID of the Sonar project (typically defaulted to the Git SHA) is used as the name.
- `sonar.analysis.kosli_attestation=<YourAttestationName>`
    - The name you want to give to the attestation. If not provided, a default name "sonar" is used. If using dot-notation (of the form `<YourTargetArtifact.YourAttestationName>`), either the artifact fingerprint or git commit is also required (see below).
- `sonar.analysis.kosli_git_commit=<GitCommitSHA>`
    - The git commit for the attestation. If not provided the revision ID of the Sonar project is used (provided it has the correct format for a git SHA).
- `sonar.analysis.kosli_artifact_fingerprint=<YourArtifactFingerprint>`
    - The fingerprint of the artifact you want the attestation to be attached to. Requires that the artifact has already been reported to Kosli.
- `sonar.analysis.kosli_flow_description=<DescriptionOfYourKosliFlow>`
    - The description for the Kosli Flow being created by this webhook. This will not be used if attesting to an already-existing Flow (i.e. will not change any existing descriptions).
- `sonar.analysis.kosli_trail_description=<DescriptionOfYourKosliTrail>`
    - The description for the Kosli Trail being created by this webhook. This will not be used if attesting to an already-existing Trail (i.e. will not change any existing descriptions).

## Testing the integration

To test the webhook once configured, simply scan a project in SonarCloud or SonarQube. If successful, the results of the scan will be attested to the relevant Flow and Trail (and artifact, if applicable) as a sonar attestation. <br>
If the webhook fails, check that you have passed the parameters to the scanner correctly, and that the trail name, attestation name and artifact fingerprint are valid.

## Live Example in CI system
View an example of a sonar attestation via webhook in Github.

In [this YAML file](https://app.kosli.com/api/v2/livedocs/cyber-dojo/yaml?ci=github&command=-Dsonar.analysis.kosli_flow), which created [this Kosli event](https://app.kosli.com/api/v2/livedocs/cyber-dojo/event?ci=github&command=-Dsonar.analysis.kosli_flow). 


## Alternatives:
If you'd rather not use webhooks, or they don't quite fit your use-case, we also have a [CLI command](/client_reference/kosli_attest_sonar/) for attesting Sonar scan results to Kosli.