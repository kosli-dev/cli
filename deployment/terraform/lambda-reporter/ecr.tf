resource "aws_ecr_pull_through_cache_rule" "this" {
  count = var.create_pull_through_cache_rule ? 1 : 0

  ecr_repository_prefix = var.ecr_repository_prefix
  upstream_registry_url = var.upstream_registry_url
}