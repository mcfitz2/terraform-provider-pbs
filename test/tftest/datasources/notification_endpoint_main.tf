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

resource "pbs_gotify_notification" "test" {
  name    = "gotify-ds-test"
  server  = "https://gotify.example.com"
  token   = "Aabcd1234567890"
  comment = "Integration test gotify endpoint for data source"
  disable = false
}

data "pbs_notification_endpoint" "test" {
  name = pbs_gotify_notification.test.name
  type = "gotify"
}

output "endpoint_name" {
  value = data.pbs_notification_endpoint.test.name
}

output "endpoint_type" {
  value = data.pbs_notification_endpoint.test.type
}

output "endpoint_url" {
  value = data.pbs_notification_endpoint.test.url
}

output "endpoint_comment" {
  value = data.pbs_notification_endpoint.test.comment
}

output "endpoint_disable" {
  value = data.pbs_notification_endpoint.test.disable
}
