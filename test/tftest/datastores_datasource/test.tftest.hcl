# Test for listing datastores via data source
#
# This test verifies that:
# 1. Multiple datastores can be created
# 2. The datastores data source can list all datastores
# 3. Test datastores appear in the list
# 4. Datastore attributes are correctly populated
#
# This was converted from a Go tfexec test that was flaky in CI due to
# timing issues. The native Terraform test framework eliminates those issues
# by using the same execution path as manual `terraform apply`.

variables {
  # Generate unique names with timestamp to avoid conflicts
  datastore1_name = "ds-list-1-${formatdate("YYYYMMDDhhmmss", timestamp())}"
  datastore2_name = "ds-list-2-${formatdate("YYYYMMDDhhmmss", timestamp())}"
}

# Provider configuration - can be overridden via environment variables
provider "pbs" {
  endpoint = var.pbs_endpoint
  username = var.pbs_username
  password = var.pbs_password
  insecure = true
}

# Run block 1: Create datastores and test listing
run "create_and_list" {
  command = apply

  variables {
    datastore1_name = var.datastore1_name
    datastore2_name = var.datastore2_name
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

  assert {
    condition     = pbs_datastore.test1.comment == "First test datastore"
    error_message = "First datastore comment does not match expected value"
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

  assert {
    condition     = pbs_datastore.test2.comment == "Second test datastore"
    error_message = "Second datastore comment does not match expected value"
  }

  # Verify data source returns a list with at least 2 datastores
  assert {
    condition     = length(data.pbs_datastores.all.stores) >= 2
    error_message = "Data source should contain at least 2 datastores (found ${length(data.pbs_datastores.all.stores)})"
  }

  # Verify our first test datastore is in the list
  assert {
    condition = contains([
      for store in data.pbs_datastores.all.stores : store.name
    ], var.datastore1_name)
    error_message = "Data source does not contain first test datastore"
  }

  # Verify our second test datastore is in the list
  assert {
    condition = contains([
      for store in data.pbs_datastores.all.stores : store.name
    ], var.datastore2_name)
    error_message = "Data source does not contain second test datastore"
  }

  # Verify the datastores have the expected attributes
  # Find our first datastore in the list and check its properties
  assert {
    condition = anytrue([
      for store in data.pbs_datastores.all.stores :
      store.name == var.datastore1_name && store.comment == "First test datastore"
    ])
    error_message = "First datastore in list does not have expected comment"
  }

  assert {
    condition = anytrue([
      for store in data.pbs_datastores.all.stores :
      store.name == var.datastore2_name && store.comment == "Second test datastore"
    ])
    error_message = "Second datastore in list does not have expected comment"
  }
}
