module "aws_ecr_repository_app" {
  source                  = "s3::https://s3-eu-central-1.amazonaws.com/terraform-modules-dacef8339fbd41ce31c346f854a85d0c74f7c4e8/terraform-modules.zip//ecr/v1"
  ecr_repository_name     = var.app_name
  ecr_replication_targets = var.ecr_replication_targets
  ecr_replication_origin  = var.ecr_replication_origin
  tags                    = module.tags.result
}
