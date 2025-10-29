# PBS API Coverage Report

**Generated:** October 27, 2025  
**Provider Version:** Current Development  
**PBS API Version:** 4.0

## Executive Summary

This report analyzes the Proxmox Backup Server (PBS) Terraform provider's coverage of the PBS 4.0 API endpoints. The analysis is based on the official PBS API schema from `apidoc.js`.

### Coverage Statistics

- **Total API Endpoint Categories:** 18
- **Fully Implemented Categories:** 4
- **Partially Implemented Categories:** 4
- **Not Implemented Categories:** 10
- **Overall Coverage:** ~22%

---

## API Endpoint Coverage Matrix

### ✅ Fully Implemented (4 categories)

#### 1. `/config/datastore` - Datastore Configuration
**Status:** ✅ Complete  
**Provider Resource:** `pbs_datastore`  
**API Client:** `pbs/datastores/datastores.go`

| Endpoint | Method | Implemented | Notes |
|----------|--------|-------------|-------|
| `/config/datastore` | GET | ✅ | List datastores |
| `/config/datastore` | POST | ✅ | Create datastore |
| `/config/datastore/{name}` | GET | ✅ | Get datastore config |
| `/config/datastore/{name}` | PUT | ✅ | Update datastore |
| `/config/datastore/{name}` | DELETE | ✅ | Delete datastore |

**Features:**
- Full CRUD operations
- Optimistic locking with digest
- GC schedule, verification, tuning, maintenance mode
- Delete array support for clearing optional fields

---

#### 2. `/config/s3` (alias `/config/system/s3-endpoint`) - S3 Configuration
**Status:** ✅ Complete  
**Provider Resource:** `pbs_s3_endpoint`  
**API Client:** `pbs/endpoints/s3.go`

| Endpoint | Method | Implemented | Notes |
|----------|--------|-------------|-------|
| `/config/s3` | GET | ✅ | List S3 endpoints |
| `/config/s3` | POST | ✅ | Create S3 endpoint |
| `/config/s3/{id}` | GET | ✅ | Get S3 config |
| `/config/s3/{id}` | PUT | ✅ | Update S3 endpoint |
| `/config/s3/{id}` | DELETE | ✅ | Delete S3 endpoint |
| `/config/s3/{id}/scan/buckets` | GET | ✅ | List buckets |

**Features:**
- AWS, Backblaze B2, Scaleway support
- Provider quirks (skip-if-none-match-header)
- Rate limiting configuration
- Bucket scanning capability

---

#### 3. `/config/remote` - Remote Configuration
**Status:** ✅ Complete (Recently added)  
**Provider Resource:** `pbs_remote`  
**API Client:** `pbs/remotes/remotes.go`  
**Data Sources:** `pbs_remote_stores`, `pbs_remote_namespaces`, `pbs_remote_groups`

| Endpoint | Method | Implemented | Notes |
|----------|--------|-------------|-------|
| `/config/remote` | GET | ✅ | List remotes |
| `/config/remote` | POST | ✅ | Create remote |
| `/config/remote/{name}` | GET | ✅ | Get remote config |
| `/config/remote/{name}` | PUT | ✅ | Update remote |
| `/config/remote/{name}` | DELETE | ✅ | Delete remote |
| `/config/remote/{name}/scan/stores` | GET | ✅ | List remote stores |
| `/config/remote/{name}/scan/{store}/namespaces` | GET | ✅ | List namespaces |
| `/config/remote/{name}/scan/{store}/groups` | GET | ✅ | List groups |

**Features:**
- Full CRUD operations
- Password and fingerprint handling
- Discovery helpers (stores, namespaces, groups)
- Import capability with warnings

---

#### 4. `/config/metrics/*` - Metrics Configuration
**Status:** ✅ Complete  
**Provider Resource:** `pbs_metrics_server`  
**API Client:** `pbs/metrics/metrics.go`

