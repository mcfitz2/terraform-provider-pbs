run "create_sync_job" {
  variables {
    job_type         = "sync"
    job_id           = "sync-${var.test_id}"
    remote           = "remote1"
    remote_store     = "backup"
    remote_namespace = "prod"
    namespace        = "mirror"
    schedule         = "hourly"
    remove_vanished  = true
    resync_corrupt   = true
    rate_in          = "10M"
    rate_out         = "5M"
    burst_in         = "15M"
    burst_out        = "10M"
    comment          = "Test sync job"
  }

  assert {
    condition     = pbs_sync_job.test[0].id == var.job_id
    error_message = "Job ID mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].store == "datastore1"
    error_message = "Store mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].remote == "remote1"
    error_message = "Remote mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].remote_store == "backup"
    error_message = "remote_store mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].remote_namespace == "prod"
    error_message = "remote_namespace mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].namespace == "mirror"
    error_message = "namespace mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].schedule == "hourly"
    error_message = "schedule mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].remove_vanished == true
    error_message = "remove_vanished mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].resync_corrupt == true
    error_message = "resync_corrupt mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].rate_in == "10M"
    error_message = "rate_in mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].rate_out == "5M"
    error_message = "rate_out mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].burst_in == "15M"
    error_message = "burst_in mismatch"
  }

  assert {
    condition     = pbs_sync_job.test[0].burst_out == "10M"
    error_message = "burst_out mismatch"
  }
}

run "update_sync_job" {
  variables {
    job_type         = "sync"
    job_id           = "sync-${var.test_id}"
    remote           = "remote1"
    remote_store     = "backup"
    remote_namespace = "prod"
    namespace        = "mirror"
    schedule         = "daily"
    remove_vanished  = false
    verified_only    = true
    run_on_mount     = true
    transfer_last    = 3600
    rate_in          = "20M"
    rate_out         = "8M"
    burst_in         = "25M"
    burst_out        = "12M"
    comment          = "Updated test sync job"
  }

  assert {
    condition     = pbs_sync_job.test[0].schedule == "daily"
    error_message = "schedule was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].remove_vanished == false
    error_message = "remove_vanished was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].verified_only == true
    error_message = "verified_only was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].run_on_mount == true
    error_message = "run_on_mount was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].transfer_last == 3600
    error_message = "transfer_last was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].rate_in == "20M"
    error_message = "rate_in was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].rate_out == "8M"
    error_message = "rate_out was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].burst_in == "25M"
    error_message = "burst_in was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].burst_out == "12M"
    error_message = "burst_out was not updated"
  }

  assert {
    condition     = pbs_sync_job.test[0].comment == "Updated test sync job"
    error_message = "comment was not updated"
  }
}

run "sync_job_with_group_filter" {
  variables {
    job_type     = "sync"
    job_id       = "sync-f-${var.test_id}"
    remote       = "remote1"
    remote_store = "backup"
    schedule     = "daily"
    namespace    = "production"
    group_filter = ["group:vm/node1", "group:ct/node2"]
    comment      = "Sync job with filters"
  }

  assert {
    condition     = length(pbs_sync_job.test[0].group_filter) == 2
    error_message = "group_filter length mismatch"
  }

  assert {
    condition     = contains(pbs_sync_job.test[0].group_filter, "group:vm/node1")
    error_message = "group_filter missing vm/node1"
  }

  assert {
    condition     = contains(pbs_sync_job.test[0].group_filter, "group:ct/node2")
    error_message = "group_filter missing ct/node2"
  }

  assert {
    condition     = pbs_sync_job.test[0].namespace == "production"
    error_message = "namespace mismatch"
  }
}
