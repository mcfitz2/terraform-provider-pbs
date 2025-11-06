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

# Gotify variables
variable "gotify_name" {
  type = string
}

variable "gotify_server" {
  type = string
}

variable "gotify_token" {
  type      = string
  sensitive = true
}

# Sendmail variables
variable "sendmail_name" {
  type = string
}

variable "sendmail_mailto" {
  type = list(string)
}

variable "sendmail_from" {
  type = string
}

# Webhook variables
variable "webhook_name" {
  type = string
}

variable "webhook_url" {
  type = string
}

variable "webhook_method" {
  type    = string
  default = "post"
}

# Matcher variables
variable "matcher_name" {
  type = string
}

variable "matcher_mode" {
  type    = string
  default = "all"
}

variable "match_severity" {
  type    = list(string)
  default = ["error", "warning"]
}

variable "match_calendar" {
  type    = list(string)
  default = null
}

variable "invert_match" {
  type    = bool
  default = false
}

# Gotify notification
resource "pbs_gotify_notification" "test" {
  name    = var.gotify_name
  server  = var.gotify_server
  token   = var.gotify_token
  comment = "Test Gotify notification"
  disable = false
}

# Sendmail notification
resource "pbs_sendmail_notification" "test" {
  name         = var.sendmail_name
  mailto       = var.sendmail_mailto
  from_address = var.sendmail_from
  author       = "PBS System"
  comment      = "Test Sendmail notification"
  disable      = false
}

# Webhook notification
resource "pbs_webhook_notification" "test" {
  name    = var.webhook_name
  url     = var.webhook_url
  method  = var.webhook_method
  comment = "Test Webhook notification"
  disable = false
}

# Notification matcher
resource "pbs_notification_matcher" "test" {
  name           = var.matcher_name
  targets        = [pbs_gotify_notification.test.name]
  match_severity = var.match_severity
  match_calendar = var.match_calendar
  mode           = var.matcher_mode
  invert_match   = var.invert_match
  comment        = "Test notification matcher"
  disable        = false
}

output "gotify_name" {
  value = pbs_gotify_notification.test.name
}

output "sendmail_name" {
  value = pbs_sendmail_notification.test.name
}

output "webhook_name" {
  value = pbs_webhook_notification.test.name
}

output "matcher_name" {
  value = pbs_notification_matcher.test.name
}

output "matcher_mode" {
  value = pbs_notification_matcher.test.mode
}