| Endpoint | Method | Implemented | Notes |
|----------|--------|-------------|-------|
| `/config/metrics/influxdb-http` | GET | ✅ | List InfluxDB HTTP |
| `/config/metrics/influxdb-http` | POST | ✅ | Create InfluxDB HTTP |
| `/config/metrics/influxdb-http/{id}` | GET | ✅ | Get config |
| `/config/metrics/influxdb-http/{id}` | PUT | ✅ | Update config |
| `/config/metrics/influxdb-http/{id}` | DELETE | ✅ | Delete config |
| `/config/metrics/influxdb-udp` | GET | ✅ | List InfluxDB UDP |
| `/config/metrics/influxdb-udp` | POST | ✅ | Create InfluxDB UDP |
| `/config/metrics/influxdb-udp/{id}` | GET | ✅ | Get config |
| `/config/metrics/influxdb-udp/{id}` | PUT | ✅ | Update config |
| `/config/metrics/influxdb-udp/{id}` | DELETE | ✅ | Delete config |

**Features:**
- Both HTTP and UDP protocols
- Organization, bucket, token configuration
- TLS verification options

---

### 🔶 Partially Implemented (4 categories)

#### 5. `/config/notifications/*` - Notification Configuration
**Status:** 🔶 Partial (Endpoints: 80%, Matchers: 100%)  
**Provider Resources:** `pbs_smtp_notification`, `pbs_gotify_notification`, `pbs_sendmail_notification`, `pbs_webhook_notification`, `pbs_notification_matcher`  
**API Client:** `pbs/notifications/notifications.go`

##### Implemented:
| Endpoint | Method | Status |
|----------|--------|--------|
| `/config/notifications/endpoints/smtp` | GET/POST | ✅ |
| `/config/notifications/endpoints/smtp/{name}` | GET/PUT/DELETE | ✅ |
| `/config/notifications/endpoints/gotify` | GET/POST | ✅ |
| `/config/notifications/endpoints/gotify/{name}` | GET/PUT/DELETE | ✅ |
| `/config/notifications/endpoints/sendmail` | GET/POST | ✅ |
| `/config/notifications/endpoints/sendmail/{name}` | GET/PUT/DELETE | ✅ |
| `/config/notifications/endpoints/webhook` | GET/POST | ✅ |
| `/config/notifications/endpoints/webhook/{name}` | GET/PUT/DELETE | ✅ |
| `/config/notifications/matchers` | GET/POST | ✅ |
| `/config/notifications/matchers/{name}` | GET/PUT/DELETE | ✅ |

##### Missing:
| Endpoint | Impact |
|----------|--------|
| `/config/notifications/targets/test` | Low - Test notification sending |

**Features:**
- All 4 notification endpoint types
- Matcher routing with targets and filters
- Base64 secrets and headers
- Calendar event matching

---

#### 6. `/config/prune`, `/config/sync`, `/config/verify` - Job Configuration
**Status:** 🔶 Partial (3/4 job types)  
**Provider Resources:** `pbs_prune_job`, `pbs_sync_job`, `pbs_verify_job`  
**API Client:** `pbs/jobs/jobs.go`

##### Implemented:
| Endpoint | Method | Status |
|----------|--------|--------|
| `/config/prune` | GET/POST | ✅ |
| `/config/prune/{id}` | GET/PUT/DELETE | ✅ |
| `/config/sync` | GET/POST | ✅ |
| `/config/sync/{id}` | GET/PUT/DELETE | ✅ |
| `/config/verify` | GET/POST | ✅ |
| `/config/verify/{id}` | GET/PUT/DELETE | ✅ |

##### Missing:
| Endpoint | Impact |
|----------|--------|
| `/config/traffic-control` | Medium - Network bandwidth control |
| `/tape/backup-job/*` | Medium - Tape backup jobs |

**Features:**
- Schedule configuration (calendar events)
- Namespace filtering (ns, max-depth)
- Comment and disable flags
- Delete array for clearing optional fields

---

#### 7. `/admin/datastore/{store}` - Datastore Admin Operations
**Status:** 🔶 Partial (Read-only data appropriate for Terraform)  
**Implementation:** Via datastore resource

