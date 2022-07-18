variable "env" {
  type = string
}

variable "name" {
  type    = string
  default = "kosli-reporter"
}

variable "REPORTER_TAG" {
  type    = string
  default = "test"
}

variable "tags" {
  type = map(string)
}

variable "kosli_host" {
  type        = string
  default     = "https://app.kosli.com"
  description = "The Kosli endpoint."
}

variable "ecs_cluster" {
  type        = string
  default     = "merkely"
  description = "The name of the ECS cluster."
}

variable "kosli_user" {
  type        = string
  default     = "compliancedb"
  description = "The Kosli user or organization."
}

variable "create_pull_through_cache_rule" {
  type        = string
  default     = false
  description = "Whether to create pull through cache rule to allow caching repositories in remote public registries in your private Amazon ECR registry."
}

variable "ecr_repository_prefix" {
  type        = string
  default     = "ecr-public"
  description = "The repository name prefix to use when caching images from the source registry."
}

variable "upstream_registry_url" {
  type        = string
  default     = "public.ecr.aws"
  description = "The registry URL of the upstream public registry to use as the source."
}

variable "cloudwatch_logs_retention_in_days" {
  type    = number
  default = 7
}

