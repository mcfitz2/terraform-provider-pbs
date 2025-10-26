# List available datastores on a remote PBS server
data "pbs_remote_stores" "backup_stores" {
  remote_name = pbs_remote.backup_server.name
}

# Use the list of stores in output
output "available_stores" {
  description = "List of datastores available on the remote server"
  value       = data.pbs_remote_stores.backup_stores.stores
}

# Example: validate a specific store exists before creating a sync job
locals {
  target_store = "offsite-backups"
  store_exists = contains(data.pbs_remote_stores.backup_stores.stores, local.target_store)
}

# Reference in sync job configuration
resource "pbs_sync_job" "example" {
  id          = "sync-to-backup"
  store       = "local-backups"
  remote      = pbs_remote.backup_server.name
  remote_store = local.target_store
  schedule    = "daily"
  
  # Ensure the remote store exists
  depends_on = [data.pbs_remote_stores.backup_stores]
}
