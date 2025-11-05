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
  name         = "smtp-matchers-ds-test"
  server       = "smtp.example.com"
  port         = 587
  username     = "test@example.com"
  password     = "secret"
  mailto       = ["admin@example.com"]
  from_address = "pbs@example.com"
}

resource "pbs_notification_matcher" "test1" {
  name           = "matcher1-ds-test"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["error"]
  mode           = "all"
  comment        = "Matcher 1"
}

resource "pbs_notification_matcher" "test2" {
  name           = "matcher2-ds-test"
  targets        = [pbs_smtp_notification.target.name]
  match_severity = ["warning"]
  mode           = "any"
  comment        = "Matcher 2"
}

data "pbs_notification_matchers" "all" {
  depends_on = [
    pbs_notification_matcher.test1,
    pbs_notification_matcher.test2
  ]
}

output "matchers_count" {
  value = length(data.pbs_notification_matchers.all.matchers)
}

output "matchers" {
  value = data.pbs_notification_matchers.all.matchers
}
