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

resource "pbs_gotify_notification" "test1" {
  name    = "gotify-plural-ds-test"
  server  = "https://gotify.example.com"
  token   = "Aabcd1234567890"
  comment = "Gotify endpoint 1"
  disable = false
}

resource "pbs_smtp_notification" "test2" {
  name         = "smtp-plural-ds-test"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = ["admin@example.com"]
  from_address = "pbs@example.com"
  comment      = "SMTP endpoint 1"
  disable      = false
}

data "pbs_notification_endpoints" "all" {
  depends_on = [
    pbs_gotify_notification.test1,
    pbs_smtp_notification.test2
  ]
}

output "endpoints_count" {
  value = length(data.pbs_notification_endpoints.all.endpoints)
}

output "endpoints" {
  value = data.pbs_notification_endpoints.all.endpoints
}
