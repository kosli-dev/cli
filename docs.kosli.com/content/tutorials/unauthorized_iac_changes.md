---
title: "Detecting unauthorized Terraform IaC changes"
bookCollapseSection: false
weight: 506
---

# Detecting unauthorized Terraform IaC changes

Authorized Terraform changes follow a predefined process that maintains a certain level of quality, security and safety for the underlying infrastructure. Unauthorized changes, however, can undermine the integrity and reliability of the infrastructure. Hence the importance of prompt detection of such changes.

Unauthorized Terraform changes happen in one of two ways:
1. Bypassing Terraform and making direct changes via cloud APIs, clients or UI consoles. This leads to drift, where the desired state does not match the actual state. This kind of unauthorized change can be detected and corrected with [Terraform drift detection](https://developer.hashicorp.com/terraform/tutorials/state/resource-drift).
2. Bypassing the predefined process for Terraform changes. For example, a developer running terraform directly from their machine without going through CI.

This tutorial shows how you can use Kosli to track and detect the second type of unauthorized changes.

## Prerequisites

To follow the steps in this tutorial, you need to:
* [Install Terraform on your machine](https://developer.hashicorp.com/terraform/install).
* (Optional)[Setup Snyk on your machine](https://docs.snyk.io/snyk-cli/getting-started-with-the-snyk-cli#install-the-snyk-cli-and-authenticate-your-machine).
* [Create a Kosli account](https://app.kosli.com/) (Skip if you already have one).
* [Install Kosli CLI](/getting_started/install/).
* [Get a Kosli API token](/getting_started/service-accounts/).
* Set the `KOSLI_ORG` environment variable to your personal org name and `KOSLI_API_TOKEN` to your token:
  ```shell {.command}
  export KOSLI_ORG=<your-personal-kosli-org-name>
  export KOSLI_API_TOKEN=<your-api-token>
  ```
* Clone the tutorial git repo
  ```shell {.command}
  git clone https://github.com/kosli-dev/iac-changes-tutorial.git 
  cd iac-changes-tutorial
  ```

## Creating a Kosli flow

We will start by creating a Kosli flow to represent the process for authorized Terraform changes.
For simplicity, we will not define any requirements for this process by using `--use-empty-template`

```shell {.command}
kosli create flow tf-tutorial --use-empty-template
```

## Making and tracking an authorized change

{{<hint info>}}
In production, an authorized change will normally go though CI.
In this tutorial, however, we run the commands that you would otherwise do in CI locally for simplicity.
{{</hint>}}

Let's create a trail to represent a single instance of making an authorized change. We will call it `authorized-1`.

```shell {.command}
kosli begin trail authorized-1 --flow=tf-tutorial
```
Next, we can scan our terraform config scripts for security issues. We capture the SARIF output from the scan and attest it to Kosli.

```shell {.command}
snyk iac test main.tf --sarif-file-output=sarif.json
kosli attest snyk --name=security --flow=tf-tutorial --trail=authorized-1 --scan-results=sarif.json
```

We are now ready to run terraform. We create a plan and save it to a file. Then attest the plan file to Kosli to build a historical audit log. 

```shell {.command}
terraform init
terraform plan -out=tf.plan
kosli attest generic --name=tf-plan --flow=tf-tutorial --trail=authorized-1 --attachments=tf.plan
```

Finally, we apply the terraform plan, and attest the produced terraform state file as an artifact.
This will calculate a SHA256 fingerprint for the state file based on its contents. The fingerprint will later be used to determine if a change is 
authorized or not.

{{<hint info>}}
In this tutorial, we use a simple setup where the terraform state file is stored locally.
In production cases, however, the state file would be stored in some cloud storage (e.g. AWS S3). 
In such cases, you would need to download the state file from the remote backend after it was updated by the authorized change.

Note that we set both `--build-url` and `--commit-url` to fake URLs. These are normally defaulted in CI.
{{</hint>}}

```shell {.command}
terraform apply -auto-approve tf.plan
kosli attest artifact terraform.tfstate --name=state-file --artifact-type=file --flow=tf-tutorial --trail=authorized-1 \
   --build-url=https://example.com --commit-url=https://example.com --commit=HEAD
```

## Monitoring the state file

Every time a change to the infrastructure happens via Terraform, the state file content would be changed. 
To detect when an **unauthorized** change happens, we can monitor the state file for changes and record those changes in
a Kosli environment.

Let's start by creating an environment of type `server`. 

```shell {.command}
kosli create env terraform-state --type=server
```

We can report the state file to the environment we created:

{{<hint info>}}
In this tutorial, we run the environment reporting manually. 
In production, you would configure the environment reporting to run periodically or on changes. 
See [reporting AWS environments](../report_aws_envs) if you are using S3 as a backend for your state files.
{{</hint>}}

```shell {.command}
kosli snapshot path terraform-state --name=tf-state --path=terraform.tfstate
```

You can get the latest snapshot of the environment by running:

```shell
kosli get snapshot terraform-state
COMMIT   ARTIFACT                                                                       FLOW         COMPLIANCE     RUNNING_SINCE  REPLICAS
d881b2f  Name: tf-state                                                                 tf-tutorial  NON-COMPLIANT  28 minutes ago   1
         Fingerprint: a57667a7b921b91d438631afa1a1fe35300b4da909a19d2b61196580f30f1d0c
```

Note that the `FLOW` column indicates that this artifact came from the `tf-tutorial` flow which means Kosli has provenance for 
where this change came from.

You can also view the environment status in the Kosli UI by navigating to: `Environments > terraform-state`.
At this point you should see one artifact with a compliant status since we have provenance for the change that happened.

{{< figure src="/images/tutorials/iac-changes/authorized-iac-change.png" alt="Environment shows an authorized change" width="90%" >}}

## Introducing an unauthorized change

Now let's see how Kosli can help catching an unauthorized change. 
We can simulate such change by modifying the `random_pet_result` output on line 6 in main.tf to `random_pet_name` and running:

```shell {.command}
terraform apply --auto-approve
```

This updates the state file. Let's report the updated state file to the Kosli environment.

{{<hint info>}}
In production, this step won't be necessary because you would have configured environment reporting to happen
automatically (either on state file change or periodically).
{{</hint>}}

```shell {.command}
kosli snapshot path terraform-state --name=tf-state --path=terraform.tfstate
```

Getting the latest snapshot of the environment by running the command below shows that the `FLOW` is unknown. 
This means that Kosli does not have provenance for that change (i.e. it is an unauthorized change).

```shell
kosli get snapshot terraform-state
COMMIT  ARTIFACT                                                                       FLOW  COMPLIANCE     RUNNING_SINCE   REPLICAS
N/A     Name: tf-state                                                                 N/A   NON-COMPLIANT  8 minutes ago  1
        Fingerprint: edd93dcde27718ed493222ceb218275655555f3f3bfefa95628c599e678ac325
```

When you navigate to the environment page again, you will see a non-compliant artifact running.

{{< figure src="/images/tutorials/iac-changes/unauthorized-iac-change.png" alt="Environment shows an unauthorized change" width="90%" >}}

## Next steps

Now that we can detect unauthorized changes in Terraform IaC, the next step would be to receive notifications or
trigger automated actions when this happens. You can achieve that by configuring [Kosli actions](/integrations/actions/).
