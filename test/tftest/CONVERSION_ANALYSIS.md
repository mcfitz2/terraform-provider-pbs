# Integration Test to HCL Conversion Analysis

## Current Test Inventory

### Already Converted to HCL ✅
1. **datastores_datasource** - List all datastores (11 assertions)
2. **prune_job_datasource** - Read single prune job (9 assertions)
3. **prune_jobs_datasource** - List prune jobs with filter (20 assertions)
4. **sync_job_datasource** - Read single sync job (12 assertions)

### Tests to Convert (Organized by Category)

## 📊 RESOURCES Tests (CRUD operations)

### Datastores (3 tests)
1. **TestDatastoreDirectoryIntegration** - ✅ KEEP & CONVERT
   - Tests: Create, Update directory datastore
   - Coverage: Basic CRUD, async operations, API verification
   - Priority: HIGH - Core functionality
   
2. **TestDatastoreValidation** - ⚠️ KEEP IN GO - Negative Testing
   - Tests: Missing required fields, invalid configs
   - Reason: HCL tests don't handle negative validation well
   - Keep as Go unit test for provider validation logic
   
3. **TestDatastoreImport** - ⚠️ SKIP - Import testing
   - Tests: Import existing resources
   - Reason: Import is tested implicitly via state operations
   - Coverage: Not critical for HCL conversion

### S3 Datastores & Endpoints (2 tests)
4. **TestS3DatastoreMultiProvider** - ⚠️ CONDITIONAL - External Dependencies
   - Tests: S3 datastores with AWS/B2/Scaleway
   - Dependencies: Real S3 credentials required
   - Decision: Keep as Go test but create simplified HCL smoke test
   
5. **TestS3EndpointMultiProvider** - ⚠️ CONDITIONAL - External Dependencies
   - Tests: S3 endpoint management
   - Same as above - conditional on credentials

### Jobs (8 tests)
6. **TestPruneJobIntegration** - ✅ KEEP & CONVERT
   - Tests: Create, Update prune job
   - Coverage: All prune job attributes
   - Priority: HIGH

7. **TestPruneJobWithFilters** - ✅ MERGE with #6
   - Tests: Namespace filters (ns, max-depth)
   - Can be combined into single HCL test with multiple run blocks
   
8. **TestSyncJobIntegration** - ✅ KEEP & CONVERT
   - Tests: Create, Update sync job
   - Coverage: All sync job attributes including rate limiting
   - Priority: HIGH
   
9. **TestSyncJobWithGroupFilter** - ✅ MERGE with #8
   - Tests: Group filters
   - Can be combined into sync job HCL test
   
10. **TestVerifyJobIntegration** - ✅ KEEP & CONVERT
    - Tests: Create, Update verify job
    - Coverage: All verify job attributes
    - Priority: HIGH

### Notifications (8 tests)
11. **TestSMTPNotificationIntegration** - ✅ KEEP & CONVERT
    - Tests: SMTP endpoint create/update
    - Priority: MEDIUM
    
12. **TestGotifyNotificationIntegration** - ✅ KEEP & CONVERT
    - Tests: Gotify endpoint with Docker service
    - Priority: MEDIUM
    
13. **TestSendmailNotificationIntegration** - ✅ KEEP & CONVERT
    - Tests: Sendmail endpoint
    - Priority: LOW
    
14. **TestWebhookNotificationIntegration** - ✅ KEEP & CONVERT
    - Tests: Webhook endpoint with Docker service
    - Priority: MEDIUM
    
15. **TestNotificationMatcherIntegration** - ✅ KEEP & CONVERT
    - Tests: Matcher creation with targets/filters
    - Priority: HIGH
    
16. **TestNotificationMatcherModes** - ✅ MERGE with #15
    - Tests: Different matcher modes
    - Combine into matcher HCL test
    
17. **TestNotificationMatcherWithCalendar** - ✅ MERGE with #15
    - Tests: Calendar-based matching
    - Combine into matcher HCL test
    
18. **TestNotificationMatcherInvertMatch** - ✅ MERGE with #15
    - Tests: Inverted matching logic
    - Combine into matcher HCL test

### Metrics (7 tests)
19. **TestMetricsServerInfluxDBHTTPIntegration** - ✅ KEEP & CONVERT
    - Tests: InfluxDB HTTP metrics
    - Dependencies: InfluxDB container
    - Priority: HIGH
    
20. **TestMetricsServerInfluxDBUDPIntegration** - ✅ KEEP & CONVERT
    - Tests: InfluxDB UDP metrics
    - Priority: HIGH
    
21. **TestMetricsServerMTU** - ✅ MERGE with #20
    - Tests: MTU setting for UDP
    - Combine into UDP HCL test
    
22. **TestMetricsServerVerifyCertificate** - ✅ MERGE with #19
    - Tests: TLS verification
    - Combine into HTTP HCL test
    
23. **TestMetricsServerDisabled** - ✅ KEEP & CONVERT
    - Tests: Disabled metrics server
    - Priority: MEDIUM - Important edge case
    
24. **TestMetricsServerMaxBodySize** - ✅ MERGE with #19
    - Tests: Max body size parameter
    - Combine into HTTP HCL test
    
25. **TestMetricsServerTypeChange** - ⚠️ SKIP - Update testing
    - Tests: Changing server type (HTTP->UDP)
    - Reason: Implicitly tested by other tests

