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

data "aws_ssm_parameter" "merkely_api_token" {
  name = "merkely_api_token"
}

resource "aws_ecr_repository" "this" {
  name = "${var.app_name_lambda}"
  image_scanning_configuration {
    scan_on_push = "true"
  }
  encryption_configuration {
    encryption_type = "AES256"
  }
  tags = module.tags.result
}

resource "aws_ecr_lifecycle_policy" "this" {
  repository = aws_ecr_repository.this.name
  policy     = <<EOF
{
    "rules": [
        {
            "rulePriority": 1,
            "description": "Expire untagged images older than 30 days",
            "selection": {
                "tagStatus": "untagged",
                "countType": "sinceImagePushed",
                "countUnit": "days",
                "countNumber": 30
            },
            "action": {
                "type": "expire"
            }
        },
        {
            "rulePriority": 2,
            "description": "Expire images if there we are approaching limit",
            "selection": {
                "tagStatus": "any",
                "countType": "imageCountMoreThan",
                "countNumber": 300
            },
            "action": {
                "type": "expire"
            }
        }
    ]
}
EOF
}

resource "aws_ecr_pull_through_cache_rule" "this" {
  ecr_repository_prefix = "ecr-public"
  upstream_registry_url = "public.ecr.aws"
}

module "reporter_lambda" {
  for_each = var.reporter_apps

  source  = "terraform-aws-modules/lambda/aws"
  version = "3.3.1"

  attach_policy_json = true
  policy_json = data.aws_iam_policy_document.ecs_list_allow.json

  function_name = "${var.app_name_lambda}-${each.key}"
  description   = "Send reports to the Kosli app"

  create_package = false

  image_uri    = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${data.aws_region.current.name}.amazonaws.com/ecr-public/kosli/kosli-reporter:${var.REPORTER_TAG}"
  package_type = "Image"

  image_config_command = ["merkely", "environment", "report", "ecs", "${var.merkely_env}", "-C", "merkely", "--owner", "compliancedb"]

  role_name     = "${var.app_name_lambda}-${each.key}"
  timeout       = 30

  environment_variables = {
    MERKELY_HOST = each.value.merkely_host
    MERKELY_API_TOKEN = data.aws_ssm_parameter.merkely_api_token.value
  }

  cloudwatch_logs_retention_in_days = 7

  tags = module.tags.result
}
