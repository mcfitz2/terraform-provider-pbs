run "create_remote" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    remote_name  = "tftest-remote"
    host         = "pbs.example.com"
    auth_id      = "sync@pbs!test-token"
    password     = "test-password-12345"
    comment      = "Test remote server"
  }

  assert {
    condition     = pbs_remote.test.name == var.remote_name
    error_message = "Remote name mismatch"
  }

  assert {
    condition     = pbs_remote.test.host == "pbs.example.com"
    error_message = "Host mismatch"
  }

  assert {
    condition     = pbs_remote.test.auth_id == "sync@pbs!test-token"
    error_message = "auth_id mismatch"
  }

  assert {
    condition     = pbs_remote.test.comment == "Test remote server"
    error_message = "comment mismatch"
  }

  assert {
    condition     = pbs_remote.test.digest != null && pbs_remote.test.digest != ""
    error_message = "digest should be populated"
  }
}

run "update_remote" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    remote_name  = "tftest-remote"
    host         = "backup.example.com"
    port         = 8008
    auth_id      = "sync@pbs!test-token"
    password     = "test-password-12345"
    fingerprint  = "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
    comment      = "Updated test remote server"
  }

  assert {
    condition     = pbs_remote.test.host == "backup.example.com"
    error_message = "host was not updated"
  }

  assert {
    condition     = pbs_remote.test.port == 8008
    error_message = "port was not updated"
  }

  assert {
    condition     = pbs_remote.test.fingerprint == "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
    error_message = "fingerprint was not updated"
  }

  assert {
    condition     = pbs_remote.test.comment == "Updated test remote server"
    error_message = "comment was not updated"
  }

  assert {
    condition     = pbs_remote.test.name == var.remote_name
    error_message = "name should remain unchanged"
  }
}

run "clear_optional_fields" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    remote_name  = "tftest-remote"
    host         = "backup.example.com"
    auth_id      = "sync@pbs!test-token"
    password     = "test-password-12345"
  }

  assert {
    condition     = pbs_remote.test.port == null
    error_message = "port should be cleared"
  }

  assert {
    condition     = pbs_remote.test.fingerprint == null
    error_message = "fingerprint should be cleared"
  }

  assert {
    condition     = pbs_remote.test.comment == null
    error_message = "comment should be cleared"
  }

  assert {
    condition     = pbs_remote.test.host == "backup.example.com"
    error_message = "host should remain"
  }
}

run "update_password" {
  variables {
    pbs_endpoint = "https://${PBS_ADDRESS}:8007"
    pbs_insecure = true
    remote_name  = "tftest-pass-remote"
    host         = "pbs.example.com"
    auth_id      = "admin@pam"
    password     = "new-password-54321"
  }

  assert {
    condition     = pbs_remote.test.name == var.remote_name
    error_message = "Remote should exist with new password"
  }

  assert {
    condition     = pbs_remote.test.host == "pbs.example.com"
    error_message = "Host should remain unchanged"
  }
}
