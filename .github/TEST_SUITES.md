# Integration Test Suites

The integration tests are organized into three distinct suites based on their requirements and dependencies.

## Test Suite Overview

### 1. Core Integration Tests (Non-S3, Non-Hardware)

**Location**: Runs in Docker workflow  
**Coverage**: All core PBS features except S3 and hardware-dependent tests

**Includes:**
- ✅ Quick smoke tests (`TestQuickSmoke`)
- ✅ Datastore tests - Directory, Network Storage (NFS/CIFS), Validation, Import
  - ❌ **Excludes**: `TestDatastoreZFS` (hardware-dependent)
- ✅ Metrics server tests - HTTP, UDP, MTU, Verify Certificate, Disabled state, Max body size, Type changes
- ✅ Job management tests - Prune, Sync, Verify, GC jobs with filters
- ✅ Notification tests - SMTP, Gotify, Sendmail, Webhook, Endpoints, Matchers

**Requirements:**
- PBS container
- InfluxDB 2.7 (HTTP) container
- InfluxDB 1.8 (UDP) container
- Gotify container (for notification tests)
- Webhook receiver container (for notification tests)
- NFS server container (for network storage tests)
- CIFS/Samba server container (for network storage tests)

**Run Pattern:**
```bash
go test ./test/integration \
  -run "TestQuickSmoke|TestDatastore|TestMetrics|Test.*Job|Test.*Notification" \
  -skip "TestDatastoreZFS|TestS3|TestCleanup"
```

---

### 2. S3 Integration Tests

**Location**: Runs in Docker workflow (separate jobs per provider)  
**Coverage**: All S3-related functionality across multiple cloud providers

**Includes:**
- ✅ S3 endpoint tests - Multi-provider support
- ✅ Provider-specific features - AWS, Backblaze B2, Scaleway
- ✅ S3 cleanup tests - Bucket cleanup across all providers

**Requirements:**
- PBS container
- Valid S3 credentials (repository secrets):
  - `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY`
  - `B2_ACCESS_KEY_ID` / `B2_SECRET_ACCESS_KEY`
  - `SCALEWAY_ACCESS_KEY` / `SCALEWAY_SECRET_KEY`

**Jobs:**
- `s3-aws-tests` - AWS S3 integration tests
- `s3-backblaze-tests` - Backblaze B2 integration tests
- `s3-scaleway-tests` - Scaleway integration tests
- `s3-cleanup-tests` - Cleanup test buckets across all providers

**Run Pattern:**
```bash
# AWS tests
go test ./test/integration -run "AWS"

# Backblaze B2 tests
go test ./test/integration -run "Backblaze|B2"

# Scaleway tests
go test ./test/integration -run "Scaleway"

# Cleanup tests
go test ./test/integration -run "TestCleanup"
```

---

### 3. Hardware-Dependent Tests

**Location**: Runs ONLY in VM workflow  
**Coverage**: Tests that require specific hardware or kernel features

**Includes:**
- ✅ ZFS datastore tests (`TestDatastoreZFSIntegration`)
  - Requires ZFS kernel modules
  - Needs actual block devices or ZFS pools
  - Cannot run in standard Docker containers

**Requirements:**
- Self-hosted VM runner
- ZFS kernel modules installed
- ZFS pool or disk available
- All containers from Core suite

**Run Pattern:**
```bash
# Run ALL tests including ZFS
go test ./test/integration
```

---

## Workflow Execution

### Docker Integration Tests Workflow

**File**: `.github/workflows/integration-tests.yml`

**Runs:**
1. ✅ Unit tests (all platforms)
2. ✅ Core integration tests (Non-S3, Non-Hardware)
3. ✅ S3 integration tests (if credentials available)
4. ❌ Hardware-dependent tests (EXCLUDED)

**Trigger:**
- On push to `main` or `develop`
- On pull requests to `main`
- Manual workflow dispatch

---

### VM Integration Tests Workflow

**File**: `.github/workflows/vm-integration-tests.yml`

**Runs:**
1. ✅ ALL Core integration tests
2. ✅ ALL S3 integration tests (if credentials available)
3. ✅ ALL Hardware-dependent tests (ZFS)

**Trigger:**
- On push to `main` or `develop`
- On pull requests to `main`
- Manual workflow dispatch
- Weekly schedule (Sunday 2 AM UTC)

**Runner:** Self-hosted VM with ZFS support

---

## Test Execution Matrix