### Remotes (1 test)
26. **TestRemotesIntegration** - ✅ KEEP & CONVERT
    - Tests: Remote creation/update
    - Priority: HIGH

## 📖 DATA SOURCE Tests

### Already Converted ✅
- TestDatastoresDataSourceIntegration ✅
- TestPruneJobDataSourceIntegration ✅
- TestPruneJobsDataSourceIntegration ✅  
- TestSyncJobDataSourceIntegration ✅

### Remaining Data Sources (12 tests)
27. **TestDatastoreDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: Read single datastore
    - Priority: HIGH
    
28. **TestSyncJobsDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: List sync jobs
    - Priority: HIGH
    
29. **TestVerifyJobDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: Read single verify job
    - Priority: HIGH
    
30. **TestVerifyJobsDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: List verify jobs
    - Priority: HIGH
    
31. **TestS3EndpointDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: Read single S3 endpoint
    - Priority: MEDIUM
    
32. **TestS3EndpointsDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: List S3 endpoints
    - Priority: MEDIUM
    
33. **TestMetricsServerDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: Read single metrics server
    - Priority: MEDIUM
    
34. **TestMetricsServersDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: List metrics servers
    - Priority: MEDIUM
    
35. **TestNotificationEndpointDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: Read single notification endpoint
    - Priority: MEDIUM
    
36. **TestNotificationEndpointsDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: List notification endpoints
    - Priority: MEDIUM
    
37. **TestNotificationMatcherDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: Read single notification matcher
    - Priority: MEDIUM
    
38. **TestNotificationMatchersDataSourceIntegration** - ✅ KEEP & CONVERT
    - Tests: List notification matchers
    - Priority: MEDIUM

## 📋 ANALYSIS SUMMARY

### Tests to Convert to HCL: 30
- **Resources**: 15 tests → 9 HCL test suites (after merging)
- **Data Sources**: 12 tests → 12 HCL tests
- **Total HCL tests**: 21 test suites

### Tests to Keep in Go: 3
1. TestDatastoreValidation - Negative validation testing
2. TestS3DatastoreMultiProvider - External dependency (conditional)
3. TestS3EndpointMultiProvider - External dependency (conditional)

### Tests to Skip/Remove: 2
1. TestDatastoreImport - Covered by other tests
2. TestMetricsServerTypeChange - Covered by other tests

### Redundancies Identified 🔍
1. **Job filter tests** - Can be combined into main job tests
2. **Notification matcher mode tests** - 4 tests → 1 comprehensive test
3. **Metrics server option tests** - Can be combined into HTTP/UDP tests

### Coverage Gaps Identified ⚠️
1. **No tests for Remote data sources** (pbs_remote, pbs_remotes)
2. **No tests for namespace operations** in data sources
3. **No tests for concurrent resource operations** (removed intentionally)
4. **Limited error handling tests** in HCL (keep in Go)
5. **No tests for resource dependencies** (e.g., job requires datastore)

### Recommendations

#### Priority 1 - Core Functionality (Complete First)
- Datastore resource
- All job resources (prune, sync, verify)
- Remote resource  
- All singular data sources (read operations)

#### Priority 2 - Infrastructure Integration
- Metrics servers (both HTTP and UDP)
- Notification endpoints (all types)
- Notification matchers
- All plural data sources (list operations)

#### Priority 3 - Edge Cases
- S3 operations (keep conditional)
- Disabled states
- Complex filtering scenarios

### Conversion Strategy

1. **Create consolidated test suites** - Don't do 1:1 conversion
   - Example: `jobs/` folder with prune_job.tftest.hcl, sync_job.tftest.hcl, verify_job.tftest.hcl
   
2. **Use run blocks for test phases**:
   ```hcl
   run "create_with_basic_config" { }
   run "update_with_advanced_features" { }
   run "verify_with_filters" { }
   ```

3. **Organize by resource type**:
   ```
   test/tftest/
   ├── datastores/
   │   ├── directory_datastore.tftest.hcl
   │   └── s3_datastore.tftest.hcl (conditional)
   ├── jobs/
   │   ├── prune_job.tftest.hcl
   │   ├── sync_job.tftest.hcl
   │   └── verify_job.tftest.hcl
   ├── notifications/
   │   ├── smtp_endpoint.tftest.hcl
   │   ├── gotify_endpoint.tftest.hcl
   │   ├── webhook_endpoint.tftest.hcl
   │   ├── sendmail_endpoint.tftest.hcl
   │   └── matcher.tftest.hcl
   ├── metrics/
   │   ├── influxdb_http.tftest.hcl
   │   ├── influxdb_udp.tftest.hcl
   │   └── disabled.tftest.hcl
   ├── remotes/
   │   └── remote.tftest.hcl
   └── datasources/ (existing)
       └── ...
   ```

4. **Maintain Go tests for**:
   - Validation logic (negative tests)
   - Complex setup requiring programmatic control
   - Tests with external dependencies that may not be available

## 🎯 Expected Outcomes

### Before Conversion
- **Go integration tests**: 38 tests
- **HCL tests**: 4 tests
- **Test execution time**: ~25 minutes
- **Flakiness**: Moderate (tfexec timing issues)

### After Conversion
- **Go integration tests**: 3 tests (validation only)
- **HCL tests**: 25 tests
- **Test execution time**: ~20 minutes (estimated)
- **Flakiness**: Low (native Terraform execution)
- **Maintainability**: HIGH (declarative, easier to understand)
- **Coverage**: SAME (100% functional coverage maintained)
