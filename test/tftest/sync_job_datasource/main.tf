# Main Terraform configuration for sync_job_datasource test
#
# This configuration is shared across all test run blocks.
# Variables are passed in from the .tftest.hcl file.

terraform {
  required_version = ">= 1.6.0"
  
  required_providers {
    pbs = {
      source = "registry.terraform.io/micah/pbs"
      version = "0.2.0"
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

variable "remote_name" {
  type        = string
  description = "Name of the test remote"
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

# Create a test remote
resource "pbs_remote" "test" {
  name        = var.remote_name
  host        = "remote.example.com"
  auth_id     = "test@pbs"
  password    = "testpassword"
  fingerprint = "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff:00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff"
}

# Create a test sync job
resource "pbs_sync_job" "test" {
  id           = var.job_id
  store        = pbs_datastore.test.name
  remote       = pbs_remote.test.name
  remote_store = "backup"
  schedule     = "hourly"
  comment      = "Test sync job for data source"
}

# Read the sync job via data source
data "pbs_sync_job" "test" {
  id = pbs_sync_job.test.id
}