| Test Suite | Docker Workflow | VM Workflow |
|------------|----------------|-------------|
| Unit Tests | ✅ | ✅ |
| Core Integration (Non-ZFS) | ✅ | ✅ |
| S3 Integration | ✅ | ✅ |
| Hardware (ZFS) | ❌ | ✅ |

---

## Running Tests Locally

### Core Integration Tests

```bash
# Start required containers
cd test
docker-compose up -d

# Run core tests
PBS_ADDRESS="https://localhost:8007" \
PBS_USERNAME="admin@pbs" \
PBS_PASSWORD="pbspbs" \
PBS_INSECURE_TLS="true" \
TEST_INFLUXDB_HOST="localhost" \
TEST_INFLUXDB_PORT="8086" \
TEST_INFLUXDB_UDP_HOST="localhost" \
TEST_INFLUXDB_UDP_PORT="8089" \
TF_ACC=1 go test -v ./test/integration \
  -run "TestQuickSmoke|TestDatastore|TestMetrics|Test.*Job|Test.*Notification" \
  -skip "TestDatastoreZFS|TestS3|TestCleanup"
```

### S3 Integration Tests

```bash
# Set S3 credentials
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-west-2"

# Run S3 tests
PBS_ADDRESS="https://localhost:8007" \
PBS_USERNAME="admin@pbs" \
PBS_PASSWORD="pbspbs" \
PBS_INSECURE_TLS="true" \
TF_ACC=1 go test -v ./test/integration -run "AWS"
```

### Hardware-Dependent Tests (ZFS)

```bash
# Requires VM with ZFS kernel modules and pool
# Ensure ZFS pool is available first
zpool list

# Run all tests including ZFS
PBS_ADDRESS="https://192.168.1.108:8007" \
PBS_USERNAME="root@pam" \
PBS_PASSWORD="pbspbs123" \
PBS_INSECURE_TLS="true" \
PBS_TESTPOOL="testpool" \
TF_ACC=1 go test -v -timeout 40m ./test/integration
```

---

## Coverage Reporting

Each test suite generates separate coverage reports:

- `coverage-unit.out` - Unit test coverage
- `coverage-core.out` - Core integration test coverage
- `coverage-aws.out` - AWS S3 test coverage
- `coverage-b2.out` - Backblaze B2 test coverage
- `coverage-scaleway.out` - Scaleway test coverage
- `coverage-vm-integration.out` - Full VM test coverage (includes all suites)

All coverage reports are:
1. Uploaded as workflow artifacts (7 day retention)
2. Reported to Codecov with appropriate flags
3. Combined for overall project coverage metrics

---

## Adding New Tests

When adding new tests, consider which suite they belong to:

**Add to Core Suite** if:
- Tests core PBS functionality
- Doesn't require S3 credentials
- Doesn't require special hardware/kernel features
- Can run in Docker containers

**Add to S3 Suite** if:
- Tests S3 datastore endpoints
- Requires cloud provider credentials
- Tests provider-specific S3 features

**Add to Hardware Suite** if:
- Requires ZFS kernel modules
- Needs real hardware resources
- Cannot run in standard containers
- Requires specific kernel features

---

## Troubleshooting

### Core Tests Failing in Docker

1. Check container health: `docker ps`
2. Verify PBS is responding: `curl -k https://localhost:8007`
3. Check InfluxDB containers: `docker logs influxdb-test`
4. Verify environment variables are set

### S3 Tests Skipping

- Verify credentials are set in repository secrets
- Check credential format matches provider requirements
- Confirm secrets are available for your branch

### ZFS Tests Failing in VM

1. Verify ZFS modules: `lsmod | grep zfs`
2. Check ZFS pool: `zpool list`
3. Ensure test pool exists: `zpool status testpool`
4. Check PBS access: `curl -k https://192.168.1.108:8007`

---

## CI/CD Optimization

### Parallel Execution

Tests are organized to run in parallel where possible:

```
Unit Tests
    ↓
Core Integration Tests (15-20m)
    ↓
├─ S3 AWS Tests (20m)
├─ S3 B2 Tests (20m)
└─ S3 Scaleway Tests (20m)
    ↓
S3 Cleanup Tests (10m)
```

### Conditional Execution

- S3 tests only run on `push` or `workflow_dispatch` (not on every PR)
- S3 tests skip gracefully if credentials unavailable
- Hardware tests only run in VM workflow
- VM workflow runs weekly + on important branches

This organization ensures:
- ✅ Fast feedback for core changes
- ✅ Comprehensive testing before merge
- ✅ Resource-efficient CI/CD pipeline
- ✅ Clear separation of concerns
