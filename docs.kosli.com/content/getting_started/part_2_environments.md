---
title: "Part 2: Environments in Kosli"
bookCollapseSection: false
weight: 220
---
# Part 2: Environments in Kosli

Recording the status of runtime environments is one of the fundamental features of Kosli. Kosli records the status of runtime environments by detecting artifacts running in any given environment and reporting the information.

If the list of running artifacts is different than what was reported previously a new snapshot is created. Snapshots are immutable and can't be tampered with.

There is range of `kosli environment report [...]` commands, allowing you to report a variety of environments. To record a current status of your environment you simply run one of them. You can do it manually but typically recording commands would run automatically, e.g. via a cron job or scheduled CI job.

In all the commands below we skip required `--api-token` and `owner` flags - these can be easily configured via [config file](/faq/#what-is-the---config-file-flag) or [environment variables](/getting_started/step_3/#using-environment-variables) so you don't have type them over and over again.

After you started reporting, you can - at any point - check exactly what is running in your environment using CLI command:

```shell {.command}
kosli environment get quickstart
```
```plaintext {.light-console}
COMMIT  ARTIFACT                                                                             PIPELINE  RUNNING_SINCE  REPLICAS
N/A     Name: nginx@sha256:0047b7(...))59cce6d40291ccfb4e039f5dc7efd33286  N/A       7 days ago     1
        Fingerprint: 0047b729188(...)959cce6d40291ccfb4e039f5dc7efd33286                                 
```

Or with UI, by clicking at the name of the environment (after selecting "Environments" in the left hand side menu):

{{<figure src="/images/env-snap-1.png" alt="Environment, Snapshot #1" width="900">}}

## Create an environment

A Kosli *environment* stores snapshots containing information about the software artifacts that you are running in your runtime environments. 

Before you start reporting what's running in your environments you need to create an environment in Kosli and make sure it matches the type of the environment you'll be reporting, e.g. `docker` or `k8s`. You can see all the available environment types in help text for `--environment-type` flag in  [`kosli environment declare`](/client_reference/kosli_environment_declare/) command. 

### Example

#### CLI

```shell {.command}
kosli environment declare \
    --name quickstart \
    --environment-type docker \
    --description "quickstart environment for tutorial"
```
```plaintext {.light-console}
environment quickstart was created
```

You can verify that the Kosli environment was created:

```shell {.command}
kosli environment ls
```

```plaintext {.light-console}
NAME        TYPE    LAST REPORT  LAST MODIFIED
quickstart  docker               2022-11-01T15:30:56+01:00
```

#### UI

You can also create an environment directly from [app.kosli.com](https://app.kosli.com).

Make sure you've selected the organization you want to use (`docs-demo` here) and click on "Environments". You'll find an "Add new environment" button there

{{<figure src="/images/add-new-env.png" alt="Add new Environment" width="600">}}

Fill in the form - type, name and a description - and click "Save Environment" button

{{<figure src="/images/save-env.png" alt="Save Environment" width="650">}}

After the new environment is created you'll be redirected to its page - with "No events have been found for [...]" message. Once you start reporting your actual runtime environment to Kosli you'll see all the events (like which artifacts started or stopped running) listed on that page.

To see the list of all your environments, just click on the "Environments" again

{{<figure src="/images/env-list.png" alt="Environments" width="900">}}

## Report an environment

{{< tabs "env-reports" "col-no-wrap" >}}

{{< tab "docker" >}}
## Record docker environment

Run `kosli environment report docker` to report running containers data from docker host to Kosli.  

**Were to run:** The command has to be run on the actual docker host, to be able to detect running containers.

### Example

```
kosli environment report docker docs-demo-docker
```
```plaintext {.light-console}
[1] containers were reported to environment quickstart
```
More details in [`kosli environment report docker` reference](/client_reference/kosli_environment_report_docker/)
{{< /tab >}}

{{< tab "ecs" >}}
## Record ecs environment

Run `kosli environment report ecs` to report images data from AWS ECS cluster to Kosli.  

**Were to run:**  The command can be run anywhere and requires following environment variables to be set, to be able to connect to AWS:
* `AWS_REGION`
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

### Example

```
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report ecs yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName
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

**Were to run:**  The command can be run anywhere. You can use either flags or environment variables to provide AWS secrets.

### Example

```
# report what is running in the latest version AWS Lambda function (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report lambda myEnvironment \
	--function-name yourFunctionName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a specific version of an AWS Lambda function (AWS auth provided in flags):
kosli environment report lambda myEnvironment \
	--function-name yourFunctionName \
	--function-version yourFunctionVersion \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName
```

More details in [`kosli environment report lambda` reference](/client_reference/kosli_environment_report_lambda/)
{{< /tab >}}

{{< tab "s3" >}}
## Record s3 environment

Run `kosli environment report s3` to report artifact from AWS S3 bucket to Kosli.  

**Were to run:**  The command can be run anywhere. You can use either flags or environment variables to provide AWS secrets.

### Example

```
# report what is running in an AWS S3 bucket (AWS auth provided in env variables):
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey

kosli environment report s3 yourEnvironmentName \
	--bucket yourBucketName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in an AWS S3 bucket (AWS auth provided in flags):
kosli environment report s3 yourEnvironmentName \
	--bucket yourBucketName \
	--aws-key-id yourAWSAccessKeyID \
	--aws-secret-key yourAWSSecretAccessKey \
	--aws-region yourAWSRegion \
	--api-token yourAPIToken \
	--owner yourOrgName
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
kosli environment report server docs-demo-server --paths build/index.html 
```
```plaintext {.light-console}
[1] artifacts were reported to environment docs-demo-server                               
```

More details in [`kosli environment report server` reference](/client_reference/kosli_environment_report_server/)
{{< /tab >}}

{{< /tabs >}}








