# Create Cloudwatch log group to store app logs.
# https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/WhatIsCloudWatch.html
resource "aws_cloudwatch_log_group" "this" {
  name              = "/ecs/${var.app_name}-${var.env}"
  retention_in_days = var.logs_retention_in_days
}
