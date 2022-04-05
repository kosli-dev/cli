variable "aws_region" {
  type = string
}

variable "env" {
  type = string
}

variable "app_name" {
  type    = string
  default = "merkely-cli"
}

variable "cpu_limit" {
  type = number
}

variable "mem_limit" {
  type = number
}

variable "logs_retention_in_days" {
  type    = number
  default = 14
}

variable "TAGGED_IMAGE" {
  type = string
}

# App variables
variable "MERKELY_HOST" {
  type = string
}

