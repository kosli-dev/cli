# Use eventbridge rule to trigger ECS task execution
resource "aws_cloudwatch_event_rule" "this" {
  name        = "run-reporter-${var.env}"
  description = "Execute Merkely ${var.env} reporter ECS task"

  schedule_expression = "cron(* * * * ? *)"
}

resource "aws_cloudwatch_event_target" "ecs_scheduled_task" {
  arn      = var.ecs_cluster_arn
  rule     = aws_cloudwatch_event_rule.this.name
  role_arn = var.ecs_events_role_arn

  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.this.arn
    launch_type         = "EC2"
  }
}

resource "aws_ecs_task_definition" "this" {
  family                   = "${var.app_name}-${var.env}"
  task_role_arn            = var.task_role_arn
  execution_role_arn       = var.execution_role_arn
  network_mode             = "bridge"
  requires_compatibilities = ["EC2"]

  container_definitions = jsonencode([
    {
      name      = "${var.app_name}-${var.env}"
      image     = "ghcr.io/kosli-dev/${var.app_name}:${var.image_tag}"
      command   = ["merkely", "environment", "report", "ecs", "${var.merkely_env}", "-C", "merkely", "--owner", "compliancedb"]
      essential = true

      cpu               = var.cpu_limit
      memory            = var.mem_limit
      memoryReservation = var.mem_reservation
      environment = [
        {
          name  = "MERKELY_HOST"
          value = var.merkely_host
        },
        {
          name  = "AWS_REGION"
          value = data.aws_region.current.name
        }
      ],
      secrets = [
        {
          name      = "MERKELY_API_TOKEN",
          valueFrom = "${var.secret_prefix}merkely-api-token"
        }
      ]
      logConfiguration = {
        logDriver = "awslogs",
        options = {
          awslogs-region        = data.aws_region.current.name,
          awslogs-group         = aws_cloudwatch_log_group.this.name,
          awslogs-stream-prefix = format("/%s-%s", var.app_name, var.env)
        }
      }
    }
  ])
}
