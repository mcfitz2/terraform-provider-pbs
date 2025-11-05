terraform {
  required_providers {
    pbs = {
      source = "registry.terraform.io/micah/pbs"
    }
  }
}

provider "pbs" {
  endpoint = var.pbs_endpoint
  insecure = var.pbs_insecure
}

variable "pbs_endpoint" {
  type        = string
  description = "PBS endpoint URL"
}

variable "pbs_insecure" {
  type        = bool
  default     = true
  description = "Skip TLS verification"
}

variable "remote_name" {
  type = string
}

variable "host" {
  type = string
}

variable "port" {
  type    = number
  default = null
}

variable "auth_id" {
  type = string
}

variable "password" {
  type      = string
  sensitive = true
}

variable "fingerprint" {
  type    = string
  default = null
}

variable "comment" {
  type    = string
  default = null
}

resource "pbs_remote" "test" {
  name        = var.remote_name
  host        = var.host
  port        = var.port
  auth_id     = var.auth_id
  password    = var.password
  fingerprint = var.fingerprint
  comment     = var.comment
}

output "remote_name" {
  value = pbs_remote.test.name
}

output "host" {
  value = pbs_remote.test.host
}

output "comment" {
  value = pbs_remote.test.comment
}
