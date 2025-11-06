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

variable "influxdb_udp_host" {
  type        = string
  description = "InfluxDB UDP host"
  default     = "localhost"
}

variable "influxdb_udp_port" {
  type        = number
  description = "InfluxDB UDP port"
  default     = 8089
}

resource "pbs_metrics_server" "test1" {
  name         = "metrics-srv-1-ds-test"
  type         = "influxdb-http"
  url          = "http://${var.influxdb_host}:${var.influxdb_port}"
  organization = "test-org-1"
  bucket       = "test-bucket-1"
  token        = "test-token-1"
  enable       = true
  verify_tls   = false
  comment      = "Integration test metrics server 1"
}

resource "pbs_metrics_server" "test2" {
  name    = "metrics-srv-2-ds-test"
  type    = "influxdb-udp"
  server  = var.influxdb_udp_host
  port    = var.influxdb_udp_port
  enable  = false
  mtu     = 1500
  comment = "Integration test metrics server 2"
}

data "pbs_metrics_servers" "all" {
  depends_on = [
    pbs_metrics_server.test1,
    pbs_metrics_server.test2
  ]
}

output "servers_count" {
  value = length(data.pbs_metrics_servers.all.servers)
}

output "servers" {
  value = data.pbs_metrics_servers.all.servers
}
