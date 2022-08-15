resource "aws_ecr_pull_through_cache_rule" "this" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "public.ecr.aws"
}

module "lambda_reporter" {
  for_each     = var.kosli_hosts
  source       = "./lambda-reporter"
  name         = "${var.app_name}-${each.key}"
  env          = var.env
  kosli_host   = each.value
  kosli_cli_version = "1.5.9"
  ecs_cluster       = "merkely"
  kosli_user        = "compliancedb"
  tags = module.tags.result
}
