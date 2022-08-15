variable "kosli_hosts" {
  type = map(any)
  default = {
    staging = "https://staging.app.kosli.com"
    prod    = "https://app.kosli.com"
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
