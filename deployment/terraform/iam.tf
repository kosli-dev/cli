# IAM app policies
# https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_execution_IAM_role.html
# https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-iam-roles.html
data "aws_iam_policy_document" "assume" {
  statement {
    sid     = "Assume"
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "secret_get_allow" {
  statement {
    sid    = "SecretGet"
    effect = "Allow"
    actions = [
      "secretsmanager:GetSecretValue"
    ]
    resources = [
      "${local.secret_prefix}*"
    ]
  }
}

resource "aws_iam_policy" "secret_get_allow" {
  name        = "secret_get_allow_${var.app_name}"
  description = "Policy to allow ECS task to read secrets from AWS secret manager"
  path        = "/ecs/"
  policy      = data.aws_iam_policy_document.secret_get_allow.json
}

data "aws_iam_policy_document" "ecs_list_allow" {
  statement {
    sid    = "ECSList"
    effect = "Allow"
    actions = [
      "ecs:ListTasks",
      "ecs:DescribeTasks"
    ]
    resources = [
      "*"
    ]
  }
}

resource "aws_iam_policy" "ecs_list_allow" {
  name        = "ecs_list_allow_${var.app_name}"
  description = "Policy to allow ECS task to list cluster tasks"
  path        = "/ecs/"
  policy      = data.aws_iam_policy_document.ecs_list_allow.json
}

resource "aws_iam_role" "exec" {
  name               = "${var.app_name}-ecs-task-execution-role"
  description        = "${var.app_name} ECS task execution role"
  path               = "/ecs/"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

resource "aws_iam_role" "task" {
  name               = "${var.app_name}-ecs-task-role"
  description        = "${var.app_name} ECS task role"
  path               = "/ecs/"
  assume_role_policy = data.aws_iam_policy_document.assume.json
}

resource "aws_iam_role_policy_attachment" "exec" {
  role       = aws_iam_role.exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy_attachment" "exec_secret_get_allow" {
  role       = aws_iam_role.exec.name
  policy_arn = aws_iam_policy.secret_get_allow.arn
}

resource "aws_iam_role_policy_attachment" "task_ecs_list_allow" {
  role       = aws_iam_role.task.name
  policy_arn = aws_iam_policy.ecs_list_allow.arn
}
