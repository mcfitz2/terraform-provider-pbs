# List namespaces available in a remote datastore
data "pbs_remote_namespaces" "backup_namespaces" {
  remote_name = pbs_remote.backup_server.name
  store       = "offsite-backups"
}

# Use namespace data in output
output "available_namespaces" {
  description = "List of namespaces in the remote datastore"
  value       = data.pbs_remote_namespaces.backup_namespaces.namespaces
}

# Example: filter sync job to specific namespace
resource "pbs_sync_job" "namespace_sync" {
  id           = "sync-prod-namespace"
  store        = "local-backups"
  remote       = pbs_remote.backup_server.name
  remote_store = "offsite-backups"
  
  # Sync only the production namespace
  remote_ns = "production"
  schedule  = "02:00"
  
  # Verify namespace exists on remote
  depends_on = [data.pbs_remote_namespaces.backup_namespaces]
}

# Example: dynamically create sync jobs for each namespace
resource "pbs_sync_job" "per_namespace_sync" {
  for_each = toset(data.pbs_remote_namespaces.backup_namespaces.namespaces)
  
  id           = "sync-${replace(each.value, "/", "-")}"
  store        = "local-backups"
  remote       = pbs_remote.backup_server.name
  remote_store = "offsite-backups"
  remote_ns    = each.value
  schedule     = "03:00"
}
