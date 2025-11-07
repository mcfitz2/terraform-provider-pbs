#!/usr/bin/env python3
"""
Script to add skip markers to converted Go integration tests.
This marks tests as converted to HCL without deleting the code.
"""

import re
import sys
from pathlib import Path

# Map of test files to their converted tests
CONVERTED_TESTS = {
    "test/integration/jobs_test.go": [
        ("TestPruneJobIntegration", "test/tftest/jobs/prune_job.tftest.hcl"),
        ("TestPruneJobWithFilters", "test/tftest/jobs/prune_job.tftest.hcl (merged)"),
        ("TestSyncJobIntegration", "test/tftest/jobs/sync_job.tftest.hcl"),
        ("TestSyncJobWithGroupFilter", "test/tftest/jobs/sync_job.tftest.hcl (merged)"),
        ("TestVerifyJobIntegration", "test/tftest/jobs/verify_job.tftest.hcl"),
    ],
    "test/integration/remotes_test.go": [
        ("TestRemotesIntegration", "test/tftest/remotes/remote.tftest.hcl"),
        ("TestRemotePasswordUpdate", "test/tftest/remotes/remote.tftest.hcl (merged)"),
        ("TestRemoteValidation", "removed as redundant"),
        ("TestRemoteImport", "covered by HCL tests"),
    ],
    "test/integration/datasources_test.go": [
        ("TestDatastoreDataSourceIntegration", "test/tftest/datasources/datastore.tftest.hcl"),
        ("TestSyncJobsDataSourceIntegration", "test/tftest/datasources/sync_jobs.tftest.hcl"),
        ("TestVerifyJobDataSourceIntegration", "test/tftest/datasources/verify_job.tftest.hcl"),
        ("TestVerifyJobsDataSourceIntegration", "test/tftest/datasources/verify_jobs.tftest.hcl"),
        ("TestS3EndpointDataSourceIntegration", "test/tftest/datasources/s3_endpoint.tftest.hcl"),
        ("TestS3EndpointsDataSourceIntegration", "test/tftest/datasources/s3_endpoints.tftest.hcl"),
        ("TestMetricsServerDataSourceIntegration", "test/tftest/datasources/metrics_server.tftest.hcl"),
        ("TestMetricsServersDataSourceIntegration", "test/tftest/datasources/metrics_servers.tftest.hcl"),
    ],
    "test/integration/metrics_test.go": [
        ("TestMetricsServerInfluxDBHTTPIntegration", "test/tftest/metrics/influxdb_http.tftest.hcl"),
        ("TestMetricsServerInfluxDBUDPIntegration", "test/tftest/metrics/influxdb_udp.tftest.hcl"),
        ("TestMetricsServerMTU", "test/tftest/metrics/influxdb_udp.tftest.hcl (merged)"),
        ("TestMetricsServerDisabled", "test/tftest/metrics/influxdb_udp.tftest.hcl (merged)"),
        ("TestMetricsServerTypeChange", "covered by other tests"),
    ],
    "test/integration/notifications_test.go": [
        ("TestSMTPNotificationIntegration", "test/tftest/notifications/smtp.tftest.hcl"),
        ("TestGotifyNotificationIntegration", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl"),
        ("TestSendmailNotificationIntegration", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl"),
        ("TestWebhookNotificationIntegration", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl"),
        ("TestNotificationMatcherIntegration", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl"),
        ("TestNotificationMatcherModes", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl (merged)"),
        ("TestNotificationMatcherWithCalendar", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl (merged)"),
        ("TestNotificationMatcherInvertMatch", "test/tftest/notifications/endpoints_and_matcher.tftest.hcl (merged)"),
        ("TestNotificationEndpointDataSourceIntegration", "test/tftest/datasources/notification_endpoint.tftest.hcl"),
        ("TestNotificationEndpointsDataSourceIntegration", "test/tftest/datasources/notification_endpoints.tftest.hcl"),
        ("TestNotificationMatcherDataSourceIntegration", "test/tftest/datasources/notification_matcher.tftest.hcl"),
        ("TestNotificationMatchersDataSourceIntegration", "test/tftest/datasources/notification_matchers.tftest.hcl"),
    ],
}

def add_skip_marker(filepath, test_name, hcl_path):
    """Add a skip marker to the beginning of a test function."""
    path = Path(filepath)
    if not path.exists():
        print(f"⚠️  File not found: {filepath}")
        return False
    
    content = path.read_text()
    
    # Find the function definition
    pattern = rf'func {test_name}\(t \*testing\.T\) {{'
    match = re.search(pattern, content)
    
    if not match:
        print(f"⚠️  Test function not found: {test_name} in {filepath}")
        return False
    
    # Insert skip marker after the function definition
    insert_pos = match.end()
    skip_line = f'\n\tt.Skip("✅ CONVERTED TO HCL: {hcl_path}")\n'
    
    new_content = content[:insert_pos] + skip_line + content[insert_pos:]
    
    path.write_text(new_content)
    return True

def main():
    print("Adding skip markers to converted Go integration tests...\n")
    
    success_count = 0
    total_count = 0
    
    for filepath, tests in CONVERTED_TESTS.items():
        print(f"Processing {filepath}:")
        for test_name, hcl_path in tests:
            total_count += 1
            if add_skip_marker(filepath, test_name, hcl_path):
                print(f"  ✅ {test_name}")
                success_count += 1
            else:
                print(f"  ❌ {test_name}")
        print()
    
    print(f"\nSummary: {success_count}/{total_count} tests marked")
    
    if success_count == total_count:
        print("✅ All tests successfully marked!")
        return 0
    else:
        print(f"⚠️  {total_count - success_count} tests could not be marked")
        return 1

if __name__ == "__main__":
    sys.exit(main())
