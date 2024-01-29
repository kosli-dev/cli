terraform {
  backend "s3" {}
}

provider "aws" {
  alias  = "eu-central-1"
  region = "eu-central-1"
}

module "tags" {
  source            = "fivexl/tag-generator/aws"
  version           = "2.0.0"
  terraform_managed = "1"
  environment_name  = "prod"
  data_owner        = "kosli"
}

module "evidence-reporter" {
  source  = "../../terraform-aws-evidence-reporter"

  log_uploader_name        = "evidence-log-uploader-test"
  identity_reporter_name   = "evidence-identity-reporter-test"
  kosli_org_name           = "test-org"
  ecs_exec_log_bucket_name = "ecs-exec-logs-e517928541bdf4e16ed019571c54eeb88689aec1"
  recreate_missing_package = false
  kosli_flow_name   = "a-test-workflow"
  kosli_cli_version        = "trail2"
  reporter_releases_host = "https://reporter-releases.kosli.com"
  kosli_api_token_ssm_parameter_name = "kosli_api_token_test"
  kosli_host = "https://staging.app.kosli.com/"
  tags                     = module.tags.result
}
