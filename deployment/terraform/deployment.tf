# module "reporter_app" {
#   for_each            = var.reporter_apps
#   source              = "./ecs-task"
#   app_name            = var.app_name
#   ecs_cluster_arn     = data.aws_ecs_cluster.this.arn
#   env                 = each.key
#   subnets             = local.private_subnets
#   merkely_env         = var.merkely_env
#   merkely_host        = each.value.merkely_host
#   ecs_events_role_arn = aws_iam_role.ecs_events.arn
#   task_role_arn       = aws_iam_role.task.arn
#   execution_role_arn  = aws_iam_role.exec.arn
#   cpu_limit           = each.value.cpu_limit
#   mem_limit           = each.value.mem_limit
#   mem_reservation     = each.value.mem_reservation
#   image_tag           = var.IMAGE_TAG
#   secret_prefix       = local.secret_prefix
# }

resource "aws_ecr_pull_through_cache_rule" "this" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "public.ecr.aws"
}

module "lambda_reporter" {
  for_each            = var.reporter_apps
  source              = "./lambda-reporter"
  name                = "${var.app_name_lambda}-${each.key}"
  merkely_env         = var.merkely_env
  merkely_host        = each.value.merkely_host
  REPORTER_TAG        = var.REPORTER_TAG
  tags                = module.tags.result
}