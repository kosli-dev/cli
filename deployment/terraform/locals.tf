locals {
  secret_prefix   = "arn:aws:secretsmanager:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:secret:${var.app_name}/"
  private_subnets = split(",", data.aws_ssm_parameter.vpc_private_subnets.value)
}
