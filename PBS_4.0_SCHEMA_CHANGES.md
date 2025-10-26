# PBS 4.0 API Schema Changes

This document details all API schema changes between PBS 3.x and PBS 4.0 that affect the Terraform provider.

## Summary

PBS 4.0 not only changed API paths but also modified parameter schemas for several resources. This requires updates to resource definitions, API clients, and integration tests.

## API Path Changes

### 1. Metrics API
- **Old**: `/config/metrics/server/{type}/{name}`
- **New**: `/config/metrics/{type}/{name}`
- **Change**: Removed `/server` segment from path

### 2. Notifications API
- **Old**: `/config/notifications/targets/{type}/{name}`
- **New**: `/config/notifications/endpoints/{type}/{name}`
- **Change**: Changed `targets` to `endpoints`

### 3. Verification Jobs API
- **Old**: `/config/verification`
- **New**: `/config/verify`
- **Change**: Shortened path

### 4. Garbage Collection Jobs API
- **Old**: `/config/garbage-collection`
- **New**: **REMOVED** - GC is now configured at datastore level
- **Change**: Endpoint no longer exists

## Schema Changes

### Metrics - InfluxDB HTTP

**Old Schema (PBS 3.x):**
```json
{
  "name": "string",
  "type": "influxdb-http",
  "server": "hostname",
  "port": 443,
  "organization": "string",
  "bucket": "string",
  "token": "string",
  "timeout": 5,
  "verify_certificate": true,
  "enable": true,
  "comment": "string"
}
```

**New Schema (PBS 4.0):**
```json
{
  "name": "string",
  "type": "influxdb-http",
  "url": "https://hostname:port",  // Combined server+port into single URL
  "organization": "string",
  "bucket": "string",
  "token": "string",
  "max-body-size": 25000000,       // New field
  "verify-tls": true,              // Renamed from verify_certificate
  "enable": true,
  "comment": "string"
}
```

**Changes:**
- ✅ `server` + `port` → `url` (single field)
- ✅ `verify_certificate` → `verify-tls`
- ✅ `timeout` → Removed
- ✅ Added `max-body-size`

**References:**
- Schema: `pbs-api-types/src/metrics.rs` lines 98-142
- API: `src/api2/config/metrics/influxdbhttp.rs`
- Frontend: `www/window/InfluxDbEdit.js`

### Metrics - InfluxDB UDP

**Old Schema (PBS 3.x):**
```json
{
  "name": "string",
  "type": "influxdb-udp",
  "server": "hostname",
  "port": 8089,
  "proto": "udp",
  "mtu": 1500,
  "enable": true,
  "comment": "string"
}
```

**New Schema (PBS 4.0):**
```json
{
  "name": "string",
  "type": "influxdb-udp",
  "host": "hostname:port",  // Combined server+port with colon separator
  "mtu": 1500,
  "enable": true,
  "comment": "string"
}
```

**Changes:**
- ✅ `server` + `port` → `host` (format: `hostname:port`)
- ✅ `proto` → Removed (always UDP)

**References:**
- Schema: `pbs-api-types/src/metrics.rs` lines 38-74
- API: `src/api2/config/metrics/influxdbudp.rs`
- Frontend: `www/window/InfluxDbEdit.js` lines 118-220

### Notifications - SMTP

**Old Schema (PBS 3.x):**
```json
{
  "name": "string",
  "server": "smtp.example.com",
  "port": 587,
  "from_address": "sender@example.com",
  "mailto": "recipient@example.com",  // String
  "comment": "string"
}
```

**New Schema (PBS 4.0):**
```json
{
  "name": "string",
  "server": "smtp.example.com",
  "port": 587,
  "from-address": "sender@example.com",
  "mailto": ["recipient@example.com"],  // Array of strings
  "comment": "string"
}
```

**Changes:**
- ✅ `from_address` → `from-address` (kebab-case)
- ✅ `mailto`: String → Array of strings

**Error Message:**
```
parameter verification failed - 'mailto': Expected array - got scalar value.
```

### Notifications - Sendmail

**Same changes as SMTP:**
- ✅ `mailto`: String → Array of strings

### Notifications - Webhook

**Old Schema:**
```json
{
  "method": "POST"
}
```

