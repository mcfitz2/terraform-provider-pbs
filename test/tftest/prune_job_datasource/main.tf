# Main Terraform configuration for prune_job_datasource test
#
# This configuration is shared across all test run blocks.
# Variables are passed in from the .tftest.hcl file.

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

variable "datastore_name" {
  type        = string
  description = "Name of the test datastore"
}

variable "job_id" {
  type        = string
  description = "ID of the test job"
}

# Create a test datastore
resource "pbs_datastore" "test" {
  name = var.datastore_name
  path = "/datastore/${var.datastore_name}"
}

# Create a test prune job
resource "pbs_prune_job" "test" {
  id         = var.job_id
  store      = pbs_datastore.test.name
  schedule   = "daily"
  keep_last  = 7
  keep_daily = 14
  comment    = "Test prune job for data source"
}

# Read the prune job via data source
data "pbs_prune_job" "test" {
  id = pbs_prune_job.test.id
}
