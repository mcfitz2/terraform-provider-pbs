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

variable "schedule" {
  type = string
}

variable "namespace" {
  type    = string
  default = null
}

variable "ignore_verified" {
  type    = bool
  default = null
}

variable "outdated_after" {
  type    = number
  default = null
}

variable "max_depth" {
  type    = number
  default = null
}

variable "comment" {
  type = string
}

resource "pbs_verify_job" "test" {
  id              = var.job_id
  store           = var.store
  schedule        = var.schedule
  namespace       = var.namespace
  ignore_verified = var.ignore_verified
  outdated_after  = var.outdated_after
  max_depth       = var.max_depth
  comment         = var.comment
}

output "job_id" {
  value = pbs_verify_job.test.id
}

output "schedule" {
  value = pbs_verify_job.test.schedule
}

output "comment" {
  value = pbs_verify_job.test.comment
}
