run "read_verify_job_datasource" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
  }

  assert {
    condition     = data.pbs_verify_job.test.id == pbs_verify_job.test.id
    error_message = "Data source id doesn't match resource"
  }

  assert {
    condition     = data.pbs_verify_job.test.store == pbs_verify_job.test.store
    error_message = "Data source store doesn't match resource"
  }

  assert {
    condition     = data.pbs_verify_job.test.schedule == pbs_verify_job.test.schedule
    error_message = "Data source schedule doesn't match resource"
  }

  assert {
    condition     = data.pbs_verify_job.test.namespace == pbs_verify_job.test.namespace
    error_message = "Data source namespace doesn't match resource"
  }

  assert {
    condition     = data.pbs_verify_job.test.ignore_verified == pbs_verify_job.test.ignore_verified
    error_message = "Data source ignore_verified doesn't match resource"
  }

  assert {
    condition     = data.pbs_verify_job.test.outdated_after == pbs_verify_job.test.outdated_after
    error_message = "Data source outdated_after doesn't match resource"
  }
}
