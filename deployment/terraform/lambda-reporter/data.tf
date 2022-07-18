data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_ssm_parameter" "kosli_api_token" {
  name = "kosli_api_token"
}
