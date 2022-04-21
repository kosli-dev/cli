variable "app_name" {
  type = string
}

variable "env" {
  type = string
}

variable "merkely_env" {
  type = string
}

variable "subnets" {
  type    = list(string)
  default = []
}

variable "ecs_cluster_arn" {
  type = string
}

variable "ecs_events_role_arn" {
  type = string
}

variable "task_sg" {
  type = string
}

variable "execution_role_arn" {
  type = string
}

variable "task_role_arn" {
  type = string
}

variable "cpu_limit" {
  type = number
}

variable "mem_limit" {
  type = number
}

variable "mem_reservation" {
  type = number
}

variable "logs_retention_in_days" {
  type    = number
  default = 14
}

variable "image_tag" {
  type = string
}

variable "secret_prefix" {
  type = string
}

# App variables
variable "merkely_host" {
  type = string
}
