# Main Terraform configuration for datastores_datasource test
#
# This configuration creates multiple datastores and tests listing them.

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

variable "datastore1_name" {
  type        = string
  description = "Name of the first test datastore"
}

variable "datastore2_name" {
  type        = string
  description = "Name of the second test datastore"
}

# Create first test datastore
resource "pbs_datastore" "test1" {
  name    = var.datastore1_name
  path    = "/datastore/${var.datastore1_name}"
  comment = "First test datastore"
}

# Create second test datastore
resource "pbs_datastore" "test2" {
  name    = var.datastore2_name
  path    = "/datastore/${var.datastore2_name}"
  comment = "Second test datastore"
}

# List all datastores
data "pbs_datastores" "all" {
  depends_on = [
    pbs_datastore.test1,
    pbs_datastore.test2
  ]
}
