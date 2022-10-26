---
title: Record
bookCollapseSection: false
weight: 30
---
# Record your environments in Kosli

Recording the status of runtime environments it's one of the fundamental features of Kosli. Our CLI detects artifacts running in givent environment and reports the information to Kosli. 

If the list of running artifacts is different than what was reported previously a new snapshot is created. Snapshots are immutable and can't be tampered with.

There is range of `kosli environment report [...]` commands, allowing you to report a variety of environments. To record a current status of your environment you simnply run on of them. You can do it manually but typically recording commands would run automatically, e.g. via a cron job or scheduled CI job.

## Record docker environment

Run `kosli environment report docker` to report running containers data from docker host to Kosli.  
The command has to be run on the actual docker host, to be able to detect running containers.

### Example

```
kosli environment report docker yourEnvironmentName \
	--api-token yourAPIToken \r
	--owner yourOrgName
```

Details [here](/client_reference/kosli_environment_report_docker/)

## Record ecs environment

Run `kosli environment report ecs` to report images data from AWS ECS cluster to Kosli.  
The command requires following environment variables to be set, to be able to connect to AWS:
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

Details [here](/client_reference/kosli_environment_report_ecs/)

## Record k8s environment

Run `kosli environment report k8s` to report images data from specific namespace(s) or entire cluster to Kosli.  
The command can be run anywhere and requires `kubeconfig` to be able to connect to the cluster (you can skip providing the location of `kubeconfig` if it resides in default `$HOME/.kube/config` folder).

### Example

```
# report what is running in an entire cluster using kubeconfig at $HOME/.kube/config:
kosli environment report k8s yourEnvironmentName \
	--api-token yourAPIToken \
	--owner yourOrgName

# report what is running in a given namesoace using kubeconfig at a custom path:
kosli environment report k8s yourEnvironmentName \
	--kubeconfig /path/to/kubeconfig \
	--namespace your-namespace \
	--api-token yourAPIToken \
	--owner yourOrgName

```

Details [here](/client_reference/kosli_environment_report_k8s/)

## Record lambda environment

Run `kosli environment report lambda` to report artifact from AWS Lambda to Kosli.  
You can use either flags or environment variables to provide AWS secrets.

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

Details [here](/client_reference/kosli_environment_report_lambda/)

## Record s3 environment

Run `kosli environment report s3` to report artifact from AWS S3 bucket to Kosli.  
You can use either flags or environment variables to provide AWS secrets.

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

Details [here](/client_reference/kosli_environment_report_s3/)

## Record server environment

Run `kosli environment report server` to report directory or file artifacts data in the given list of paths to Kosli.  
The command has to be run on the actual server (physical or vm), to be able to detect artifacts. 

Use `--paths` flag to provide a comma separated list of directories and files you want to be reported. Keep in mind that each directory will be treated as a single artifact and in order to make sure they are correctly identified in Kosli they should also be reported as a single artifact.

### Example 

```
kosli environment report server yourEnvironmentName \
	--paths a/b/c, e/f/g \
	--api-token yourAPIToken \
	--owner yourOrgName
```


Details [here](/client_reference/kosli_environment_report_server/)


