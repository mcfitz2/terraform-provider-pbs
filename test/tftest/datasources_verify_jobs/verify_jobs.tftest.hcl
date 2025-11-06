run "list_verify_jobs" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
  }

  assert {
    condition     = length(data.pbs_verify_jobs.all.jobs) >= 2
    error_message = "Should have at least 2 verify jobs"
  }

  assert {
    condition     = contains([for j in data.pbs_verify_jobs.all.jobs : j.id], "tftest-verify-list-1")
    error_message = "Should contain test verify job 1"
  }

  assert {
    condition     = contains([for j in data.pbs_verify_jobs.all.jobs : j.id], "tftest-verify-list-2")
    error_message = "Should contain test verify job 2"
  }

  assert {
    condition     = output.job_count >= 2
    error_message = "Job count output mismatch"
  }
}
