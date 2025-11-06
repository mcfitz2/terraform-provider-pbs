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

# Create multiple verify jobs
resource "pbs_verify_job" "test1" {
  id       = "tftest-verify-list-1"
  store    = "datastore1"
  schedule = "daily"
  comment  = "First verify job"
}

resource "pbs_verify_job" "test2" {
  id       = "tftest-verify-list-2"
  store    = "datastore1"
  schedule = "weekly"
  comment  = "Second verify job"
}

# List all verify jobs
data "pbs_verify_jobs" "all" {
  depends_on = [
    pbs_verify_job.test1,
    pbs_verify_job.test2
  ]
}

output "job_count" {
  value = length(data.pbs_verify_jobs.all.jobs)
}

output "job_ids" {
  value = [for j in data.pbs_verify_jobs.all.jobs : j.id]
}