##### Missing Terraform-Appropriate Operations:
| Endpoint | Impact | Terraform Appropriate? |
|----------|--------|----------------------|
| `/admin/datastore/{store}/namespace` | High - Namespace management | ✅ YES - State-based CRUD |
| `/admin/datastore/{store}/snapshots` | Medium - Snapshot browsing | ✅ YES - Read-only data source |
| `/admin/datastore/{store}/status` | Medium - Status info | ✅ YES - Read-only data source |
| `/admin/datastore/{store}/notes` | Low - Backup notes | ✅ YES - State-based metadata |

##### Imperative Operations (Out of Scope for Terraform):
| Endpoint | Why Out of Scope |
|----------|-----------------|
| `/admin/datastore/{store}/gc` POST | Imperative action - use CLI/scheduled jobs |
| `/admin/datastore/{store}/verify` POST | Imperative action - use CLI/scheduled jobs |
| `/admin/datastore/{store}/prune` POST | Imperative action - use CLI/scheduled jobs |
| `/admin/datastore/{store}/download` | File transfer operation - use CLI |
| `/admin/datastore/{store}/upload-backup-log` | File upload operation - use CLI |

**Note:** The provider already supports declarative scheduling via `pbs_prune_job`, `pbs_verify_job`, and datastore GC schedule configuration. Manual triggers should be performed using `proxmox-backup-manager` CLI or API directly.

---

#### 8. `/admin/sync` - Sync Operations
**Status:** ✅ Complete (via scheduled jobs)

The provider supports sync operations through the `pbs_sync_job` resource, which is the appropriate Terraform approach. Manual sync triggers are imperative operations better suited to CLI tools.

| Endpoint | Terraform Support |
|----------|------------------|
| `/admin/sync` POST | ❌ Imperative - use CLI or `pbs_sync_job` |
| `/admin/sync/{upid}/status` GET | 🔶 Could add as data source for monitoring |

---

### ❌ Not Implemented (10 categories)

#### 9. `/access/*` - Access Control
**Status:** ❌ Not Implemented  
**Impact:** HIGH

##### Critical Missing Endpoints:
| Endpoint Path | Purpose | Impact |
|---------------|---------|--------|
| `/access/users` | User management | High |
| `/access/users/{userid}/token` | API token management | High |
| `/access/domains` | Authentication realms | High |
| `/access/acl` | ACL management | High |
| `/access/roles` | Role definitions | Medium |
| `/access/permissions` | Permission queries | Medium |
| `/access/tfa` | Two-factor auth | Medium |
| `/access/ticket` | Auth tickets | Low (handled by client) |

**Recommendation:** High priority for v2.0 - Essential for multi-user environments

---

#### 10. `/config/acme/*` - ACME/Let's Encrypt
**Status:** ❌ Not Implemented  
**Impact:** MEDIUM

##### Missing Endpoints:
| Endpoint Path | Purpose |
|---------------|---------|
| `/config/acme/account` | ACME account management |
| `/config/acme/plugins` | DNS challenge plugins |
| `/config/acme/challenge-schema` | Plugin schemas |

**Recommendation:** Medium priority - Useful for certificate automation

---

#### 11. `/config/tape/*` - Tape Library Configuration
**Status:** ❌ Not Implemented  
**Impact:** MEDIUM (niche use case)

##### Missing Endpoints:
| Endpoint Path | Purpose |
|---------------|---------|
| `/config/tape/changer` | Tape changer config |
| `/config/tape/drive` | Tape drive config |
| `/config/media-pool` | Media pool config |
| `/config/tape/encryption-key` | Tape encryption |

**Recommendation:** Low-Medium priority - Niche enterprise feature

---

#### 12. `/tape/*` - Tape Operations
**Status:** ❌ Not Implemented  
**Impact:** MEDIUM (niche use case)

##### Missing Endpoints:
| Endpoint Path | Purpose |
|---------------|---------|
| `/tape/backup` | Tape backup jobs |
| `/tape/restore` | Tape restore ops |
| `/tape/media` | Media management |
| `/tape/drive/{drive}/*` | Drive operations |
| `/tape/changer/{name}/*` | Changer operations |

**Recommendation:** Low priority - Requires tape hardware

---

#### 13. `/nodes/{node}/*` - Node Management
**Status:** ❌ Not Implemented  
**Impact:** MEDIUM-HIGH

