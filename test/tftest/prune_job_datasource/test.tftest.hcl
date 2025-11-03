# Test for reading a single prune job via data source
#
# This test verifies that:
# 1. A datastore can be created
# 2. A prune job can be created that references the datastore
# 3. The prune job data source can read the job
# 4. All attributes match between the resource and data source
#
# This was converted from a Go tfexec test that was flaky in CI due to
# timing issues. The native Terraform test framework eliminates those issues
# by using the same execution path as manual `terraform apply`.

variables {
  # Generate unique names with timestamp to avoid conflicts
  datastore_name = "ds-prune-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  job_id = "prune-ds-${formatdate("YYYYMMDDhhmmss", timestamp())}"
}

# Provider configuration - can be overridden via environment variables
provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

# Run block 1: Create datastore and prune job, then read via data source
run "create_and_read" {
  command = apply

  # Create the configuration inline
  variables {
    datastore_name = var.datastore_name
    job_id = var.job_id
  }

  # Verify datastore was created
  assert {
    condition     = pbs_datastore.test.name == var.datastore_name
    error_message = "Datastore name does not match expected value"
  }

  assert {
    condition     = pbs_datastore.test.path == "/datastore/${var.datastore_name}"
    error_message = "Datastore path does not match expected value"
  }

  # Verify prune job was created
  assert {
    condition     = pbs_prune_job.test.id == var.job_id
    error_message = "Prune job ID does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test.store == var.datastore_name
    error_message = "Prune job store does not match datastore name"
  }

  assert {
    condition     = pbs_prune_job.test.schedule == "daily"
    error_message = "Prune job schedule does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test.keep_last == 7
    error_message = "Prune job keep_last does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test.keep_daily == 14
    error_message = "Prune job keep_daily does not match expected value"
  }

  # Verify data source reads the job correctly
  assert {
    condition     = data.pbs_prune_job.test.id == pbs_prune_job.test.id
    error_message = "Data source ID does not match resource ID"
  }

  assert {
    condition     = data.pbs_prune_job.test.store == pbs_prune_job.test.store
    error_message = "Data source store does not match resource store"
  }

  assert {
    condition     = data.pbs_prune_job.test.schedule == pbs_prune_job.test.schedule
    error_message = "Data source schedule does not match resource schedule"
  }

  assert {
    condition     = data.pbs_prune_job.test.keep_last == pbs_prune_job.test.keep_last
    error_message = "Data source keep_last does not match resource keep_last"
  }

  assert {
    condition     = data.pbs_prune_job.test.keep_daily == pbs_prune_job.test.keep_daily
    error_message = "Data source keep_daily does not match resource keep_daily"
  }

  assert {
    condition     = data.pbs_prune_job.test.comment == pbs_prune_job.test.comment
    error_message = "Data source comment does not match resource comment"
  }
}
