---
title: Record
bookCollapseSection: false
weight: 30
---
# Record your environments in Kosli

Recording the status of runtime environments it's one of the fundamental features of Kosli. Our CLI detects artifacts running in givent environment and reports the information to Kosli. 

If the list of running artifacts is different than what was reported previously a new snapshot is created. Snapshots are immutable and can't be tampered with.

There is range of `kosli environment report [...]` commands, allowing you to report a variety of environments. To record a current status of your environment you simnply run on of them. You can do it manually but typically recording commands would run automatically, e.g. via a cron job or scheduled CI job.

## Recording commands

### kosli environment report docker

Report running containers data from docker host to Kosli.

Details [here](/client_reference/kosli_environment_report_docker/)

### kosli environment report ecs

Report images data from AWS ECS cluster to Kosli.

Details [here](/client_reference/kosli_environment_report_ecs/)

### kosli environment report k8s

Report images data from specific namespace(s) or entire cluster to Kosli.

Details [here](/client_reference/kosli_environment_report_k8s/)

### kosli environment report lambda

Report artifact from AWS Lambda to Kosli.

Details [here](/client_reference/kosli_environment_report_lambda/)

### kosli environment report s3

Report artifact from AWS S3 bucket to Kosli.

Details [here](/client_reference/kosli_environment_report_s3/)

### kosli environment report server

Report directory or file artifacts data in the given list of paths to Kosli.

Details [here](/client_reference/kosli_environment_report_server/)