##### Critical Missing Endpoints:
| Endpoint Path | Purpose | Impact |
|---------------|---------|--------|
| `/nodes/{node}/status` | Node status | Medium |
| `/nodes/{node}/services` | Service management | Medium |
| `/nodes/{node}/apt` | Package management | Medium |
| `/nodes/{node}/certificates` | Certificate mgmt | Medium |
| `/nodes/{node}/disks` | Disk management | Medium |
| `/nodes/{node}/network` | Network config | Low |
| `/nodes/{node}/subscription` | Subscription mgmt | Low |
| `/nodes/{node}/syslog` | System logs | Low |
| `/nodes/{node}/tasks` | Task management | High |
| `/nodes/{node}/time` | Time config | Low |

**Recommendation:** Medium-High priority for v2.0 - Important for automation

---

#### 14. `/nodes/{node}/tasks/*` - Task Management
**Status:** ❌ Not Implemented  
**Impact:** HIGH

##### Missing Endpoints:
| Endpoint Path | Purpose | Impact |
|---------------|---------|--------|
| `/nodes/{node}/tasks` | List tasks | High |
| `/nodes/{node}/tasks/{upid}/status` | Task status | High |
| `/nodes/{node}/tasks/{upid}/log` | Task logs | High |
| `/nodes/{node}/tasks/{upid}` | Stop task | Medium |

**Recommendation:** High priority - Essential for monitoring long-running ops

---

#### 15. `/admin/backup` - Manual Backup Operations
**Status:** ❌ Out of Scope (Imperative Operations)  
**Impact:** N/A

These endpoints trigger one-time backup operations, which are imperative actions inappropriate for Terraform's declarative model.

##### Endpoints:
| Endpoint Path | Purpose | Terraform Appropriate? |
|---------------|---------|----------------------|
| `/admin/datastore/{store}/gc` | POST - Trigger GC | ❌ NO - Use CLI or scheduled job |
| `/admin/datastore/{store}/verify` | POST - Trigger verify | ❌ NO - Use CLI or scheduled job |
| `/admin/datastore/{store}/prune` | POST - Trigger prune | ❌ NO - Use CLI or scheduled job |
| `/admin/sync` | POST - Trigger sync | ❌ NO - Use CLI or scheduled job |

**Recommendation:** Not applicable - Users should use `proxmox-backup-manager` CLI for ad-hoc operations, and the provider's scheduled job resources (`pbs_prune_job`, `pbs_sync_job`, `pbs_verify_job`) for regular operations.

---

#### 16. `/backup/*` (HTTP/2 API) - Backup Client API
**Status:** ❌ Not Implemented  
**Impact:** LOW (client-side)

This is the backup client protocol (proxmox-backup-client), not typically needed for infrastructure provisioning.

**Recommendation:** Very low priority - Out of scope for Terraform provider

---

#### 17. `/config/access/*` - Additional Access Config
**Status:** ❌ Not Implemented  
**Impact:** MEDIUM

##### Missing Endpoints:
| Endpoint Path | Purpose |
|---------------|---------|
| `/config/access/ad` | Active Directory |
| `/config/access/ldap` | LDAP config |
| `/config/access/openid` | OpenID Connect |
| `/config/access/webauthn` | WebAuthn |

**Recommendation:** Medium priority - Important for enterprise auth

---

#### 18. Root Endpoints - Version & Status
**Status:** 🔶 Partial (no provider resources)

##### Available but not exposed:
| Endpoint Path | Purpose | Exposed |
|---------------|---------|---------|
| `/version` | API version | ❌ |
| `/ping` | Health check | ❌ |

These are typically used by the API client, not exposed as Terraform resources.

---

## Feature Completeness by Category

### Configuration Management
| Feature | Coverage | Priority |
|---------|----------|----------|
| Datastores | 100% ✅ | - |
| S3 Endpoints | 100% ✅ | - |
| Remotes | 100% ✅ | - |
| Metrics | 100% ✅ | - |
| Notifications (Endpoints) | 100% ✅ | - |
| Notifications (Matchers) | 100% ✅ | - |
| Jobs (Prune/Sync/Verify) | 100% ✅ | - |
| Jobs (Tape) | 0% ❌ | Low |
| ACME | 0% ❌ | Medium |
| Tape Library | 0% ❌ | Low |
| Access Control | 0% ❌ | High |
| Authentication Realms | 0% ❌ | Medium |

