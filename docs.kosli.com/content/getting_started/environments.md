---
title: "Part 3: Environments"
bookCollapseSection: false
weight: 220
---
# Part 3: Environments

Recording the status of runtime environments is one of the fundamental features of Kosli. Kosli records the status of runtime environments by detecting artifacts running in any given environment and reporting the information.

## Create an environment

A Kosli *environment* stores snapshots containing information about the software artifacts that are running in your runtime environments. 

Before you start reporting what's running in your environments you need to create an environment in Kosli and make sure it matches the type of the environment you'll be reporting, e.g. `docker` or `k8s`. You can see all the available environment types in the help text for the `--environment-type` flag in the [`kosli create environment`](/client_reference/kosli_create_environment/) command. 

{{< hint warning >}}
In all the commands below we skip required `--api-token` and `--org` flags - these can be easily configured via [config file](/kosli_overview/kosli_tools/#config-file) or [environment variables](/kosli_overview/kosli_tools/#environment-variables) so you don't have type them over and over again.
{{< /hint >}}


### Example

#### CLI
{{< tabs "create env" "col-no-wrap" >}}

{{< tab "v2" >}}
```shell {.command}
$ kosli create environment quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"

environment quickstart was created
```
{{< /tab >}}

{{< tab "v0.1.x" >}}
```shell {.command}
$ kosli environment declare \
    --name quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"

environment quickstart was created
```
{{< /tab >}}

{{< /tabs >}}


You can verify that the Kosli environment called *quickstart* was created:

{{< tabs "ls env" "col-no-wrap" >}}

{{< tab "v2" >}}
```shell {.command}
$ kosli ls environments

NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```
{{< /tab >}}

{{< tab "v0.1.x" >}}
```shell {.command}
$ kosli environment ls

NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```
{{< /tab >}}

{{< /tabs >}}


#### UI

You can also create an environment directly from [app.kosli.com](https://app.kosli.com).

Make sure you've selected the organization you want to use (`docs-demo` here) and click on "Environments". You'll find an "Add new environment" button there

{{<figure src="/images/add-new-env.png" alt="Add new Environment" width="600">}}

After the new environment is created you'll be redirected to its page - with "No events have been found for [...]" message. Once you start reporting your actual runtime environment to Kosli you'll see all the events (such as which artifacts started or stopped running) listed on that page.

## Report an environment

There is range of `kosli snapshot [...]` commands, allowing you to report a variety of environments. To record the current status of your environment you simply run one of them. You can do it manually but typically recording commands would run automatically, e.g. via a cron job or scheduled CI job.

Whenever an environment report is received, if the received list of running artifacts is different than what was reported previously a new snapshot is created. Snapshots are immutable and can't be tampered with.

After you started reporting, you can - at any point - check exactly what is running in your environment using the CLI command:

{{< tabs "get env" "col-no-wrap" >}}

{{< tab "v2" >}}
```shell {.command}
$ kosli get snapshot quickstart

COMMIT   ARTIFACT                                                                       FLOW  RUNNING_SINCE  REPLICAS
9f14efa  Name: nginx:1.21                                                               N/A   18 hours ago   1
         Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514
```
{{< /tab >}}

{{< tab "v0.1.x" >}}
```shell {.command}
$ kosli environment get quickstart

COMMIT   ARTIFACT                                                                       FLOW  RUNNING_SINCE  REPLICAS
9f14efa  Name: nginx:1.21                                                               N/A   18 hours ago   1
         Fingerprint: 2bcabc23b45489fb0885d69a06ba1d648aeda973fae7bb981bafbb884165e514  
```
{{< /tab >}}

{{< /tabs >}}



Or in the UI, by clicking on the name of the environment (after selecting "Environments" in the left hand side menu):

{{<figure src="/images/env-snap-1.png" alt="Environment, Snapshot #1" width="900">}}


{{< tabs "env-reports" "col-no-wrap" >}}

{{< tab "docker" >}}
## Record docker environment

Run `kosli snapshot docker` to report running containers data from docker host to Kosli.  

**Where to run:** The command has to be run on the actual docker host, to be able to detect running containers.

### Example

```shell {.command}
$ kosli snapshot docker docs-demo-docker

[1] containers were reported to environment quickstart
```

More details in [kosli snapshot docker](/client_reference/kosli_snapshot_docker/) reference  
for v0.1.x: [kosli environment report docker](/legacy_ref/v0.1.41/kosli_environment_report_docker/) 
{{< /tab >}}

{{< tab "ecs" >}}
## Record ecs environment

Run `kosli snapshot ecs` to report images data from AWS ECS cluster to Kosli.  

**Were to run:**  The command can be run anywhere.  
To authenticate to AWS, you can either: 
1. provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
2. export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
3. Use a shared config/credentials file under the $HOME/.aws  

Option 1 takes highest precedence, while option 3 is the lowest.


### Example

```shell {.command}
$ kosli snapshot ecs ecs-prod \
	--cluster prod-cluster
	--aws-key-id *** \
	--aws-secret-key *** \
	--aws-region eu-central-1 

[2] containers were reported to environment ecs-prod
```

More details in [kosli snapshot ecs](/client_reference/kosli_snapshot_ecs/) reference  
for v0.1.x: [kosli environment report ecs](/legacy_ref/v0.1.41/kosli_environment_report_ecs/) 
{{< /tab >}}

{{< tab "k8s" >}}
## Record k8s environment

Run `kosli snapshot k8s` to report images data from specific namespace(s) or entire cluster to Kosli. You can also select multiple namespaces to report from (using `--namespace` and comma separated list when running a command) or use `--exclude-namespace` to report from a whole cluster except the namespaces from the comma spearated list given to the flag

**Were to run:**  The command can be run anywhere and requires `kubeconfig` file to be able to connect to the cluster (you can skip providing the location of `kubeconfig` if it resides in default `$HOME/.kube/config` folder).

You can also choose to run it from within the cluster - use our [helm chart](/helm/) to install the reporter as a cron job. `kubeconfig` won't be need in that case.

### Example

```
# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli snapshot k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--org yourOrgName

# report what is running in a given namespace using kubeconfig at a custom path:
kosli snapshot k8s yourEnvironmentName \
	--kubeconfig /path/to/kubeconfig \
	--namespace your-namespace \
	--api-token yourAPIToken \
	--org yourOrgName

```

More details in [kosli snapshot k8s](/client_reference/kosli_snapshot_k8s/) reference  
for v0.1.x: [kosli environment report k8s](/legacy_ref/v0.1.41/kosli_environment_report_k8s/) 
{{< /tab >}}

{{< tab "lambda" >}}
## Record lambda environment

Run `kosli snapshot lambda` to report artifact from AWS Lambda to Kosli.  

**Were to run:**  The command can be run anywhere.   
To authenticate to AWS, you can either: 
1. provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
2. export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
3. Use a shared config/credentials file under the $HOME/.aws  

Option 1 takes highest precedence, while option 3 is the lowest.

### Example

```shell {.command}
$ kosli snapshot lambda lambda-prod \
	--function-name reporter-kosli-prod \
	--aws-key-id *** \
	--aws-secret-key *** \
	--aws-region eu-central-1 

reporter-app-prod lambda function was reported to environment lambda-prod
```

More details in [kosli snapshot lambda](/client_reference/kosli_snapshot_lambda/) reference   
for v0.1.x: [kosli environment report lambda](/legacy_ref/v0.1.41/kosli_environment_report_lambda/) 
{{< /tab >}}

{{< tab "s3" >}}
## Record s3 environment

Run `kosli snapshot s3` to report artifact from AWS S3 bucket to Kosli.  

**Were to run:**  The command can be run anywhere.   
To authenticate to AWS, you can either: 
1. provide the AWS static credentials via flags or by exporting the equivalent KOSLI env vars (e.g. KOSLI_AWS_KEY_ID)
2. export the AWS env vars (e.g. AWS_ACCESS_KEY_ID).
3. Use a shared config/credentials file under the $HOME/.aws  

Option 1 takes highest precedence, while option 3 is the lowest.

### Example

```shell {.command}
$ kosli snapshot s3 s3-prod \
 	--bucket app-public \
	--aws-key-id *** \
	--aws-secret-key *** \
	--aws-region eu-central-1 

bucket app-public was reported to environment s3-prod
```

More details in [kosli snapshot s3](/client_reference/kosli_snapshot_s3/) reference  
for v0.1.x: [kosli environment report s3](/legacy_ref/v0.1.41/kosli_environment_report_s3/) 
{{< /tab >}}

{{< tab "server" >}}
## Record server environment

Run `kosli snapshot server` to report directory or file artifacts from the given list of paths to Kosli.  

**Were to run:**  The command has to be run on the actual server (physical or vm), to be able to detect artifacts. 

Use `--paths` flag to provide a comma separated list of directories and files you want to be reported. Keep in mind that each directory will be treated as a single artifact and in order to make sure they are correctly identified in Kosli they should also be reported to Kosli flow as a single artifact.

For example, if you provide a following list: `--paths /home/server/web, /home/monitor.exe, /home/server/calculator` kosli will calculate fingerprints and report as running 3 artifacts to Kosli:
* directory `web`
* directory `calculator` 
* file `monitor.exe`

And it will try to find matching artifacts reported to any flow belonging to the same organization as the environment.

### Example 

```shell {.command}
$ kosli snapshot server docs-demo-server --paths build/index.html 

[1] artifacts were reported to environment docs-demo-server       
```

More details in [kosli snapshot server](/client_reference/kosli_snapshot_server/)reference  
for v0.1.x: [kosli environment report server](/legacy_ref/v0.1.41/kosli_environment_report_server/) 
{{< /tab >}}

{{< /tabs >}}








