resource "aws_cloudwatch_event_rule" "cron_every_minute" {
  name        = "run-${var.name}-lambda-reporter"
  description = "Execute ${var.name} lambda reporter"

  schedule_expression = "rate(1 minute)"
}

resource "aws_cloudwatch_event_target" "lambda_reporter" {
  arn       = module.reporter_lambda.lambda_function_arn
  rule      = aws_cloudwatch_event_rule.cron_every_minute.name
  target_id = module.reporter_lambda.lambda_function_name
}
