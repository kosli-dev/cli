variable "merkely_env" {
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

variable "merkely_host" {
  type = string
}
