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

# Create multiple sync jobs
resource "pbs_sync_job" "test1" {
  id           = "tftest-sync-list-1"
  store        = "datastore1"
  remote       = "remote1"
  remote_store = "backup"
  schedule     = "daily"
  comment      = "First sync job"
}

resource "pbs_sync_job" "test2" {
  id           = "tftest-sync-list-2"
  store        = "datastore1"
  remote       = "remote1"
  remote_store = "backup"
  schedule     = "weekly"
  comment      = "Second sync job"
}

# List all sync jobs
data "pbs_sync_jobs" "all" {
  depends_on = [
    pbs_sync_job.test1,
    pbs_sync_job.test2
  ]
}

output "job_count" {
  value = length(data.pbs_sync_jobs.all.jobs)
}

output "job_ids" {
  value = [for j in data.pbs_sync_jobs.all.jobs : j.id]
}
