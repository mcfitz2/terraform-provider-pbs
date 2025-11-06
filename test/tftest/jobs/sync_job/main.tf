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
  type = string
}

variable "store" {
  type    = string
  default = "datastore1"
}

variable "remote" {
  type    = string
  default = "remote1"
}

variable "remote_store" {
  type = string
}

variable "schedule" {
  type = string
}

variable "remote_namespace" {
  type    = string
  default = null
}

variable "namespace" {
  type    = string
  default = null
}

variable "remove_vanished" {
  type    = bool
  default = null
}

variable "resync_corrupt" {
  type    = bool
  default = null
}

variable "verified_only" {
  type    = bool
  default = null
}

variable "run_on_mount" {
  type    = bool
  default = null
}

variable "transfer_last" {
  type    = number
  default = null
}

variable "rate_in" {
  type    = string
  default = null
}

variable "rate_out" {
  type    = string
  default = null
}

variable "burst_in" {
  type    = string
  default = null
}

variable "burst_out" {
  type    = string
  default = null
}

variable "group_filter" {
  type    = list(string)
  default = null
}

variable "comment" {
  type = string
}

resource "pbs_sync_job" "test" {
  id               = var.job_id
  store            = var.store
  remote           = var.remote
  remote_store     = var.remote_store
  remote_namespace = var.remote_namespace
  namespace        = var.namespace
  schedule         = var.schedule
  remove_vanished  = var.remove_vanished
  resync_corrupt   = var.resync_corrupt
  verified_only    = var.verified_only
  run_on_mount     = var.run_on_mount
  transfer_last    = var.transfer_last
  rate_in          = var.rate_in
  rate_out         = var.rate_out
  burst_in         = var.burst_in
  burst_out        = var.burst_out
  group_filter     = var.group_filter
  comment          = var.comment
}

output "job_id" {
  value = pbs_sync_job.test.id
}

output "schedule" {
  value = pbs_sync_job.test.schedule
}

output "comment" {
  value = pbs_sync_job.test.comment
}
