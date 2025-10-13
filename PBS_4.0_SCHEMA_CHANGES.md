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

### Jobs - Verification

**Old Schema:**
```json
{
  "store": "datastore-name",
  "schedule": "daily",
  "disable": false,  // Had disable field
  "comment": "string"
}
```

**New Schema:**
```json
{
  "store": "datastore-name",
  "schedule": "daily",
  // No disable field
  "comment": "string"
}
```

**Changes:**
- ✅ `disable` field removed (not allowed in schema)

**Error Message:**
```
parameter verification failed - 'disable': schema does not allow additional properties
```

### Jobs - Prune

**Schema Issues:**
- ✅ `backup-type` field not allowed in schema
- ✅ `ns` value doesn't match regex pattern (namespace format changed)

**Error Messages:**
```
- 'backup-type': schema does not allow additional properties
- 'ns': value does not match the regex pattern
```

### Jobs - Sync

**Schema Issues:**
- ✅ `disable` field not allowed
- ✅ `rate-in`: Expected string value (not number)
- ✅ `group-filter[0]`: Format changed to `<group:GROUP||type:<vm|ct|host>|regex:REGEX>`

**Error Messages:**
```
- 'disable': schema does not allow additional properties
- 'group-filter/[0]': input doesn't match expected format '<group:GROUP||type:<vm|ct|host>|regex:REGEX>'
- 'rate-in': Expected string value.
```

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
1. Update verification job `fwprovider/resources/jobs/verify_job.go`:
   - Remove `disable` attribute

2. Update prune job `fwprovider/resources/jobs/prune_job.go`:
   - Remove `backup_type` attribute
   - Fix `ns` attribute format validation

3. Update sync job `fwprovider/resources/jobs/sync_job.go`:
   - Remove `disable` attribute
   - Change `rate_in` to string type
   - Fix `group_filter` format validation

4. Handle GC job `fwprovider/resources/jobs/gc_job.go`:
   - Add deprecation notice
   - Document that GC is now configured at datastore level

5. Update `pbs/jobs/jobs.go` API client:
   - Remove references to removed fields
   - Update validation logic

6. Update tests in `test/integration/jobs_test.go`:
   - Remove usage of removed fields
   - Update field formats

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
   - Remove `disable` attribute (not supported in PBS 4.0)
   - Use resource lifecycle instead

6. **Jobs - GC:**
   - Resource deprecated - configure GC at datastore level instead

## Testing Checklist

- [ ] Metrics InfluxDB HTTP create/read/update/delete
- [ ] Metrics InfluxDB UDP create/read/update/delete
- [ ] SMTP notification with array mailto
- [ ] Sendmail notification with array mailto
- [ ] Webhook notification with lowercase method
- [ ] Gotify notification (already passing)
- [ ] Verification job without disable field
- [ ] Prune job with correct field formats
- [ ] Sync job with string rate_in and correct group_filter
- [ ] GC job deprecation warning
- [ ] All S3 tests (already passing)
- [ ] Backwards compatibility tests

## References

- PBS 4.0 Source: https://github.com/proxmox/proxmox-backup
- Metrics API Types: `pbs-api-types/src/metrics.rs`
- Metrics HTTP API: `src/api2/config/metrics/influxdbhttp.rs`
- Metrics UDP API: `src/api2/config/metrics/influxdbudp.rs`
- Frontend UI: `www/window/InfluxDbEdit.js`, `www/config/MetricServerView.js`
