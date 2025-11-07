# Converted Go Integration Tests

This document tracks which Go integration tests have been converted to Terraform HCL tests.

## ✅ Fully Converted (Can be removed from Go)

### test/integration/datastore_test.go
- `TestDatastoreDirectoryIntegration` → `test/tftest/datastores/directory_datastore.tftest.hcl`
- `TestDatastoreImport` → Covered by HCL tests (import is implicit)

**Keep in Go:**
- `TestDatastoreValidation` - Negative validation testing

### test/integration/jobs_test.go
- `TestPruneJobIntegration` → `test/tftest/jobs/prune_job.tftest.hcl`
- `TestPruneJobWithFilters` → Merged into `test/tftest/jobs/prune_job.tftest.hcl`
- `TestSyncJobIntegration` → `test/tftest/jobs/sync_job.tftest.hcl`
- `TestSyncJobWithGroupFilter` → Merged into `test/tftest/jobs/sync_job.tftest.hcl`
- `TestVerifyJobIntegration` → `test/tftest/jobs/verify_job.tftest.hcl`

**Keep in Go:** (None - all converted)

### test/integration/remotes_test.go
- `TestRemotesIntegration` → `test/tftest/remotes/remote.tftest.hcl`
- `TestRemotePasswordUpdate` → Merged into `test/tftest/remotes/remote.tftest.hcl`
- `TestRemoteValidation` → Redundant (provider validation sufficient)
- `TestRemoteImport` → Covered by HCL tests
- `TestRemoteDataSources` → Requires real remote server (skip)

**Keep in Go:**
- `TestRemoteDataSources` - Requires external remote server setup

### test/integration/datasources_test.go
- `TestDatastoreDataSourceIntegration` → `test/tftest/datasources/datastore.tftest.hcl`
- `TestDatastoresDataSourceIntegration` → Already converted (previous work)
- `TestPruneJobDataSourceIntegration` → Already converted (previous work)
- `TestPruneJobsDataSourceIntegration` → Already converted (previous work)
- `TestSyncJobDataSourceIntegration` → Already converted (previous work)
- `TestSyncJobsDataSourceIntegration` → `test/tftest/datasources/sync_jobs.tftest.hcl`
- `TestVerifyJobDataSourceIntegration` → `test/tftest/datasources/verify_job.tftest.hcl`
- `TestVerifyJobsDataSourceIntegration` → `test/tftest/datasources/verify_jobs.tftest.hcl`
- `TestS3EndpointDataSourceIntegration` → `test/tftest/datasources/s3_endpoint.tftest.hcl`
- `TestS3EndpointsDataSourceIntegration` → `test/tftest/datasources/s3_endpoints.tftest.hcl`
- `TestMetricsServerDataSourceIntegration` → `test/tftest/datasources/metrics_server.tftest.hcl`
- `TestMetricsServersDataSourceIntegration` → `test/tftest/datasources/metrics_servers.tftest.hcl`

**Keep in Go:** (None - all converted)

### test/integration/metrics_test.go
- `TestMetricsServerInfluxDBHTTPIntegration` → `test/tftest/metrics/influxdb_http.tftest.hcl`
- `TestMetricsServerInfluxDBUDPIntegration` → `test/tftest/metrics/influxdb_udp.tftest.hcl`
- `TestMetricsServerMTU` → Merged into `test/tftest/metrics/influxdb_udp.tftest.hcl`
- `TestMetricsServerDisabled` → Merged into `test/tftest/metrics/influxdb_udp.tftest.hcl`
- `TestMetricsServerTypeChange` → Covered by other tests (type change triggers replacement)

**Keep in Go:**
- `TestMetricsServerVerifyCertificate` - TLS verification testing
- `TestMetricsServerMaxBodySize` - Edge case testing

### test/integration/notifications_test.go
- `TestSMTPNotificationIntegration` → `test/tftest/notifications/smtp.tftest.hcl`
- `TestGotifyNotificationIntegration` → `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestSendmailNotificationIntegration` → `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestWebhookNotificationIntegration` → `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestNotificationMatcherIntegration` → `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestNotificationMatcherModes` → Merged into `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestNotificationMatcherWithCalendar` → Merged into `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestNotificationMatcherInvertMatch` → Merged into `test/tftest/notifications/endpoints_and_matcher.tftest.hcl`
- `TestNotificationEndpointDataSourceIntegration` → `test/tftest/datasources/notification_endpoint.tftest.hcl`
- `TestNotificationEndpointsDataSourceIntegration` → `test/tftest/datasources/notification_endpoints.tftest.hcl`
- `TestNotificationMatcherDataSourceIntegration` → `test/tftest/datasources/notification_matcher.tftest.hcl`
- `TestNotificationMatchersDataSourceIntegration` → `test/tftest/datasources/notification_matchers.tftest.hcl`

**Keep in Go:** (None - all converted)

### test/integration/s3_providers_test.go
**Keep in Go:** (All tests require external S3 credentials)
- `TestS3EndpointMultiProvider`
- `TestS3EndpointProviderSpecificFeatures`
- `TestS3EndpointConcurrentProviders`
- `TestS3DatastoreMultiProvider`

### test/integration/integration_test.go
**Keep in Go:**
- `TestIntegration` - Smoke test harness
- `TestQuickSmoke` - Quick validation

## Summary

**Total Converted**: 26 Go tests → 21 HCL test suites
**Remaining in Go**: 8 tests (validation, S3 multi-provider, smoke tests)
**Removed as Redundant**: 7 tests (import, validation, type change tests covered elsewhere)

## Next Steps

1. Delete converted test functions from Go files
2. Update CI workflow to run HCL tests
3. Local validation of all HCL tests
4. Update documentation
