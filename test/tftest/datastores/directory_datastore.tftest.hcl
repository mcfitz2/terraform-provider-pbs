provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

run "create_directory_datastore" {
  variables {
    test_name   = "tftest-dir-${var.test_id}"
    comment     = "Test directory datastore"
    gc_schedule = "daily"
  }

  # Verify resource creation
  assert {
    condition     = pbs_datastore.test_directory.name == var.test_name
    error_message = "Datastore name mismatch"
  }

  assert {
    condition     = pbs_datastore.test_directory.path == "/datastore/${var.test_name}"
    error_message = "Datastore path mismatch"
  }

  assert {
    condition     = pbs_datastore.test_directory.comment == "Test directory datastore"
    error_message = "Datastore comment mismatch"
  }

  assert {
    condition     = pbs_datastore.test_directory.gc_schedule == "daily"
    error_message = "Datastore gc_schedule mismatch"
  }

  assert {
    condition     = output.datastore_name == var.test_name
    error_message = "Output datastore_name mismatch"
  }
}

run "update_directory_datastore" {
  variables {
    test_name   = "tftest-dir-${var.test_id}"
    comment     = "Updated test directory datastore"
    gc_schedule = "weekly"
  }

  # Verify update was applied
  assert {
    condition     = pbs_datastore.test_directory.comment == "Updated test directory datastore"
    error_message = "Datastore comment was not updated"
  }

  assert {
    condition     = pbs_datastore.test_directory.gc_schedule == "weekly"
    error_message = "Datastore gc_schedule was not updated"
  }

  assert {
    condition     = pbs_datastore.test_directory.name == var.test_name
    error_message = "Datastore name should remain unchanged"
  }

  assert {
    condition     = output.datastore_comment == "Updated test directory datastore"
    error_message = "Output comment mismatch after update"
  }

  assert {
    condition     = output.datastore_gc_schedule == "weekly"
    error_message = "Output gc_schedule mismatch after update"
  }
}
