env         = "staging"
merkely_env = "staging-aws"

# Allow to replicate app docker images to these accounts
ecr_replication_targets = [
  {
    "account_id" = "358426185766",
    "region"     = "eu-central-1"
  }
]
