run "create_verify_job" {
  variables {
    pbs_endpoint    = "https://${PBS_ADDRESS}:8007"
    pbs_insecure    = true
    job_id          = "tftest-verify-job"
    schedule        = "weekly"
    namespace       = "prod"
    ignore_verified = true
    outdated_after  = 30
    max_depth       = 3
    comment         = "Test verify job"
  }

  assert {
    condition     = pbs_verify_job.test.id == var.job_id
    error_message = "Job ID mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.store == "datastore1"
    error_message = "Store mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.schedule == "weekly"
    error_message = "Schedule mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.namespace == "prod"
    error_message = "Namespace mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.ignore_verified == true
    error_message = "ignore_verified mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.outdated_after == 30
    error_message = "outdated_after mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.max_depth == 3
    error_message = "max_depth mismatch"
  }

  assert {
    condition     = pbs_verify_job.test.comment == "Test verify job"
    error_message = "comment mismatch"
  }
}

run "update_verify_job" {
  variables {
    pbs_endpoint    = "https://${PBS_ADDRESS}:8007"
    pbs_insecure    = true
    job_id          = "tftest-verify-job"
    schedule        = "monthly"
    namespace       = "prod"
    ignore_verified = false
    outdated_after  = 60
    max_depth       = 5
    comment         = "Updated test verify job"
  }

  assert {
    condition     = pbs_verify_job.test.schedule == "monthly"
    error_message = "schedule was not updated"
  }

  assert {
    condition     = pbs_verify_job.test.ignore_verified == false
    error_message = "ignore_verified was not updated"
  }

  assert {
    condition     = pbs_verify_job.test.outdated_after == 60
    error_message = "outdated_after was not updated"
  }

  assert {
    condition     = pbs_verify_job.test.max_depth == 5
    error_message = "max_depth was not updated"
  }

  assert {
    condition     = pbs_verify_job.test.comment == "Updated test verify job"
    error_message = "comment was not updated"
  }

  assert {
    condition     = pbs_verify_job.test.namespace == "prod"
    error_message = "namespace should remain unchanged"
  }
}
