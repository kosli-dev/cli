---
title: "Part 2: Environments"
bookCollapseSection: false
weight: 220
---
# Part 2: Environments

Recording the status of runtime environments is one of the fundamental features of Kosli. Kosli records the status of runtime environments by detecting artifacts running in any given environment and reporting the information.

If the list of running artifacts is different than what was reported previously a new snapshot is created. Snapshots are immutable and can't be tampered with.

There is range of `kosli environment report [...]` commands, allowing you to report a variety of environments. To record the current status of your environment you simply run one of them. You can do it manually but typically recording commands would run automatically, e.g. via a cron job or scheduled CI job.


{{< hint warning >}}
In all the commands below we skip required `--api-token` and `--owner` flags - these can be easily configured via [config file](/kosli_overview/kosli_tools/#config-file) or [environment variables](/kosli_overview/kosli_tools/#environment-variables) so you don't have type them over and over again.
{{< /hint >}}


After you started reporting, you can - at any point - check exactly what is running in your environment using the CLI command:

```shell {.command}
$ kosli environment get quickstart

COMMIT  ARTIFACT                                                           PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: nginx@sha256:0047b7(...))59cce6d40291ccfb4e039f5dc7efd33286  N/A       7 days ago     1
        Fingerprint: 0047b729188(...)959cce6d40291ccfb4e039f5dc7efd33286   
```

Or in the UI, by clicking at the name of the environment (after selecting "Environments" in the left hand side menu):

{{<figure src="/images/env-snap-1.png" alt="Environment, Snapshot #1" width="900">}}

## Create an environment

A Kosli *environment* stores snapshots containing information about the software artifacts that are running in your runtime environments. 

Before you start reporting what's running in your environments you need to create an environment in Kosli and make sure it matches the type of the environment you'll be reporting, e.g. `docker` or `k8s`. You can see all the available environment types in the help text for the `--environment-type` flag in the [`kosli environment declare`](/client_reference/kosli_environment_declare/) command. 

### Example

#### CLI

```shell {.command}
$ kosli environment declare \
    --name quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"

environment quickstart was created
```
```plaintext {.light-console}
```

You can verify that the Kosli environment called *quickstart* was created:

```shell {.command}
$ kosli environment ls

NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```

#### UI

You can also create an environment directly from [app.kosli.com](https://app.kosli.com).

Make sure you've selected the organization you want to use (`docs-demo` here) and click on "Environments". You'll find an "Add new environment" button there

{{<figure src="/images/add-new-env.png" alt="Add new Environment" width="600">}}

Fill in the form - type, name and a description - and click "Save Environment" button

{{<figure src="/images/save-env.png" alt="Save Environment" width="650">}}

After the new environment is created you'll be redirected to its page - with "No events have been found for [...]" message. Once you start reporting your actual runtime environment to Kosli you'll see all the events (such as which artifacts started or stopped running) listed on that page.

To see the list of all your environments, just click on the "Environments" again

{{<figure src="/images/env-list.png" alt="Environments" width="900">}}

## Report an environment

{{< tabs "env-reports" "col-no-wrap" >}}

{{< tab "docker" >}}
## Record docker environment

Run `kosli environment report docker` to report running containers data from docker host to Kosli.  

**Where to run:** The command has to be run on the actual docker host, to be able to detect running containers.

### Example

```shell {.command}
$ kosli environment report docker docs-demo-docker

[1] containers were reported to environment quickstart
```

More details in [`kosli environment report docker` reference](/client_reference/kosli_environment_report_docker/)
{{< /tab >}}

{{< tab "ecs" >}}
## Record ecs environment

Run `kosli environment report ecs` to report images data from AWS ECS cluster to Kosli.  

**Were to run:**  The command can be run anywhere.  
To authenticate to AWS, you can either: 
1. provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
2. export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
3. Use a shared config/credentials file under the $HOME/.aws  

Option 1 takes highest precedence, while option 3 is the lowest.


### Example

```shell {.command}
$ kosli environment report ecs ecs-prod \
	--cluster prod-cluster
	--aws-key-id *** \
	--aws-secret-key *** \
	--aws-region eu-central-1 

[2] containers were reported to environment ecs-prod
```

More details in [`kosli environment report ecs` reference](/client_reference/kosli_environment_report_ecs/)
{{< /tab >}}

{{< tab "k8s" >}}
## Record k8s environment

Run `kosli environment report k8s` to report images data from specific namespace(s) or entire cluster to Kosli. You can also select multiple namespaces to report from (using `--namespace` and comma separated list when running a command) or use `--exclude-namespace` to report from a whole cluster except the namespaces from the comma spearated list given to the flag

**Were to run:**  The command can be run anywhere and requires `kubeconfig` file to be able to connect to the cluster (you can skip providing the location of `kubeconfig` if it resides in default `$HOME/.kube/config` folder).

You can also choose to run it from within the cluster - use our [helm chart](/helm/) to install the reporter as a cron job. `kubeconfig` won't be need in that case.

### Example

```
# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a given namespace using kubeconfig at a custom path:
kosli environment report k8s yourEnvironmentName \
	--kubeconfig /path/to/kubeconfig \
	--namespace your-namespace \
	--api-token yourAPIToken \
	--owner yourOrgName

```

More details in [`kosli environment report k8s` reference](/client_reference/kosli_environment_report_k8s/)
{{< /tab >}}

{{< tab "lambda" >}}
## Record lambda environment

Run `kosli environment report lambda` to report artifact from AWS Lambda to Kosli.  

**Were to run:**  The command can be run anywhere.   
To authenticate to AWS, you can either: 
1. provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
2. export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
3. Use a shared config/credentials file under the $HOME/.aws  

Option 1 takes highest precedence, while option 3 is the lowest.

### Example

```shell {.command}
$ kosli environment report lambda lambda-prod \
	--function-name reporter-kosli-prod \
	--aws-key-id *** \
	--aws-secret-key *** \
	--aws-region eu-central-1 

reporter-app-prod lambda function was reported to environment lambda-prod
```

More details in [`kosli environment report lambda` reference](/client_reference/kosli_environment_report_lambda/)
{{< /tab >}}

{{< tab "s3" >}}
## Record s3 environment

Run `kosli environment report s3` to report artifact from AWS S3 bucket to Kosli.  

**Were to run:**  The command can be run anywhere.   
To authenticate to AWS, you can either: 
1. provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
2. export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
3. Use a shared config/credentials file under the $HOME/.aws  

Option 1 takes highest precedence, while option 3 is the lowest.

### Example

```shell {.command}
$ kosli environment report s3 s3-prod \
 	--bucket app-public \
	--aws-key-id *** \
	--aws-secret-key *** \
	--aws-region eu-central-1 

bucket app-public was reported to environment s3-prod
```

More details in [`kosli environment report s3` reference](/client_reference/kosli_environment_report_s3/)
{{< /tab >}}

{{< tab "server" >}}
## Record server environment

Run `kosli environment report server` to report directory or file artifacts from the given list of paths to Kosli.  

**Were to run:**  The command has to be run on the actual server (physical or vm), to be able to detect artifacts. 

Use `--paths` flag to provide a comma separated list of directories and files you want to be reported. Keep in mind that each directory will be treated as a single artifact and in order to make sure they are correctly identified in Kosli they should also be reported to Kosli pipeline as a single artifact.

For example, if you provide a following list: `--paths /home/server/web, /home/monitor.exe, /home/server/calculator` kosli will calculate fingerprints and report as running 3 artifacts to Kosli:
* directory `web`
* directory `calculator` 
* file `monitor.exe`

And it will try to find matching artifacts reported to any pipeline belonging to the same organization as the environment.

### Example 

```shell {.command}
$ kosli environment report server docs-demo-server --paths build/index.html 

[1] artifacts were reported to environment docs-demo-server       
```

More details in [`kosli environment report server` reference](/client_reference/kosli_environment_report_server/)
{{< /tab >}}

{{< /tabs >}}








