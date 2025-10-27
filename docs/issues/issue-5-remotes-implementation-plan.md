# Implementation Plan for Issue #5: PBS Remotes Support

**Issue:** [#5 - Add Terraform support for PBS remotes](https://github.com/mcfitz2/terraform-provider-pbs/issues/5)  
**Date:** October 26, 2025  
**Status:** In Progress

## Overview
Add first-class Terraform support for PBS remotes, enabling users to manage remote configurations through Terraform and use helper data sources to discover remote stores/namespaces for sync job configuration.

---

## Phase 1: API Client Layer (pbs/remotes)

### 1. Create `pbs/remotes/remotes.go` API client package
- Define `Remote` struct matching PBS API schema:
  - Core fields: `name`, `host`, `port` (optional int), `auth-id`, `password`, `fingerprint` (optional)
  - Metadata: `comment`, `digest`, `delete` ([]string for updates)
- Implement CRUD methods:
  - `ListRemotes(ctx)` → GET `/config/remote` (returns minimal list)
  - `GetRemote(ctx, name)` → GET `/config/remote/{name}` (full details)
  - `CreateRemote(ctx, remote)` → POST `/config/remote`
  - `UpdateRemote(ctx, name, remote)` → PUT `/config/remote/{name}` with digest + delete semantics
  - `DeleteRemote(ctx, name, digest)` → DELETE `/config/remote/{name}`
- Follow existing patterns from `pbs/endpoints/s3.go` and `pbs/datastores/datastores.go`

### 2. Add remote scan helper methods
- `ListRemoteStores(ctx, name)` → GET `/config/remote/{name}/scan`
  - Returns array of datastore info (name, comment, maintenance status)
- `ListRemoteNamespaces(ctx, name, store)` → GET `/config/remote/{name}/scan/{store}/namespaces`
  - Returns array of namespace objects (ns, comment)
- `ListRemoteGroups(ctx, name, store, namespace)` → GET `/config/remote/{name}/scan/{store}/groups`
  - Optional namespace parameter for filtering
  - Returns backup group info (backup-type, backup-id, count, last-backup, owner, etc.)

### 3. Wire remotes client into `pbs/client.go`
- Add `Remotes *remotes.Client` field to `Client` struct
- Initialize in `NewClient()` alongside existing clients

---

## Phase 2: Terraform Resource (fwprovider/resources/remotes)

### 4. Implement `pbs_remote` resource in `remote.go`
- **Schema attributes:**
  - `name` (string, Required, ForceNew) - Remote ID
  - `host` (string, Required) - Remote PBS address (hostname/IP)
  - `port` (int, Optional) - Port number
  - `auth_id` (string, Required) - Authentication ID (e.g., `user@pam`)
  - `password` (string, Required, Sensitive) - Password or auth token
  - `fingerprint` (string, Optional) - X509 cert fingerprint (sha256)
  - `comment` (string, Optional) - Description
  - `digest` (string, Computed, UseStateForUnknown) - For optimistic locking

- **CRUD implementation:**
  - **Create**: Call `CreateRemote`, read back with `GetRemote` to capture digest
  - **Read**: Call `GetRemote`, handle password write-only semantics (don't overwrite state if API doesn't return it)
  - **Update**: Build `delete` array for cleared optional fields (port, fingerprint, comment), send with digest
  - **Delete**: Call `DeleteRemote` with digest from state
  
- **Validation**: Add validators for `auth_id` pattern, port range, fingerprint format matching PBS regex

---

## Phase 3: Data Sources (fwprovider/datasources/remotes)

### 5. Create `pbs_remote_stores` data source
- Input: `remote` (string, Required)
- Output: `stores` (list of objects with `name`, `comment`, `maintenance`)
- Calls `ListRemoteStores` API method

### 6. Create `pbs_remote_namespaces` data source
- Input: `remote` (Required), `store` (Required)
- Output: `namespaces` (list of objects with `ns`, `comment`)
- Calls `ListRemoteNamespaces` API method

### 7. Create `pbs_remote_groups` data source
- Input: `remote` (Required), `store` (Required), `namespace` (Optional)
- Output: `groups` (list with backup-type, backup-id, backup-count, last-backup, owner, comment)
- Calls `ListRemoteGroups` API method

---

## Phase 4: Provider Integration

### 8. Register components in `fwprovider/provider.go`
- Add `remotes.NewRemoteResource` to `Resources()` function
- Add all three data source constructors to `DataSources()` function
- Create directory structure: `fwprovider/datasources/remotes/`

### 9. Update sync job resource (optional enhancement in `sync_job.go`)
- Add validation hint or plan modifier suggesting use of `pbs_remote` resource for the `remote` attribute
- Document dependency pattern in resource description

---

## Phase 5: Documentation

### 10. Create `docs/resources/pbs_remote.md`
- Basic CRUD examples
- Password handling semantics (write-only, use sensitive values properly)
- Optimistic locking explanation (digest conflicts)
- Integration pattern with sync jobs
- Example showing:
  ```hcl
  resource "pbs_remote" "backup_server" {
    name        = "remote-pbs"
    host        = "backup.example.com"
    port        = 8007
    auth_id     = "sync@pbs"
    password    = var.remote_password
    fingerprint = "AA:BB:CC:..."
    comment     = "Production backup server"
  }
  
  data "pbs_remote_stores" "available" {
    remote = pbs_remote.backup_server.name
  }
  
  resource "pbs_sync_job" "pull_backups" {
    id           = "pull-from-prod"
    store        = "local-store"
    remote       = pbs_remote.backup_server.name
    remote_store = data.pbs_remote_stores.available.stores[0].name
    schedule     = "daily"
  }
  ```

### 11. Update `docs/resources/pbs_sync_job.md`
- Reference `pbs_remote` resource in examples
- Show data source usage for discovering remote stores/namespaces
- Document best practice of managing remotes via Terraform

---

## Phase 6: Testing

### 12. Create `test/integration/remotes_test.go`
- **TestRemoteResource**:
  - Create remote with all fields
  - Read back and verify
  - Update mutable fields (host, port, comment, fingerprint)
  - Update with stale digest → expect 4xx conflict error
  - Clear optional field (test delete mechanics)
  - Delete remote
  
- **TestRemoteDataSources**:
  - Create test remote pointing to same PBS instance
  - Use `pbs_remote_stores` → verify returns datastores
  - Use `pbs_remote_namespaces` with discovered store → verify structure
  - Use `pbs_remote_groups` → verify backup group list format

- **TestSyncJobWithRemote** (end-to-end):
  - Provision remote via Terraform
  - Use data sources to pick store/namespace
  - Create sync job referencing remote + discovered values
  - Apply and verify job exists in PBS
  - Clean up (job, then remote)

### 13. Test execution strategy
- Remote tests require two PBS instances or mocked remote endpoint
- Can use docker-compose.test.yml pattern with two PBS containers
- Or mock remote endpoints in test fixtures for CI
- Integration with existing `test/run_docker_tests.sh` workflow

---

## Acceptance Criteria Coverage

✅ **Terraform can create, update, and delete remotes entirely via API**
- Phases 2 & 6 cover full CRUD with verification

✅ **Changing mutable fields updates PBS; clearing optional fields uses delete semantics**
- Update implementation includes `delete` array computation
- Test suite validates this behavior

✅ **Stale digest triggers conflict error**
- Update implementation passes digest
- Test explicitly verifies conflict scenario

✅ **Data sources list remote stores/namespaces/groups**
- Phase 3 implements all three data sources
- Phase 6 tests against live PBS

✅ **End-to-end sync job workflow**
- TestSyncJobWithRemote demonstrates full pattern
- Documentation shows complete example

---

## Technical Notes

### API Patterns to Follow
- Use same CRUD structure as `pbs/endpoints/s3.go`
- Handle digest in updates like `pbs/datastores/datastores.go`
- Password write-only semantics: don't compare in state, only send on create/update
- Remote fields use kebab-case in JSON (`auth-id`, `remote-ns`), map to snake_case in Terraform

### Validation
- `auth_id` pattern: `/^(?:(?:[^\s:/[:cntrl:]]+)@(?:[A-Za-z0-9_][A-Za-z0-9._\-]*)|(?:[^\s:/[:cntrl:]]+)@(?:[A-Za-z0-9_][A-Za-z0-9._\-]*)!(?:[A-Za-z0-9_][A-Za-z0-9._\-]*))$/`
- `fingerprint` pattern: `/^(?:[0-9a-fA-F][0-9a-fA-F])(?::[0-9a-fA-F][0-9a-fA-F]){31}$/`
- `name` pattern: `/^[A-Za-z0-9_][A-Za-z0-9._\-]*$/` (3-32 chars)

### Dependencies
- Sync jobs already reference `remote` by name - no code changes needed there
- Data sources are read-only, no state management complexity
- Remote permissions: `Remote.Modify` for CRUD, `Remote.Audit` for read/scan

---

## Implementation Checklist

- [ ] Phase 1: API Client Layer
  - [ ] Create pbs/remotes/remotes.go with Remote struct and CRUD methods
  - [ ] Add scan helper methods (ListRemoteStores, ListRemoteNamespaces, ListRemoteGroups)
  - [ ] Wire into pbs/client.go
  
- [ ] Phase 2: Terraform Resource
  - [ ] Implement fwprovider/resources/remotes/remote.go
  
- [ ] Phase 3: Data Sources
  - [ ] Create pbs_remote_stores data source
  - [ ] Create pbs_remote_namespaces data source
  - [ ] Create pbs_remote_groups data source
  
- [ ] Phase 4: Provider Integration
  - [ ] Register in fwprovider/provider.go
  
- [ ] Phase 5: Documentation
  - [ ] Create docs/resources/pbs_remote.md
  - [ ] Update sync job documentation
  
- [ ] Phase 6: Testing
  - [ ] Create test/integration/remotes_test.go
  - [ ] Implement all test scenarios

---

**Next Steps:** Begin with Phase 1 - Create the PBS API client for remotes.
