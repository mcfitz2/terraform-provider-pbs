# Example PBS remote configuration with API token authentication
resource "pbs_remote" "backup_server" {
  name     = "backup-pbs"
  host     = "backup.example.com"
  port     = 8007
  auth_id  = "sync@pbs!backup-token"
  password = "abcd1234-5678-90ef-ghij-klmnopqrstuv"
  
  # Optional: X509 certificate fingerprint for TLS verification
  fingerprint = "AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF:00:11:22:33:44:55:66:77:88:99"
  
  # Optional: Description
  comment = "Backup PBS server for sync jobs"
}

# Example PBS remote with password authentication
resource "pbs_remote" "offsite_pbs" {
  name     = "offsite"
  host     = "192.168.1.100"
  auth_id  = "admin@pam"
  password = var.offsite_password  # Recommended: use variable for sensitive data
  
  comment = "Offsite PBS for disaster recovery"
}

# Example without port (defaults to 8007)
resource "pbs_remote" "local_pbs" {
  name     = "local-backup"
  host     = "pbs.local"
  auth_id  = "sync@pbs!sync-token"
  password = var.sync_password
}
