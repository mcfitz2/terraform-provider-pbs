# List backup groups in a remote datastore
data "pbs_remote_groups" "all_groups" {
  remote_name = pbs_remote.backup_server.name
  store       = "offsite-backups"
}

# List backup groups in a specific namespace
data "pbs_remote_groups" "production_groups" {
  remote_name = pbs_remote.backup_server.name
  store       = "offsite-backups"
  namespace   = "production"
}

# Use group data in outputs
output "backup_groups_count" {
  description = "Number of backup groups on remote server"
  value       = length(data.pbs_remote_groups.all_groups.groups)
}

output "production_groups" {
  description = "Backup groups in production namespace"
  value       = data.pbs_remote_groups.production_groups.groups
}

# Example: validate specific backup group exists before sync
locals {
  required_group = "vm/100"
  group_exists   = contains(data.pbs_remote_groups.production_groups.groups, local.required_group)
}

# Example: filter sync job to specific groups
resource "pbs_sync_job" "group_filtered_sync" {
  id           = "sync-specific-vms"
  store        = "local-backups"
  remote       = pbs_remote.backup_server.name
  remote_store = "offsite-backups"
  remote_ns    = "production"
  
  # Only sync specific backup types
  group_filter = "vm"
  schedule     = "04:00"
  
  depends_on = [data.pbs_remote_groups.production_groups]
}
