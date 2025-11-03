# Main Terraform configuration for prune_jobs_datasource test
#
# This configuration creates multiple datastores and prune jobs,
# then tests listing them with and without filters.

terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    pbs = {
      source = "registry.terraform.io/micah/pbs"
      version = "1.0.0"
    }
  }
}

# Variables that will be set by the test file
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

variable "datastore1_name" {
  type        = string
  description = "Name of the first test datastore"
}

variable "datastore2_name" {
  type        = string
  description = "Name of the second test datastore"
}

variable "job1_id" {
  type        = string
  description = "ID of the first test job"
}

variable "job2_id" {
  type        = string
  description = "ID of the second test job"
}

# Create first test datastore
resource "pbs_datastore" "test1" {
  name = var.datastore1_name
  path = "/datastore/${var.datastore1_name}"
}

# Create second test datastore
resource "pbs_datastore" "test2" {
  name = var.datastore2_name
  path = "/datastore/${var.datastore2_name}"
}

# Create first prune job
resource "pbs_prune_job" "test1" {
  id        = var.job1_id
  store     = pbs_datastore.test1.name
  schedule  = "daily"
  keep_last = 5
  comment   = "Test prune job 1 for listing"
}

# Create second prune job
resource "pbs_prune_job" "test2" {
  id        = var.job2_id
  store     = pbs_datastore.test2.name
  schedule  = "weekly"
  keep_last = 10
  comment   = "Test prune job 2 for listing"
}

# List all prune jobs (unfiltered)
data "pbs_prune_jobs" "all" {
  depends_on = [
    pbs_prune_job.test1,
    pbs_prune_job.test2
  ]
}

# List prune jobs filtered by store
data "pbs_prune_jobs" "filtered" {
  store = pbs_datastore.test1.name
  
  depends_on = [
    pbs_prune_job.test1,
    pbs_prune_job.test2
  ]
}
