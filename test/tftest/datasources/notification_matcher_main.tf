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

resource "pbs_smtp_notification" "target" {
  name         = "smtp-matcher-ds-test"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = ["admin@example.com"]
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test" {
  name           = "matcher-ds-test"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["error", "warning"]
  mode           = "all"
  comment        = "Integration test matcher for data source"
}

data "pbs_notification_matcher" "test" {
  name = pbs_notification_matcher.test.name
}

output "matcher_name" {
  value = data.pbs_notification_matcher.test.name
}

output "matcher_targets" {
  value = data.pbs_notification_matcher.test.targets
}

output "matcher_severity" {
  value = data.pbs_notification_matcher.test.match_severity
}

output "matcher_mode" {
  value = data.pbs_notification_matcher.test.mode
}

output "matcher_comment" {
  value = data.pbs_notification_matcher.test.comment
}
