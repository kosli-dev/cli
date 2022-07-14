variable "reporter_apps" {
  type = map(any)
  default = {
    staging = {
      merkely_host    = "https://staging.app.merkely.com"
      cpu_limit       = 100
      mem_limit       = 400
      mem_reservation = 64
    }
    prod = {
      merkely_host    = "https://app.merkely.com"
      cpu_limit       = 100
      mem_limit       = 400
      mem_reservation = 64
    }
  }
}

variable "env" {
  type = string
}

variable "merkely_env" {
  type = string
}

variable "app_name" {
  type    = string
  default = "merkely-cli"
  #default = "kosli-reporter"
}

variable "app_name_lambda" {
  type    = string
  default = "kosli-reporter"
}

variable "ecs_sluster_name" {
  type    = string
  default = "merkely-reporter"
}

variable "create_public_ecr" {
  type    = bool
  default = false
}

variable "IMAGE_TAG" {
  type = string
}

variable "REPORTER_TAG" {
  type    = string
  default = "test"
}
