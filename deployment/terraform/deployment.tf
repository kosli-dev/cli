resource "aws_cloudwatch_event_rule" "this" {
  name        = "run-reporter"
  description = "Execute Merkely reporter ECS task"

  schedule_expression = "cron(*/5 * * * ? *)"
}

resource "aws_iam_role" "ecs_events" {
  name = "ecs_events"

  assume_role_policy = <<DOC
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "events.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
DOC
}

resource "aws_iam_role_policy" "ecs_events_run_task_with_any_role" {
  name = "ecs_events_run_task_with_any_role"
  role = aws_iam_role.ecs_events.id

  policy = <<DOC
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "iam:PassRole",
            "Resource": "*",
            "Condition": {
                "StringLike": {
                    "iam:PassedToService": "ecs-tasks.amazonaws.com"
                }
            }
        },
        {
            "Effect": "Allow",
            "Action": [
                "ecs:RunTask"
            ],
            "Resource": [
                "arn:aws:ecs:*:${data.aws_caller_identity.current.account_id}:task-definition/${var.app_name}:*"
            ],
            "Condition": {
                "ArnLike": {
                    "ecs:cluster": "arn:aws:ecs:*:${data.aws_caller_identity.current.account_id}:cluster/merkely"
                }
            }
        }
    ]
}
DOC
}

resource "aws_cloudwatch_event_target" "ecs_scheduled_task" {
  arn      = data.aws_ecs_cluster.this.arn
  rule     = aws_cloudwatch_event_rule.this.name
  role_arn = aws_iam_role.ecs_events.arn

  ecs_target {
    task_count          = 1
    task_definition_arn = aws_ecs_task_definition.this.arn
    launch_type         = "EC2"
    network_configuration {
      subnets         = local.private_subnets
      security_groups = [module.sg.security_group_id]
    }
  }
}

resource "aws_ecs_task_definition" "this" {
  family                   = var.app_name
  task_role_arn            = aws_iam_role.task.arn
  execution_role_arn       = aws_iam_role.exec.arn
  network_mode             = "awsvpc"
  requires_compatibilities = ["EC2"]

  container_definitions = jsonencode([
    {
      name      = "${var.app_name}-${var.env}"
      image     = var.TAGGED_IMAGE
      command   = ["merkely", "environment", "report", "ecs", "staging-aws", "-C", "merkely", "--owner", "compliancedb"]
      cpu       = var.cpu_limit
      memory    = var.mem_limit
      essential = true
      environment = [
        {
          name  = "MERKELY_HOST"
          value = var.MERKELY_HOST
        },
        {
          name  = "AWS_REGION"
          value = var.aws_region
        }
      ],
      secrets = [
        {
          name      = "MERKELY_API_TOKEN",
          valueFrom = "${local.secret_prefix}merkely-api-token"
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
