module "reporter_app" {
  for_each            = var.reporter_apps
  source              = "./ecs-task"
  app_name            = var.app_name
  ecs_cluster_arn     = data.aws_ecs_cluster.this.arn
  env                 = each.key
  subnets             = local.private_subnets
  merkely_env         = var.merkely_env
  merkely_host        = each.value.merkely_host
  ecs_events_role_arn = aws_iam_role.ecs_events.arn
  task_role_arn       = aws_iam_role.task.arn
  execution_role_arn  = aws_iam_role.exec.arn
  cpu_limit           = each.value.cpu_limit
  mem_limit           = each.value.mem_limit
  mem_reservation     = each.value.mem_reservation
  image_tag           = var.IMAGE_TAG
  secret_prefix       = local.secret_prefix
}