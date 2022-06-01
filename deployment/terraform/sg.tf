# Allow all traffic from load balancer to app service
# https://docs.aws.amazon.com/vpc/latest/userguide/VPC_SecurityGroups.html
module "sg" {
  source       = "terraform-aws-modules/security-group/aws"
  version      = "4.9.0"
  name         = "${var.app_name}-task"
  description  = "ECS task merkely-cli"
  vpc_id       = data.aws_ssm_parameter.vpc_id.value
  egress_rules = ["all-all"]
  tags         = module.tags.result
}