### Operations & Monitoring
| Feature | Coverage | Priority | Terraform Appropriate? |
|---------|----------|----------|----------------------|
| Manual GC | N/A | - | ❌ NO - Imperative operation |
| Manual Verify | N/A | - | ❌ NO - Imperative operation |
| Manual Prune | N/A | - | ❌ NO - Imperative operation |
| Manual Sync | N/A | - | ❌ NO - Imperative operation |
| Task Monitoring (read) | 0% ❌ | High | ✅ YES - Read-only data source |
| Datastore Status (read) | 0% ❌ | Medium | ✅ YES - Read-only data source |
| Snapshot Browsing (read) | 0% ❌ | Medium | ✅ YES - Read-only data source |
| Node Status (read) | 0% ❌ | Medium | ✅ YES - Read-only data source |
| Service Configuration | 0% ❌ | Low | ✅ YES - Enabled/disabled state |
| Package Management | N/A | - | ❌ NO - Imperative operation |

### Infrastructure Management
| Feature | Coverage | Priority |
|---------|----------|----------|
| Node Configuration | 0% ❌ | Medium |
| Network Configuration | 0% ❌ | Low |
| Disk Management | 0% ❌ | Medium |
| Certificate Management | 0% ❌ | Medium |
| Subscription Management | 0% ❌ | Low |

---

## Implementation Recommendations

### Phase 1: High Priority (v2.0 Target)

#### 1.1 Access Control & Security
- **Users:** `pbs_user` resource
- **Tokens:** `pbs_api_token` resource (possibly nested under user)
- **ACLs:** `pbs_acl` resource
- **Realms:** `pbs_realm` resource (AD, LDAP, OpenID)

**Justification:** Essential for production multi-user deployments. Fully declarative - manages "who should have access" rather than "grant access now".

**API Endpoints:**
- `/access/users` (CRUD)
- `/access/users/{userid}/token` (CRUD)
- `/access/acl` (CRUD)
- `/access/domains` (CRUD)

#### 1.2 Namespace Management
- **Namespaces:** `pbs_namespace` resource
- **Integration:** Extend existing resources to support namespace references

**Justification:** Essential for organizing backups in multi-tenant scenarios. Pure state-based resource.

**API Endpoints:**
- `/admin/datastore/{store}/namespace` (CRUD)

#### 1.3 Read-Only Data Sources
- **Task Status:** `pbs_task` data source for monitoring scheduled jobs
- **Datastore Status:** `pbs_datastore_status` data source
- **Snapshot Info:** `pbs_snapshots` data source for backup browsing
- **Node Status:** `pbs_node` data source

**Justification:** Enables monitoring and conditional logic without state modification. Pure read operations.

**API Endpoints:**
- `/nodes/{node}/tasks` (GET)
- `/admin/datastore/{store}/status` (GET)
- `/admin/datastore/{store}/snapshots` (GET)
- `/nodes/{node}/status` (GET)

### Phase 2: Medium Priority (v2.1+ Target)

#### 2.1 Certificate & ACME Management
- **Certificates:** `pbs_certificate` resource
- **ACME Accounts:** `pbs_acme_account` resource
- **ACME Plugins:** `pbs_acme_plugin` resource (DNS challenges)

**Justification:** Automate certificate lifecycle management. Fully declarative state.

**API Endpoints:**
- `/nodes/{node}/certificates` (CRUD)
- `/config/acme/account` (CRUD)
- `/config/acme/plugins` (CRUD)

#### 2.2 Service Management
- **Services:** `pbs_service` resource
- **Service State:** Manage enabled/disabled state of PBS services

**Justification:** Declarative service configuration (e.g., "API service should be enabled").

**API Endpoints:**
- `/nodes/{node}/services` (CRUD)

**Note:** Service start/stop/restart operations are imperative and out of scope.

