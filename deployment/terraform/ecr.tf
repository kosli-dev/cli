resource "aws_ecrpublic_repository" "this" {
  count    = var.create_public_ecr ? 1 : 0
  provider = aws.us-east-1

  repository_name = "${var.app_name}-reporter"

  catalog_data {
    about_text = "${var.app_name}-reporter"
  }
}

resource "aws_ecrpublic_repository_policy" "this" {
  count           = var.create_public_ecr ? 1 : 0
  provider        = aws.us-east-1
  repository_name = aws_ecrpublic_repository.this[0].repository_name
  policy          = data.aws_iam_policy_document.ecr_public_write[0].json
}

data "aws_iam_policy_document" "ecr_public_write" {
  count = var.create_public_ecr ? 1 : 0
  statement {
    sid    = "ECRPublicWrite"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = setunion(data.aws_iam_roles.roles_admin[0].arns, [data.aws_ssm_parameter.oidc_role_arn[0].value])
    }
    actions = [
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "ecr:BatchCheckLayerAvailability",
      "ecr:PutImage",
      "ecr:InitiateLayerUpload",
      "ecr:UploadLayerPart",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeRepositories",
      "ecr:GetRepositoryPolicy",
      "ecr:ListImages",
      "ecr:DeleteRepository",
      "ecr:BatchDeleteImage",
      "ecr:SetRepositoryPolicy",
      "ecr:DeleteRepositoryPolicy"
    ]
  }
}

data "aws_iam_roles" "roles_admin" {
  count       = var.create_public_ecr ? 1 : 0
  name_regex  = "AWSReservedSSO_Admin.*"
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
