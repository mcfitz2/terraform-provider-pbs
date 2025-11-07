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

variable "influxdb_udp_host" {
  type    = string
  default = "localhost"
}

variable "influxdb_udp_port" {
  type    = number
  default = 8089
}

variable "mtu" {
  type    = number
  default = null
}

variable "comment" {
  type = string
}

variable "enable" {
  type    = bool
  default = true
}

resource "pbs_metrics_server" "test" {
  name     = var.server_name
  type     = "influxdb-udp"
  server   = var.influxdb_udp_host
  port     = var.influxdb_udp_port
  protocol = "udp"
  mtu      = var.mtu
  enable   = var.enable
  comment  = var.comment
}

output "server_name" {
  value = pbs_metrics_server.test.name
}

output "server_type" {
  value = pbs_metrics_server.test.type
}
