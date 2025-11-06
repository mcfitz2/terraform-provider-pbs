run "create_prune_job" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    job_id       = "tftest-prune-job"
    schedule     = "daily"
    keep_last    = 7
    keep_daily   = 14
    keep_weekly  = 8
    keep_monthly = 12
    keep_yearly  = 3
    max_depth    = 3
    comment      = "Test prune job"
  }

  assert {
    condition     = pbs_prune_job.test.id == var.job_id
    error_message = "Job ID mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.store == "datastore1"
    error_message = "Store mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.schedule == "daily"
    error_message = "Schedule mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.keep_last == 7
    error_message = "keep_last mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.keep_daily == 14
    error_message = "keep_daily mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.keep_weekly == 8
    error_message = "keep_weekly mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.keep_monthly == 12
    error_message = "keep_monthly mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.keep_yearly == 3
    error_message = "keep_yearly mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.max_depth == 3
    error_message = "max_depth mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.comment == "Test prune job"
    error_message = "comment mismatch"
  }
}

run "update_prune_job" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    job_id       = "tftest-prune-job"
    schedule     = "weekly"
    keep_last    = 10
    keep_daily   = 21
    keep_weekly  = 12
    keep_monthly = 18
    keep_yearly  = 5
    max_depth    = 5
    comment      = "Updated test prune job"
  }

  assert {
    condition     = pbs_prune_job.test.schedule == "weekly"
    error_message = "Schedule was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.keep_last == 10
    error_message = "keep_last was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.keep_daily == 21
    error_message = "keep_daily was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.keep_weekly == 12
    error_message = "keep_weekly was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.keep_monthly == 18
    error_message = "keep_monthly was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.keep_yearly == 5
    error_message = "keep_yearly was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.max_depth == 5
    error_message = "max_depth was not updated"
  }

  assert {
    condition     = pbs_prune_job.test.comment == "Updated test prune job"
    error_message = "comment was not updated"
  }
}

run "prune_job_with_namespace_filter" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    job_id       = "tftest-prune-filter"
    schedule     = "daily"
    keep_last    = 5
    keep_daily   = 0
    keep_weekly  = 0
    keep_monthly = 0
    keep_yearly  = 0
    max_depth    = 2
    comment      = "Prune job with filters"
    namespace    = "vm"
  }

  assert {
    condition     = pbs_prune_job.test.namespace == "vm"
    error_message = "namespace filter mismatch"
  }

  assert {
    condition     = pbs_prune_job.test.max_depth == 2
    error_message = "max_depth mismatch for filtered job"
  }
}
