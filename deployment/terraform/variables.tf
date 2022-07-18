variable "reporter_apps" {
  type = map(any)
  default = {
    staging = {
      kosli_host      = "https://staging.app.kosli.com"
      cpu_limit       = 100
      mem_limit       = 400
      mem_reservation = 64
    }
    prod = {
      kosli_host      = "https://app.kosli.com"
      cpu_limit       = 100
      mem_limit       = 400
      mem_reservation = 64
    }
  }
}

variable "env" {
  type = string
}

variable "kosli_env" {
  type = string
}

variable "app_name" {
  type    = string
  default = "kosli"
}

variable "ecs_sluster_name" {
  type    = string
  default = "merkely-reporter"
}

variable "create_public_ecr" {
  type    = bool
  default = false
}

variable "REPORTER_TAG" {
  type = string
}
