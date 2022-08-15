locals {
  package_url = "${var.reporter_releases_host}/kosli_lambda_${var.kosli_cli_version}.zip"
  downloaded  = "downloaded_package_${md5(local.package_url)}.zip"
}

resource "null_resource" "download_package" {
  triggers = {
    downloaded = local.downloaded
  }

  provisioner "local-exec" {
    command = "curl -L -o ${local.downloaded} ${local.package_url}"
  }
}

data "null_data_source" "downloaded_package" {
  inputs = {
    id       = null_resource.download_package.id
    filename = local.downloaded
  }
}

module "reporter_lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "3.3.1"

  attach_policy_json = true
  policy_json        = data.aws_iam_policy_document.ecs_list_allow.json

  function_name          = "reporter-${var.name}"
  description            = "Send reports to the Kosli app"
  handler                = "function.handler"
  runtime                = "provided"
  local_existing_package = data.null_data_source.downloaded_package.outputs["filename"]

  role_name      = var.name
  timeout        = 30
  create_package = false
  publish        = true

  environment_variables = {
    MERKELY_HOST      = var.kosli_host
    MERKELY_API_TOKEN = data.aws_ssm_parameter.kosli_api_token.value
    ENV               = var.env
    ECS_CLUSTER       = var.ecs_cluster
    KOSLI_USER        = var.kosli_user
  }

  allowed_triggers = {
    AllowExecutionFromCloudWatch = {
      principal  = "events.amazonaws.com"
      source_arn = aws_cloudwatch_event_rule.cron_every_minute.arn
    }
  }

  cloudwatch_logs_retention_in_days = var.cloudwatch_logs_retention_in_days

  tags = var.tags
}
