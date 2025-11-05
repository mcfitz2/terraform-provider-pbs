terraform {
  required_version = ">= 1.6.0"
  required_providers {
    pbs = {
      source  = "registry.terraform.io/micah/pbs"
      version = "0.2.0"
    }
  }
}

variable "pbs_endpoint" {
  type        = string
  description = "PBS server endpoint"
}

variable "pbs_username" {
  type        = string
  description = "PBS username"
}

variable "pbs_password" {
  type        = string
  description = "PBS password"
  sensitive   = true
}

variable "influxdb_host" {
  type        = string
  description = "InfluxDB host"
  default     = "localhost"
}

variable "influxdb_port" {
  type        = number
  description = "InfluxDB HTTP port"
  default     = 8086
}

resource "pbs_metrics_server" "test" {
  name         = "metrics-server-ds-test"
  type         = "influxdb-http"
  url          = "http://${var.influxdb_host}:${var.influxdb_port}"
  organization = "test-org"
  bucket       = "test-bucket"
  token        = "test-token-value"
  enable       = true
  verify_tls   = false
  comment      = "Integration test metrics server"
}

data "pbs_metrics_server" "test" {
  name = pbs_metrics_server.test.name
  type = pbs_metrics_server.test.type
}

output "metrics_server_name" {
  value = data.pbs_metrics_server.test.name
}

output "metrics_server_type" {
  value = data.pbs_metrics_server.test.type
}

output "metrics_server_url" {
  value = data.pbs_metrics_server.test.url
}

output "metrics_server_organization" {
  value = data.pbs_metrics_server.test.organization
}

output "metrics_server_bucket" {
  value = data.pbs_metrics_server.test.bucket
}

output "metrics_server_comment" {
  value = data.pbs_metrics_server.test.comment
}
