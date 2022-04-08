aws_region   = "eu-central-1"
env          = "staging"
merkely_env  = "staging-aws"
MERKELY_HOST = "https://staging.app.merkely.com"
mem_limit    = 64
cpu_limit    = 100
# Allow to replicate app docker images to these accounts
ecr_replication_targets = [
  {
    "account_id" = "358426185766",
    "region"     = "eu-central-1"
  }
]

