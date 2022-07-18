terraform {
  backend "s3" {}
}

provider "aws" {
  alias  = "eu-central-1"
  region = "eu-central-1"
}

provider "aws" {
  alias  = "us-east-1"
  region = "us-east-1"
}

module "tags" {
  source            = "fivexl/tag-generator/aws"
  version           = "2.0.0"
  prefix            = "kosli-cli"
  terraform_managed = "1"
  environment_name  = var.kosli_env
  data_owner        = "kosli"
  data_pci          = "0"
  data_phi          = "0"
  data_pii          = "0"
}