terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "1.0.0"
    }
  }
}

variable "pbs_endpoint" {
  type = string
}

variable "pbs_username" {
  type = string
}

variable "pbs_password" {
  type      = string
  sensitive = true
}

variable "server_name" {
  type = string
}

variable "influxdb_host" {
  type    = string
  default = "localhost"
}

variable "influxdb_port" {
  type    = number
  default = 8086
}

variable "organization" {
  type = string
}

variable "bucket" {
  type = string
}

variable "token" {
  type      = string
  sensitive = true
}

variable "comment" {
  type = string
}

resource "pbs_metrics_server" "test" {
  name         = var.server_name
  type         = "influxdb-http"
  url          = "http://${var.influxdb_host}:${var.influxdb_port}"
  organization = var.organization
  bucket       = var.bucket
  token        = var.token
  enable       = true
  comment      = var.comment
}

output "server_name" {
  value = pbs_metrics_server.test.name
}

output "server_type" {
  value = pbs_metrics_server.test.type
}

output "url" {
  value = pbs_metrics_server.test.url
}
