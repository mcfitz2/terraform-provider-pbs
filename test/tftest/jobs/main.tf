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

variable "test_id" {
  type        = string
  description = "Unique test run identifier"
  default     = "local"
}

# Common job variables
variable "job_id" {
  type        = string
  description = "Unique job identifier"
}

variable "store" {
  type        = string
  description = "Datastore name"
  default     = "datastore1"
}

variable "schedule" {
  type        = string
  description = "Job schedule"
}

variable "namespace" {
  type        = string
  description = "Namespace"
  default     = null
}

variable "max_depth" {
  type        = number
  description = "Maximum namespace depth"
  default     = null
}

variable "comment" {
  type        = string
  description = "Job comment"
  default     = null
}

# Prune job specific variables
variable "keep_last" {
  type        = number
  description = "Keep last N backups"
  default     = null
}

variable "keep_daily" {
  type        = number
  description = "Keep N daily backups"
  default     = null
}

variable "keep_weekly" {
  type        = number
  description = "Keep N weekly backups"
  default     = null
}

variable "keep_monthly" {
  type        = number
  description = "Keep N monthly backups"
  default     = null
}

variable "keep_yearly" {
  type        = number
  description = "Keep N yearly backups"
  default     = null
}

# Sync job specific variables
variable "remote" {
  type        = string
  description = "Remote name"
  default     = null
}

variable "remote_store" {
  type        = string
  description = "Remote store name"
  default     = null
}

variable "remote_namespace" {
  type        = string
  description = "Remote namespace"
  default     = null
}

variable "remove_vanished" {
  type        = bool
  description = "Remove vanished backups"
  default     = null
}

variable "resync_corrupt" {
  type        = bool
  description = "Resync corrupt backups"
  default     = null
}

variable "rate_in" {
  type        = string
  description = "Incoming rate limit"
  default     = null
}

variable "rate_out" {
  type        = string
  description = "Outgoing rate limit"
  default     = null
}

variable "burst_in" {
  type        = string
  description = "Incoming burst limit"
  default     = null
}

variable "burst_out" {
  type        = string
  description = "Outgoing burst limit"
  default     = null
}

variable "group_filter" {
  type        = list(string)
  description = "Group filters"
  default     = null
}

variable "verified_only" {
  type        = bool
  description = "Only sync verified backups"
  default     = null
}

variable "run_on_mount" {
  type        = bool
  description = "Run job when datastore is mounted"
  default     = null
}

variable "transfer_last" {
  type        = number
  description = "Only sync backups from last N seconds"
  default     = null
}

# Verify job specific variables
variable "ignore_verified" {
  type        = bool
  description = "Ignore already verified backups"
  default     = null
}

variable "outdated_after" {
  type        = number
  description = "Days after which backups are considered outdated"
  default     = null
}

# Job resources - only create the one specified by job_type
variable "job_type" {
  type        = string
  description = "Type of job: prune, sync, or verify"
  validation {
    condition     = contains(["prune", "sync", "verify"], var.job_type)
    error_message = "job_type must be one of: prune, sync, verify"
  }
}

resource "pbs_prune_job" "test" {
  count = var.job_type == "prune" ? 1 : 0

  id           = var.job_id
  store        = var.store
  schedule     = var.schedule
  keep_last    = var.keep_last
  keep_daily   = var.keep_daily
  keep_weekly  = var.keep_weekly
  keep_monthly = var.keep_monthly
  keep_yearly  = var.keep_yearly
  namespace    = var.namespace
  max_depth    = var.max_depth
  comment      = var.comment
}

resource "pbs_sync_job" "test" {
  count = var.job_type == "sync" ? 1 : 0

  id               = var.job_id
  store            = var.store
  remote           = var.remote
  remote_store     = var.remote_store
  remote_namespace = var.remote_namespace
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
  namespace        = var.namespace
  max_depth        = var.max_depth
  comment          = var.comment
}

resource "pbs_verify_job" "test" {
  count = var.job_type == "verify" ? 1 : 0

  id              = var.job_id
  store           = var.store
  schedule        = var.schedule
  namespace       = var.namespace
  ignore_verified = var.ignore_verified
  outdated_after  = var.outdated_after
  max_depth       = var.max_depth
  comment         = var.comment
}

# Outputs - use try() to handle when resource doesn't exist
output "job_id" {
  value = try(
    pbs_prune_job.test[0].id,
    pbs_sync_job.test[0].id,
    pbs_verify_job.test[0].id,
    null
  )
}

output "schedule" {
  value = try(
    pbs_prune_job.test[0].schedule,
    pbs_sync_job.test[0].schedule,
    pbs_verify_job.test[0].schedule,
    null
  )
}

output "comment" {
  value = try(
    pbs_prune_job.test[0].comment,
    pbs_sync_job.test[0].comment,
    pbs_verify_job.test[0].comment,
    null
  )
}
