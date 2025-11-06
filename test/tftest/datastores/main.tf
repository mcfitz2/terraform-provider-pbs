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
  description = "PBS endpoint URL"
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

variable "test_name" {
  type        = string
  description = "Unique test identifier"
}

# Directory datastore resource
resource "pbs_datastore" "test_directory" {
  name        = var.test_name
  path        = "/datastore/${var.test_name}"
  comment     = var.comment
  gc_schedule = var.gc_schedule
}

variable "comment" {
  type    = string
  default = "Test directory datastore"
}

variable "gc_schedule" {
  type    = string
  default = "daily"
}

output "datastore_name" {
  value = pbs_datastore.test_directory.name
}

output "datastore_path" {
  value = pbs_datastore.test_directory.path
}

output "datastore_comment" {
  value = pbs_datastore.test_directory.comment
}

output "datastore_gc_schedule" {
  value = pbs_datastore.test_directory.gc_schedule
}
