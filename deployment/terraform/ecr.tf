# Create ECS repository to store docker images
# https://docs.aws.amazon.com/AmazonECR/latest/userguide/what-is-ecr.html 
resource "aws_ecr_repository" "this" {
  name = var.app_name
  image_scanning_configuration {
    scan_on_push = "true"
  }
  encryption_configuration {
    encryption_type = "AES256"
  }
  tags = module.tags.result
}

# https://docs.aws.amazon.com/AmazonECR/latest/userguide/LifecyclePolicies.html
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