#### 2.3 Enhanced Data Sources
- **Available Updates:** `pbs_updates` data source
- **System Information:** `pbs_system_info` data source
- **Subscription Status:** `pbs_subscription` data source

**Justification:** Visibility into system state for conditional logic and monitoring.

**API Endpoints:**
- `/nodes/{node}/apt/update` (GET)
- `/nodes/{node}/status` (GET)
- `/nodes/{node}/subscription` (GET)

### Phase 3: Low Priority (Future)

#### 3.1 Tape Library (Conditional)
Only if there's significant user demand. All configuration-based, no imperative operations:
- `pbs_tape_changer` resource (changer configuration)
- `pbs_tape_drive` resource (drive configuration)
- `pbs_media_pool` resource (pool configuration)
- `pbs_tape_backup_job` resource (scheduled tape backup jobs)
- `pbs_tape_media` data source (read-only media inventory)

**Justification:** Niche enterprise feature, requires specialized hardware. All resources are declarative state management.

**API Endpoints:**
- `/config/tape/changer` (CRUD)
- `/config/tape/drive` (CRUD)
- `/config/media-pool` (CRUD)
- `/tape/media/list` (GET)

**Out of Scope:** Manual tape operations (load, eject, etc.) remain CLI/API operations.

#### 3.2 Network & Storage Configuration
- **Traffic Control:** `pbs_traffic_control` resource (network QoS rules)
- **Network Interfaces:** `pbs_network_interface` resource (if not managed by OS tooling)

**Justification:** Declarative network configuration. Users may prefer OS-level tools (Ansible, cloud-init).

**API Endpoints:**
- `/config/traffic-control` (CRUD)
- `/nodes/{node}/network` (CRUD)

**Note:** Disk management (formatting, mounting) is typically handled outside Terraform.

---

## Technical Debt & Improvements

### Current Implementation Gaps

1. **No Data Sources for Existing Resources**
   - Many resources lack corresponding data sources
   - Makes it hard to reference or discover existing infrastructure
   - Prevents data-driven conditional logic

2. **Limited Import Support**
   - Some resources lack import capability
   - Remote import has password limitation warning
   - Import IDs not consistently documented

3. **No Task Status Visibility**
   - Long-running operations don't expose UPID or status
   - Provider waits silently without progress indication
   - Can't monitor scheduled job execution results

4. **Missing Namespace Support**
   - Namespaces not exposed as first-class resources
   - Existing resources don't support namespace filtering
   - Can't manage namespace hierarchy declaratively

### Recommended Improvements

#### 1. Add Data Sources for All Resources
```hcl
# Examples of needed data sources:
data "pbs_datastore" "existing" {
  name = "backups"
}

data "pbs_s3_endpoint" "s3" {
  id = "aws-backup"
}

data "pbs_remote" "offsite" {
  name = "remote-pbs"
}

data "pbs_prune_job" "daily" {
  id = "daily-prune"
}

# NEW: Status and monitoring data sources
data "pbs_datastore_status" "backups" {
  name = "backups"
}

data "pbs_task" "recent_gc" {
  store = "backups"
  type  = "garbage-collection"
  limit = 5
}

data "pbs_snapshots" "vm100" {
  store     = "backups"
  backup_id = "vm/100"
}
```

#### 2. Enhance Import Capability
- Document import ID format for every resource
- Add validation during import
- Handle sensitive fields (passwords, tokens) gracefully
- Add import examples to documentation

#### 3. Improve Task Visibility (Read-Only)
- Add computed `last_run_upid` field to job resources
- Create `pbs_task` data source for monitoring
- Better error messages with UPID reference when operations fail
- Optional `timeout` parameter for long-running ops

#### 4. Documentation & Examples
- Add complete examples for all resources
- Document which API endpoints map to which resources
- Add troubleshooting guide for common issues
- Create integration examples (e.g., complete backup setup)

#### 5. Namespace Support
```hcl
# NEW: Namespace resource
resource "pbs_namespace" "prod" {
  datastore = pbs_datastore.backups.name
  namespace = "prod"
  comment   = "Production backups"
}

resource "pbs_namespace" "prod_vms" {
  datastore = pbs_datastore.backups.name
  namespace = "prod/vms"
  comment   = "Production VM backups"
}

# Enhanced: Add namespace support to existing resources
resource "pbs_prune_job" "prod_prune" {
  store     = pbs_datastore.backups.name
  namespace = pbs_namespace.prod.namespace
  schedule  = "daily"
  # ...
}
```