**New Schema:**
```json
{
  "method": "post"  // Lowercase enum value
}
```

**Changes:**
- ✅ `method` enum values changed to lowercase

**Error Message:**
```
parameter verification failed - 'method': value 'POST' is not defined in the enumeration.
```

### Jobs - Verify

PBS 4.0 keeps the existing core fields for verify jobs but adds optimistic locking support via digests and tightens validation around optional attributes.

**PBS 3.x Payload (conceptual):**
```json
{
   "store": "datastore-name",
   "schedule": "daily",
   "ignore-verified": false,
   "outdated-after": 30,
   "max-depth": 2,
   "disable": false,
   "comment": "string"
}
```

**PBS 4.0 Payload:**
```json
{
   "store": "datastore-name",
   "schedule": "daily",
   "ignore-verified": false,
   "outdated-after": 30,
   "max-depth": 2,
   "disable": false,
   "ns": "namespace/child",
   "comment": "string",
   "digest": "6f1f2..."
}
```

**Changes:**
- ✅ Digests are returned for every read and must be echoed back on update/delete (optimistic locking).
- ✅ `ns` continues to be the canonical field name for namespaces; the provider now validates/normalizes it through helper utilities.
- ✅ `disable` remains supported (contrary to early assumptions) and is surfaced as an optional boolean.
- ✅ Added provider-side helpers to compute delete lists when optional attributes are cleared.

**Notes:**
- PBS rejects `outdated-after` when `ignore-verified` is false; provider leaves validation to API and documents the relationship.
- Clearing optional fields requires sending the field name in the `delete` array; helpers now compute this automatically.

### Jobs - Prune

**Key Differences:**
- ✅ The legacy `backup-type` parameter was removed; provider now maps resource arguments to the supported keep-set fields only.
- ✅ Namespace values must satisfy the PBS namespace regex; the Terraform schema enforces this via validators and helper conversion.
- ✅ Digests are returned per job and are required when issuing update/delete calls; provider exposes the computed digest field.

**Error Messages Observed When Migrating:**
- `'backup-type': schema does not allow additional properties`
- `'ns': value does not match the regex pattern`

### Jobs - Sync

PBS 4.0 introduces optional namespace scoping and stricter validation for filter and throttle fields.

**Key Differences:**
- ✅ `remote-ns` (surfaced as `remote_namespace`) lets you pull from a specific namespace on the remote PBS.
- ✅ `ns` (surfaced as `namespace`) continues to scope the local target; provider enforces the PBS namespace regex.
- ✅ Bandwidth limit fields (`rate-in`, `rate-out`, `burst-in`, `burst-out`) must be encoded as strings using PBS byte-size syntax (e.g., `"10M"`).
- ✅ `group-filter` now expects `<type>/<id>[/<namespace>]` expressions; provider validates and passes lists directly.
- ✅ Digests are returned and required for update/delete; provider exposes a computed `digest` attribute and handles optimistic locking.
- ✅ `disable` is still supported; provider keeps it as an optional boolean and participates in delete tracking when unset.

**Implementation Notes:**
- Helper functions convert Terraform optional values to pointers, build delete arrays, and normalize group filters.
- Terraform schema validators enforce non-negative integers for `max_depth`/`transfer_last` and regex-match the group filter strings.

### Jobs - Garbage Collection

**Status:** Endpoint completely removed in PBS 4.0

**Error Message:**
```
Path '/api2/json/config/garbage-collection' not found.
```

**Resolution:** GC configuration moved to datastore-level settings. The GC job resource needs to be either:
1. Removed from provider (breaking change)
2. Mapped to new datastore GC configuration endpoint
3. Deprecated with migration guide

## Implementation Plan

### Phase 1: Metrics Resources
1. Update `fwprovider/resources/metrics/metrics_server.go`:
   - Add `url` field for HTTP type
   - Add `host` field for UDP type
   - Add `verify_tls` field
   - Add `max_body_size` field
   - Remove `timeout` field
   - Keep `server` and `port` in schema for backwards compatibility, construct `url`/`host` in Create/Update

2. Update `pbs/metrics/metrics.go` API client:
   - Map between Terraform schema and PBS API schema
   - HTTP: Build `url` from `server` + `port` or accept `url` directly
   - UDP: Build `host` from `server` + `port` or accept `host` directly

