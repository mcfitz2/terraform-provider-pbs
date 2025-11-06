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

variable "test_id" {
  type        = string
  description = "Unique test run identifier to avoid name conflicts"
  default     = "local"
}

provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

variable "datastore_name" {
  type = string
}

variable "datastore_path" {
  type = string
}

variable "comment" {
  type    = string
  default = "Test directory datastore"
}

variable "gc_schedule" {
  type    = string
  default = null
}

# Create directory-backed datastore
resource "pbs_datastore" "dir_test" {
  name        = var.datastore_name
  path        = var.datastore_path
  comment     = var.comment
  gc_schedule = var.gc_schedule
}

output "datastore_name" {
  value = pbs_datastore.dir_test.name
}

output "datastore_path" {
  value = pbs_datastore.dir_test.path
}

output "datastore_comment" {
  value = pbs_datastore.dir_test.comment
}

output "datastore_gc_schedule" {
  value = pbs_datastore.dir_test.gc_schedule
}
