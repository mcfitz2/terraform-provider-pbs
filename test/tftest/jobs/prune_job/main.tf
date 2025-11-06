terraform {
  required_providers {
    pbs = {
      source = "registry.terraform.io/micah/pbs"
    }
  }
}

provider "pbs" {
  endpoint = var.pbs_endpoint
  insecure = var.pbs_insecure
}

variable "pbs_endpoint" {
  type        = string
  description = "PBS endpoint URL"
}

variable "pbs_insecure" {
  type        = bool
  default     = true
  description = "Skip TLS verification"
}

variable "job_id" {
  type        = string
  description = "Unique job identifier"
}

variable "store" {
  type    = string
  default = "datastore1"
}

variable "schedule" {
  type = string
}

variable "keep_last" {
  type = number
}

variable "keep_daily" {
  type = number
}

variable "keep_weekly" {
  type = number
}

variable "keep_monthly" {
  type = number
}

variable "keep_yearly" {
  type = number
}

variable "max_depth" {
  type = number
}

variable "comment" {
  type = string
}

variable "namespace" {
  type    = string
  default = null
}

# Prune job resource
resource "pbs_prune_job" "test" {
  id           = var.job_id
  store        = var.store
  schedule     = var.schedule
  keep_last    = var.keep_last
  keep_daily   = var.keep_daily
  keep_weekly  = var.keep_weekly
  keep_monthly = var.keep_monthly
  keep_yearly  = var.keep_yearly
  max_depth    = var.max_depth
  comment      = var.comment
  namespace    = var.namespace
}

output "job_id" {
  value = pbs_prune_job.test.id
}

output "schedule" {
  value = pbs_prune_job.test.schedule
}

output "keep_last" {
  value = pbs_prune_job.test.keep_last
}

output "comment" {
  value = pbs_prune_job.test.comment
}
