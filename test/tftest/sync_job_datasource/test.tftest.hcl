# Test for reading a single sync job via data source
#
# This test verifies that:
# 1. A datastore can be created
# 2. A remote can be created
# 3. A sync job can be created that references both
# 4. The sync job data source can read the job
# 5. All attributes match between the resource and data source
#
# This was converted from a Go tfexec test that was flaky in CI due to
# timing issues. The native Terraform test framework eliminates those issues
# by using the same execution path as manual `terraform apply`.

variables {
  # Generate unique names with timestamp to avoid conflicts
  datastore_name = "ds-sync-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  remote_name = "remote-sync-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  job_id = "sync-ds-${formatdate("YYYYMMDDhhmmss", timestamp())}"
}

# Provider configuration - can be overridden via environment variables
provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

# Run block 1: Create all resources and read via data source
run "create_and_read" {
  command = apply

  variables {
    datastore_name = var.datastore_name
    remote_name = var.remote_name
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

  # Verify remote was created
  assert {
    condition     = pbs_remote.test.name == var.remote_name
    error_message = "Remote name does not match expected value"
  }

  assert {
    condition     = pbs_remote.test.host == "remote.example.com"
    error_message = "Remote host does not match expected value"
  }

  assert {
    condition     = pbs_remote.test.auth_id == "test@pbs"
    error_message = "Remote auth_id does not match expected value"
  }

  # Verify sync job was created
  assert {
    condition     = pbs_sync_job.test.id == var.job_id
    error_message = "Sync job ID does not match expected value"
  }

  assert {
    condition     = pbs_sync_job.test.store == var.datastore_name
    error_message = "Sync job store does not match datastore name"
  }

  assert {
    condition     = pbs_sync_job.test.remote == var.remote_name
    error_message = "Sync job remote does not match remote name"
  }

  assert {
    condition     = pbs_sync_job.test.remote_store == "backup"
    error_message = "Sync job remote_store does not match expected value"
  }

  assert {
    condition     = pbs_sync_job.test.schedule == "hourly"
    error_message = "Sync job schedule does not match expected value"
  }

  # Verify data source reads the job correctly
  assert {
    condition     = data.pbs_sync_job.test.id == pbs_sync_job.test.id
    error_message = "Data source ID does not match resource ID"
  }

  assert {
    condition     = data.pbs_sync_job.test.store == pbs_sync_job.test.store
    error_message = "Data source store does not match resource store"
  }

  assert {
    condition     = data.pbs_sync_job.test.remote == pbs_sync_job.test.remote
    error_message = "Data source remote does not match resource remote"
  }

  assert {
    condition     = data.pbs_sync_job.test.remote_store == pbs_sync_job.test.remote_store
    error_message = "Data source remote_store does not match resource remote_store"
  }

  assert {
    condition     = data.pbs_sync_job.test.schedule == pbs_sync_job.test.schedule
    error_message = "Data source schedule does not match resource schedule"
  }

  assert {
    condition     = data.pbs_sync_job.test.comment == pbs_sync_job.test.comment
    error_message = "Data source comment does not match resource comment"
  }
}