3. Update tests in `test/integration/metrics_test.go`:
   - Test both old field format and new format
   - Verify API accepts constructed values

### Phase 2: Notification Resources
1. Update SMTP resource `fwprovider/resources/notifications/smtp_notification.go`:
   - Change `mailto` attribute to ListAttribute
   - Convert single value to array for API

2. Update Sendmail resource `fwprovider/resources/notifications/sendmail_notification.go`:
   - Same `mailto` change

3. Update Webhook resource `fwprovider/resources/notifications/webhook_notification.go`:
   - Lowercase `method` enum values

4. Update `pbs/notifications/notifications.go` API client:
   - Handle array conversion for mailto

5. Update tests in `test/integration/notifications_test.go`:
   - Use array format for mailto
   - Use lowercase method

### Phase 3: Job Resources
1. Update verify job `fwprovider/resources/jobs/verify_job.go`:
   - Surface digest attribute and pass it through update/delete
   - Preserve `disable` as optional bool and wire helpers for delete tracking
   - Add schema validators for `outdated_after`/`max_depth` and namespace helper usage

2. Update prune job `fwprovider/resources/jobs/prune_job.go`:
   - Validate namespace format and support digest/delete handling
   - Ensure helper conversions cover optional numeric and boolean fields

3. Update sync job `fwprovider/resources/jobs/sync_job.go`:
   - Support `remote_namespace` and namespace validators
   - Require string-based rate/burst values and enforce group filter regex
   - Keep `disable` optional while honoring digest/delete semantics

4. Handle GC job `fwprovider/resources/jobs/gc_job.go`:
   - Add deprecation notice
   - Document that GC is now configured at datastore level

5. Update `pbs/jobs/jobs.go` API client:
   - Add pointer helpers/digest handling for verify, prune, and sync jobs
   - Normalize optional values and propagate delete lists

6. Update tests in `test/integration/jobs_test.go`:
   - Cover new fields (remote namespace, digest expectations, namespace validation)
   - Ensure rate limits are strings and group filters follow PBS regex

### Phase 4: Integration Testing
1. Run full test suite against PBS 4.0
2. Verify backwards compatibility where possible
3. Document breaking changes in CHANGELOG

## Breaking Changes

### Required User Actions

1. **Metrics - InfluxDB HTTP:**
   - Users must update configurations to use `url` instead of `server`/`port`
   - Update `verify_certificate` to `verify_tls`
   - OR: Provider can auto-convert for backwards compatibility

2. **Metrics - InfluxDB UDP:**
   - Users must update configurations to use `host` instead of `server`/`port`
   - OR: Provider can auto-convert for backwards compatibility

3. **Notifications - SMTP/Sendmail:**
   - `mailto` must be a list: `mailto = ["user@example.com"]`
   - Was: `mailto = "user@example.com"`

4. **Notifications - Webhook:**
   - `method` must be lowercase: `method = "post"`
   - Was: `method = "POST"`

5. **Jobs - All types:**
   - Ensure rate/burst limits are strings (`"10M"`, not numbers)
   - Update group filters to `<type>/<id>[/<namespace>]` format
   - Verify namespace values comply with PBS regex expectations

6. **Jobs - GC:**
   - Resource deprecated - configure GC at datastore level instead

## Testing Checklist

- [ ] Metrics InfluxDB HTTP create/read/update/delete
- [ ] Metrics InfluxDB UDP create/read/update/delete
- [x] SMTP notification with array mailto
- [x] Sendmail notification with array mailto
- [x] Webhook notification with lowercase method
- [x] Gotify notification (already passing)
- [x] Verification job digest/disable regression
- [x] Prune job namespace validator coverage
- [x] Sync job namespace + rate limit coverage
- [ ] GC job deprecation warning
- [ ] All S3 tests (already passing)
- [ ] Backwards compatibility tests

## References

- PBS 4.0 Source: https://github.com/proxmox/proxmox-backup
- Metrics API Types: `pbs-api-types/src/metrics.rs`
- Metrics HTTP API: `src/api2/config/metrics/influxdbhttp.rs`
- Metrics UDP API: `src/api2/config/metrics/influxdbudp.rs`
- Frontend UI: `www/window/InfluxDbEdit.js`, `www/config/MetricServerView.js`
