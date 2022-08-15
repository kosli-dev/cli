variable "env" {
  type = string
}

variable "name" {
  type = string
}

variable "kosli_cli_version" {
  type = string
}

variable "tags" {
  type = map(string)
}

variable "kosli_host" {
  type        = string
  default     = "https://app.kosli.com"
  description = "The Kosli endpoint."
}

variable "reporter_releases_host" {
  type    = string
  default = "https://reporter-releases.kosli.com"
}

variable "ecs_cluster" {
  type        = string
  default     = "app"
  description = "The name of the ECS cluster where app is running."
}

variable "kosli_user" {
  type        = string
  default     = "cyber-dojo"
  description = "The Kosli user or organization."
}

variable "cloudwatch_logs_retention_in_days" {
  type    = number
  default = 7
}
