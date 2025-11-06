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

variable "datastore_name" {
  type = string
}

# Create a datastore resource
resource "pbs_datastore" "test" {
  name        = var.datastore_name
  path        = "/datastore/${var.datastore_name}"
  comment     = "Test datastore for data source"
  gc_schedule = "daily"
}

# Read it back via data source
data "pbs_datastore" "test" {
  name = pbs_datastore.test.name
}

output "resource_name" {
  value = pbs_datastore.test.name
}

output "datasource_name" {
  value = data.pbs_datastore.test.name
}

output "datasource_path" {
  value = data.pbs_datastore.test.path
}

output "datasource_comment" {
  value = data.pbs_datastore.test.comment
}
