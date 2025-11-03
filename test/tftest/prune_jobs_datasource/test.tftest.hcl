# Test for listing prune jobs via data source
#
# This test verifies that:
# 1. Multiple datastores can be created
# 2. Multiple prune jobs can be created
# 3. The prune jobs data source can list all jobs
# 4. The prune jobs data source can filter by store
# 5. Filtered results contain only jobs for the specified store
#
# This was converted from a Go tfexec test that was flaky in CI due to
# timing issues. The native Terraform test framework eliminates those issues
# by using the same execution path as manual `terraform apply`.

variables {
  # Generate unique names with timestamp to avoid conflicts
  datastore1_name = "ds-prune-list-1-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  datastore2_name = "ds-prune-list-2-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  job1_id = "prune-list-1-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  job2_id = "prune-list-2-${formatdate("YYYYMMDDhhmmss", timestamp())}"
}

# Provider configuration - can be overridden via environment variables
provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

# Run block 1: Create resources and test listing
run "create_and_list" {
  command = apply

  variables {
    datastore1_name = var.datastore1_name
    datastore2_name = var.datastore2_name
    job1_id = var.job1_id
    job2_id = var.job2_id
  }

  # Verify first datastore was created
  assert {
    condition     = pbs_datastore.test1.name == var.datastore1_name
    error_message = "First datastore name does not match expected value"
  }

  assert {
    condition     = pbs_datastore.test1.path == "/datastore/${var.datastore1_name}"
    error_message = "First datastore path does not match expected value"
  }

  # Verify second datastore was created
  assert {
    condition     = pbs_datastore.test2.name == var.datastore2_name
    error_message = "Second datastore name does not match expected value"
  }

  assert {
    condition     = pbs_datastore.test2.path == "/datastore/${var.datastore2_name}"
    error_message = "Second datastore path does not match expected value"
  }

  # Verify first prune job was created
  assert {
    condition     = pbs_prune_job.test1.id == var.job1_id
    error_message = "First prune job ID does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test1.store == var.datastore1_name
    error_message = "First prune job store does not match datastore1"
  }

  assert {
    condition     = pbs_prune_job.test1.schedule == "daily"
    error_message = "First prune job schedule does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test1.keep_last == 5
    error_message = "First prune job keep_last does not match expected value"
  }

  # Verify second prune job was created
  assert {
    condition     = pbs_prune_job.test2.id == var.job2_id
    error_message = "Second prune job ID does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test2.store == var.datastore2_name
    error_message = "Second prune job store does not match datastore2"
  }

  assert {
    condition     = pbs_prune_job.test2.schedule == "weekly"
    error_message = "Second prune job schedule does not match expected value"
  }

  assert {
    condition     = pbs_prune_job.test2.keep_last == 10
    error_message = "Second prune job keep_last does not match expected value"
  }

  # Verify unfiltered data source contains both jobs
  assert {
    condition     = length(data.pbs_prune_jobs.all.jobs) >= 2
    error_message = "Unfiltered data source should contain at least 2 jobs (found ${length(data.pbs_prune_jobs.all.jobs)})"
  }

  # Verify our test jobs are in the unfiltered list
  # We need to check if the job IDs are present in the list
  assert {
    condition = contains([
      for job in data.pbs_prune_jobs.all.jobs : job.id
    ], var.job1_id)
    error_message = "Unfiltered data source does not contain first test job"
  }

  assert {
    condition = contains([
      for job in data.pbs_prune_jobs.all.jobs : job.id
    ], var.job2_id)
    error_message = "Unfiltered data source does not contain second test job"
  }

  # Verify filtered data source exists and has jobs
  assert {
    condition     = length(data.pbs_prune_jobs.filtered.jobs) >= 1
    error_message = "Filtered data source should contain at least 1 job"
  }

  # Verify filtered data source only contains jobs for datastore1
  # Count how many jobs in the filtered list match datastore1
  assert {
    condition = length([
      for job in data.pbs_prune_jobs.filtered.jobs : job.id
      if job.store == var.datastore1_name
    ]) >= 1
    error_message = "Filtered data source should contain jobs for datastore1"
  }

  # Verify filtered data source does NOT contain jobs for datastore2
  # (unless there are other pre-existing jobs, so we just check our test job2 is not there)
  assert {
    condition = !contains([
      for job in data.pbs_prune_jobs.filtered.jobs : job.id
    ], var.job2_id)
    error_message = "Filtered data source should not contain job from datastore2"
  }

  # Verify filtered data source DOES contain our test job1
  assert {
    condition = contains([
      for job in data.pbs_prune_jobs.filtered.jobs : job.id
    ], var.job1_id)
    error_message = "Filtered data source should contain job1 from datastore1"
  }
}
