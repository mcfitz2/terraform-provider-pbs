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
  type = string
}

variable "pbs_insecure" {
  type    = bool
  default = true
}

# Create a verify job
resource "pbs_verify_job" "test" {
  id              = "tftest-verify-ds"
  store           = "datastore1"
  schedule        = "weekly"
  namespace       = "prod"
  ignore_verified = true
  outdated_after  = 30
  comment         = "Test verify job for data source"
}

# Read it via data source
data "pbs_verify_job" "test" {
  id = pbs_verify_job.test.id
}

output "resource_id" {
  value = pbs_verify_job.test.id
}

output "datasource_id" {
  value = data.pbs_verify_job.test.id
}

output "datasource_schedule" {
  value = data.pbs_verify_job.test.schedule
}
