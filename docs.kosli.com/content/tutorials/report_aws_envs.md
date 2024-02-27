---
title: "How to report ECS, Lambda and S3 environments"
bookCollapseSection: false
weight: 508
---

# How to report ECS, Lambda and S3 environments

Kosli environments allow you to track changes in your physical/virtual runtime environments. Such changes must be reported from the runtime environment to Kosli.

This tutorial shows you how to set up reporting of running artifacts from a Kubernetes cluster to Kosli.


## Different ways for reporting

There are two different ways to report what's running in a Kubernetes cluster:

- Using Kosli CLI (suitable for testing only)
- Using the [Kosli terraform module](https://registry.terraform.io/modules/kosli-dev/kosli-reporter/aws/latest) to setup a Lambda function to be triggered on AWS changes and report to Kosli.

We describe how to use the different options below and you can choose what suites your needs.

## Prerequisites

To follow this tutorial, you will need to:

- Have access to AWS.
- [Create a Kosli account](https://app.kosli.com/sign-up) if you have not got one already.
- [Create an ECS, Lambda or S3 Kosli environment](/getting_started/environments/#create-an-environment) named `aws-env-tutorial` 
- [Get a Kosli API token](/getting_started/service-accounts/)
- [Install Kosli CLI](/getting_started/install/) (only needed if you will report using CLI)
- [Install Terraform](https://developer.hashicorp.com/terraform/install) (only needed if you will use the Kosli terraform module)

## Report snapshots using Kosli CLI

This option is **only suitable for testing purposes**.  
You need to create an AWS static credentials or equivalent and export the following environments variables:

```shell {.command}
export AWS_REGION=yourAWSRegion
export AWS_ACCESS_KEY_ID=yourAWSAccessKeyID
export AWS_SECRET_ACCESS_KEY=yourAWSSecretAccessKey
```

{{< tabs "snapshot env" "col-no-wrap" >}}

{{< tab "ECS" >}}
```shell {.command}
$ kosli snapshot ecs aws-env-tutorial \
    --cluster <your-ecs-cluster-name> \
	--api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```
{{< /tab >}}

{{< tab "Lambda" >}}
```shell {.command}
$ kosli snapshot lambda aws-env-tutorial \
    --function-names function1,function2 \
	--api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```
{{< /tab >}}

{{< tab "S3" >}}
```shell {.command}
$ kosli snapshot s3 aws-env-tutorial \
    --bucket <your-bucket-name> \
	--api-token <your-api-token-here> \
    --org <your-kosli-org-name>
```
{{< /tab >}}

{{< /tabs >}}


## Report snapshots using Terraform module

You can use the Kosli reporter terraform module to setup a Lambda function which is triggered every time your ECS cluster, Lambda function(s) or S3 bucket changes. The Lambda function will report the running artifacts to Kosli by running the Kosli CLI.

To setup the Lambda function using terraform, you need to follow these steps:

1. [Authenticate to AWS](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html)
   
2. Store the Kosli API key value in an AWS SSM parameter (SecureString type). By default, the Lambda Reporter function will search for the kosli_api_token SSM parameter, but it is also possible to set custom parameter name using kosli_api_token_ssm_parameter_name variable.
   
3. Create a Terraform configuration by copying one of the examples below into a `main.tf` file.

{{< tabs "terraform aws env" "col-no-wrap" >}}

{{< tab "ECS" >}}
```hcl {.command}
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.1"
    }
  }
}

provider "aws" {
  region = local.region

  # Make it faster by skipping some checks
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

locals {
  reporter_name = "reporter-${random_pet.this.id}"
  region        = "eu-central-1"
}

data "aws_caller_identity" "current" {}

data "aws_canonical_user_id" "current" {}

resource "random_pet" "this" {
  length = 2
}

module "lambda_reporter" {
  source  = "kosli-dev/kosli-reporter/aws"
  version = "0.5.3"

  name                              = local.reporter_name
  kosli_environment_type            = "ecs"
  kosli_cli_version                 = "v2.7.8"
  kosli_environment_name            = "aws-env-tutorial"
  kosli_org                         = "<your-org-name>"
  reported_aws_resource_name        = "<your-ecs-cluster>"
}
```
{{< /tab >}}


{{< tab "Lambda" >}}
```hcl {.command}
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.1"
    }
  }
}

provider "aws" {
  region = local.region

  # Make it faster by skipping some checks
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

locals {
  reporter_name = "reporter-${random_pet.this.id}"
  region        = "eu-central-1"
}

data "aws_caller_identity" "current" {}

data "aws_canonical_user_id" "current" {}

resource "random_pet" "this" {
  length = 2
}

variable "my_lambda_functions" {
  type    = string
  default = "function_name1, function_name2"
}

module "lambda_reporter" {
  source  = "kosli-dev/kosli-reporter/aws"
  version = "0.5.3"

  name                           = local.reporter_name
  kosli_environment_type         = "lambda"
  kosli_cli_version              = "v2.7.8"
  kosli_environment_name         = "aws-env-tutorial"
  kosli_org                      = "<your-org-name>"
  reported_aws_resource_name     = var.my_lambda_functions
  use_custom_eventbridge_pattern = true
  custom_eventbridge_pattern     = local.custom_event_pattern
}

locals {
  lambda_function_names_list = split(",", var.my_lambda_functions)

  custom_event_pattern = jsonencode({
    source      = ["aws.lambda"]
    detail-type = ["AWS API Call via CloudTrail"]
    detail = {
      requestParameters = {
        functionName = local.lambda_function_names_list
      }
      responseElements = {
        functionName = local.lambda_function_names_list
      }
    }
  })
}
```
{{< /tab >}}

{{< tab "S3" >}}
```hcl {.command}
terraform {
  required_version = ">= 1.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.63"
    }
    random = {
      source  = "hashicorp/random"
      version = ">= 3.5.1"
    }
  }
}

provider "aws" {
  region = local.region

  # Make it faster by skipping some checks
  skip_metadata_api_check     = true
  skip_region_validation      = true
  skip_credentials_validation = true
  skip_requesting_account_id  = true
}

locals {
  reporter_name = "reporter-${random_pet.this.id}"
  region        = "eu-central-1"
}

data "aws_caller_identity" "current" {}

data "aws_canonical_user_id" "current" {}

resource "random_pet" "this" {
  length = 2
}

variable "my_lambda_functions" {
  type    = string
  default = "my_lambda_function1, my_lambda_function_name2"
}

module "lambda_reporter" {
  source  = "kosli-dev/kosli-reporter/aws"
  version = "0.5.3"

  name                       = local.reporter_name
  kosli_environment_type     = "s3"
  kosli_cli_version          = "v2.7.8"
  kosli_environment_name     = "aws-env-tutorial"
  kosli_org                  = "<your-org-name>"
  reported_aws_resource_name = "<your-s3-bucket-name>"
}
```
{{< /tab >}}

{{< /tabs >}}

4. Initialize and run Terraform by running:

```shell {.command}
$ terraform init
$ terraform apply
```

5. To check Lambda reporter logs you can go to the AWS console -> Lambda service -> choose your lambda reporter function -> Monitor tab -> Logs tab.