run "read_datastore_datasource" {
  variables {
    datastore_name  = "tftest-ds-datasource"
  }

  # Verify data source matches resource
  assert {
    condition     = data.pbs_datastore.test.name == pbs_datastore.test.name
    error_message = "Data source name doesn't match resource"
  }

  assert {
    condition     = data.pbs_datastore.test.path == pbs_datastore.test.path
    error_message = "Data source path doesn't match resource"
  }

  assert {
    condition     = data.pbs_datastore.test.comment == pbs_datastore.test.comment
    error_message = "Data source comment doesn't match resource"
  }

  assert {
    condition     = data.pbs_datastore.test.gc_schedule == pbs_datastore.test.gc_schedule
    error_message = "Data source gc_schedule doesn't match resource"
  }

  assert {
    condition     = data.pbs_datastore.test.name == var.datastore_name
    error_message = "Data source name mismatch with input"
  }

  assert {
    condition     = output.datasource_name == var.datastore_name
    error_message = "Output name mismatch"
  }

  assert {
    condition     = output.datasource_path == "/datastore/${var.datastore_name}"
    error_message = "Output path mismatch"
  }
}
