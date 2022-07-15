data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_ssm_parameter" "merkely_api_token" {
  name = "merkely_api_token"
}
