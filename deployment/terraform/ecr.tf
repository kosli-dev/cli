resource "aws_ecrpublic_repository" "this" {
  count   = var.create_public_ecr ? 1 : 0
  provider = aws.us-east-1

  repository_name = var.app_name_lambda

  catalog_data {
    about_text        = var.app_name_lambda
    #architectures     = ["ARM"]
    #description       = "Description"
    #logo_image_blob   = filebase64(image.png)
    #operating_systems = ["Linux"]
    #usage_text        = "Usage Text"
  }
}

resource "aws_ecrpublic_repository_policy" "this" {
  count   = var.create_public_ecr ? 1 : 0
  provider = aws.us-east-1
  repository_name = aws_ecrpublic_repository.this[0].repository_name
  policy          = data.aws_iam_policy_document.ecr_public_write.json
}

data "aws_iam_policy_document" "ecr_public_write" {
  statement {
    sid    = "ECRPublicWrite"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = local.principals_identifiers
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

locals {
  principals_identifiers = setunion(data.aws_iam_roles.roles_admin.arns, [data.aws_ssm_parameter.oidc_role_arn.value])
  
}

data "aws_iam_roles" "roles_admin" {
  name_regex  = "AWSReservedSSO_Admin.*"
  path_prefix = "/aws-reserved/sso.amazonaws.com/"
}