---

## API Pattern Analysis

### Common Patterns Observed

1. **Optimistic Locking**
   - Most update endpoints require `digest` parameter
   - Prevents concurrent modifications
   - Provider handles this correctly

2. **Delete Arrays**
   - Clearing optional fields requires `delete` array
   - Provider implements this for jobs and notifications
   - Should extend to all resources

3. **Async Operations**
   - Many operations return UPID (task ID)
   - Require polling `/nodes/{node}/tasks/{upid}/status`
   - Provider uses WaitForTask helper

4. **Namespace Support**
   - Many endpoints accept optional `ns` parameter
   - Pattern: `ns` string + optional `max-depth` integer
   - Not fully exposed in provider resources

5. **Calendar Events**
   - Schedule format uses PBS calendar syntax
   - Provider passes through as string
   - No validation at provider level

### Lessons for Future Implementation

1. **Consistent Resource Patterns**
   - All resources should support import
   - All should handle digest for updates
   - All should use delete arrays where applicable

2. **Data Source Parity**
   - Every resource should have a corresponding data source
   - Enables cross-referencing and discovery

3. **Namespace First-Class**
   - Treat namespaces as top-level resources
   - Add namespace support to existing resources retroactively

4. **Task Awareness**
   - Resources that trigger async ops should optionally expose UPID
   - Consider adding task data source for monitoring

---

## Conclusion

The Proxmox Backup Server Terraform provider currently provides **solid coverage of core configuration management** (datastores, jobs, notifications, remotes, metrics, S3). However, significant gaps exist in:

1. **Access Control & Security** (0% coverage)
2. **Operational Tasks** (manual operations, task tracking)
3. **Node Management** (status, services, certificates)
4. **Namespace Administration** (partial coverage)

### Strengths
✅ Complete CRUD for all implemented resources  
✅ Proper async task handling  
✅ Good error handling and validation  
✅ Recent additions (remotes) show strong implementation patterns  

### Weaknesses
❌ No multi-user/multi-tenant support (no access control)  
❌ No read-only data sources for monitoring and discovery  
❌ No task status visibility for job execution  
❌ Missing namespace management as first-class resource  

### Recommended Next Steps

**Immediate (v1.x maintenance):**
1. Add data sources for existing resources (datastores, jobs, endpoints)
2. Improve documentation with complete examples
3. Enhance import capabilities and documentation
4. Add computed fields for task tracking where appropriate

**Short-term (v2.0):**
1. **Access Control** - Users, API tokens, ACLs, authentication realms
2. **Namespace Management** - First-class namespace resource with CRUD
3. **Read-Only Monitoring** - Task, status, and snapshot data sources
4. **Enhanced Discovery** - Data sources for all existing resources

**Long-term (v2.1+):**
1. **Certificate Management** - ACME accounts, plugins, certificate resources
2. **Service Management** - Declarative service configuration
3. **Network Configuration** - Traffic control and network resources (if demand exists)
4. **Tape Library** - Configuration resources (if demand exists)

### Out of Scope (Imperative Operations)

The following operations are **explicitly out of scope** for this Terraform provider as they are imperative actions rather than declarative state:

- ❌ Manual GC/verify/prune triggers
- ❌ Service start/stop/restart actions
- ❌ Tape media load/eject operations
- ❌ Package installation/updates
- ❌ Network interface up/down actions
- ❌ File upload/download operations

**Rationale:** Terraform is designed for declarative infrastructure state management. One-time actions should be performed using:
- `proxmox-backup-manager` CLI for administrative tasks
- Direct API calls for scripting
- Ansible or similar tools for imperative orchestration
- Scheduled jobs (already supported via `pbs_*_job` resources) for regular operations

This roadmap would bring the provider from **~22% to ~60-70% API coverage**, focusing exclusively on **declarative, state-based features** that align with Terraform's design philosophy.
