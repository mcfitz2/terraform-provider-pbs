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

variable "name" {
  type = string
}

variable "server" {
  type = string
}

variable "port" {
  type    = number
  default = 587
}

variable "mode" {
  type    = string
  default = "insecure"
}

variable "username" {
  type = string
}

variable "password_smtp" {
  type      = string
  sensitive = true
}

variable "mailto" {
  type = list(string)
}

variable "from_address" {
  type = string
}

variable "author" {
  type    = string
  default = "PBS Admin"
}

variable "comment" {
  type = string
}

variable "disable" {
  type    = bool
  default = false
}

resource "pbs_smtp_notification" "test" {
  name         = var.name
  server       = var.server
  port         = var.port
  mode         = var.mode
  username     = var.username
  password     = var.password_smtp
  mailto       = var.mailto
  from_address = var.from_address
  author       = var.author
  comment      = var.comment
  disable      = var.disable
}

output "name" {
  value = pbs_smtp_notification.test.name
}

output "server" {
  value = pbs_smtp_notification.test.server
}
