resource "aws_cloudwatch_event_rule" "cron_every_minute" {
  name        = "run-${var.name}-lambda-reporter"
  description = "Execute ${var.name} lambda reporter"

  schedule_expression = "cron(* * * * ? *)"
}

resource "aws_cloudwatch_event_target" "lambda_reporter" {
  arn       = module.reporter_lambda.lambda_function_arn
  rule      = aws_cloudwatch_event_rule.cron_every_minute.name
  target_id = module.reporter_lambda.lambda_function_name
}

module "reporter_lambda" {
  source  = "terraform-aws-modules/lambda/aws"
  version = "3.3.1"

  attach_policy_json = true
  policy_json        = data.aws_iam_policy_document.ecs_list_allow.json

  function_name = "reporter-${var.name}"
  description   = "Send reports to the Kosli app"

  image_uri    = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/ecr-public/c5t5r3f3/kosli-reporter:${var.REPORTER_TAG}"
  package_type = "Image"

  image_config_command = ["kosli", "environment", "report", "ecs", "${var.env}", "-C", "${var.ecs_cluster}", "--owner", "${var.kosli_user}"]

  role_name      = var.name
  timeout        = 30
  create_package = false
  publish        = true

  environment_variables = {
    KOSLI_HOST      = var.kosli_host
    KOSLI_API_TOKEN = data.aws_ssm_parameter.kosli_api_token.value
  }

  allowed_triggers = {
    AllowExecutionFromCloudWatch = {
      principal  = "events.amazonaws.com"
      source_arn = aws_cloudwatch_event_rule.cron_every_minute.arn
    }
  }

  cloudwatch_logs_retention_in_days = var.cloudwatch_logs_retention_in_days

  # To do: integrate aws-lambda-go to the kosli cli so lambda events are processed correctly
  # https://docs.aws.amazon.com/lambda/latest/dg/go-image.html
  # https://docs.aws.amazon.com/lambda/latest/dg/golang-handler.html
  # Set maximum_retry_attempts to 0 to avoid function run retries due to incorrect exit code
  create_async_event_config    = true
  maximum_event_age_in_seconds = 60
  maximum_retry_attempts       = 0

  tags = var.tags
}
