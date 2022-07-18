data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_ssm_parameter" "oidc_role_arn" {
  count = var.create_public_ecr ? 1 : 0
  name  = "oidc_role_arn"
}
