data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_ssm_parameter" "vpc_private_subnets" {
  name = "vpc_private_subnets"
}

data "aws_ssm_parameter" "vpc_id" {
  name = "vpc_id"
}

data "aws_ecs_cluster" "this" {
  cluster_name = "merkely"
}