package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/metrics"
)

// getInfluxDBHost returns the InfluxDB host from environment or default
func getInfluxDBHost() string {
	if host := os.Getenv("TEST_INFLUXDB_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// getInfluxDBPort returns the InfluxDB port from environment or default
func getInfluxDBPort() int {
	if port := os.Getenv("TEST_INFLUXDB_PORT"); port != "" {
		// Parse port from string, default to 8086 if parse fails
		var portNum int
		if _, err := fmt.Sscanf(port, "%d", &portNum); err == nil {
			return portNum
		}
	}
	return 8086
}

// getInfluxDBUDPHost returns the InfluxDB UDP host from environment or default
func getInfluxDBUDPHost() string {
	if host := os.Getenv("TEST_INFLUXDB_UDP_HOST"); host != "" {
		return host
	}
	return "localhost"
}

// getInfluxDBUDPPort returns the InfluxDB UDP port from environment or default
func getInfluxDBUDPPort() int {
	if port := os.Getenv("TEST_INFLUXDB_UDP_PORT"); port != "" {
		var portNum int
		if _, err := fmt.Sscanf(port, "%d", &portNum); err == nil {
			return portNum
		}
	}
	return 8089
}

// TestMetricsServerInfluxDBHTTPIntegration tests InfluxDB HTTP metrics server lifecycle
func TestMetricsServerInfluxDBHTTPIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/metrics/influxdb_http.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-http")
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_influxdb_http" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "testorg"
  bucket       = "pbs-metrics"
  token        = "test-token-123456"
  enable       = true
  comment      = "Test InfluxDB HTTP metrics server"
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_influxdb_http")
	assert.Equal(t, serverName, resource.AttributeValues["name"])
	assert.Equal(t, "influxdb-http", resource.AttributeValues["type"])
	expectedURL := fmt.Sprintf("http://%s:%d", influxHost, influxPort)
	assert.Equal(t, expectedURL, resource.AttributeValues["url"])
	assert.Equal(t, "testorg", resource.AttributeValues["organization"])
	assert.Equal(t, "pbs-metrics", resource.AttributeValues["bucket"])

	// Verify via API
	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.Equal(t, serverName, server.Name)
	assert.Equal(t, metrics.MetricsServerTypeInfluxDBHTTP, server.Type)
	assert.Equal(t, expectedURL, server.URL)
	assert.Equal(t, "testorg", server.Organization)
	assert.Equal(t, "pbs-metrics", server.Bucket)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_influxdb_http" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "testorg"
  bucket       = "pbs-metrics"
  token        = "test-token-123456"
  enable       = true
  comment      = "Updated InfluxDB HTTP metrics server"
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	server, err = metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.Equal(t, expectedURL, server.URL)
	assert.Equal(t, "testorg", server.Organization)
	assert.Equal(t, "Updated InfluxDB HTTP metrics server", server.Comment)
}

// TestMetricsServerInfluxDBUDPIntegration tests InfluxDB UDP metrics server lifecycle
func TestMetricsServerInfluxDBUDPIntegration(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/metrics/influxdb_udp.tftest.hcl")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-udp")
	influxUDPHost := getInfluxDBUDPHost()
	influxUDPPort := getInfluxDBUDPPort()

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_influxdb_udp" {
  name     = "%s"
  type     = "influxdb-udp"
  server   = "%s"
  port     = %d
  protocol = "udp"
  enable   = true
  comment  = "Test InfluxDB UDP metrics server"
}
`, serverName, influxUDPHost, influxUDPPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_influxdb_udp")
	assert.Equal(t, serverName, resource.AttributeValues["name"])
	assert.Equal(t, "influxdb-udp", resource.AttributeValues["type"])
	assert.Equal(t, influxUDPHost, resource.AttributeValues["server"])
	assert.Equal(t, json.Number(fmt.Sprintf("%d", influxUDPPort)), resource.AttributeValues["port"])
	assert.Equal(t, "udp", resource.AttributeValues["protocol"])

	// Verify via API
	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBUDP, serverName)
	require.NoError(t, err)
	assert.Equal(t, serverName, server.Name)
	assert.Equal(t, metrics.MetricsServerTypeInfluxDBUDP, server.Type)
	assert.Equal(t, influxUDPHost, server.Server)
	assert.Equal(t, influxUDPPort, server.Port)
}

// TestMetricsServerMTU tests metrics server with custom MTU
func TestMetricsServerMTU(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/metrics/influxdb_udp.tftest.hcl (merged)")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-mtu")
	influxUDPHost := getInfluxDBUDPHost()
	influxUDPPort := getInfluxDBUDPPort()

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_mtu" {
  name     = "%s"
  type     = "influxdb-udp"
  server   = "%s"
  port     = %d
  protocol = "udp"
  mtu      = 1400
  enable   = true
  comment  = "Metrics server with custom MTU"
}
`, serverName, influxUDPHost, influxUDPPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_mtu")
	assert.Equal(t, json.Number("1400"), resource.AttributeValues["mtu"])

	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBUDP, serverName)
	require.NoError(t, err)
	assert.NotNil(t, server.MTU)
	assert.Equal(t, 1400, *server.MTU)
}

// TestMetricsServerVerifyCertificate tests metrics server with certificate verification
func TestMetricsServerVerifyCertificate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-tls")
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_tls" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "testorg"
  bucket       = "pbs-metrics"
  token        = "test-token-123456"
  verify_tls   = false
  enable       = true
  comment      = "Metrics server with TLS verification disabled"
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_tls")
	assert.Equal(t, false, resource.AttributeValues["verify_tls"])

	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.NotNil(t, server.VerifyTLS)
	assert.False(t, *server.VerifyTLS)
}

// TestMetricsServerDisabled tests creating a disabled metrics server
func TestMetricsServerDisabled(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: test/tftest/metrics/influxdb_udp.tftest.hcl (merged)")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-disabled")
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_disabled" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "testorg"
  bucket       = "pbs-metrics"
  token        = "test-token-123456"
  enable       = false
  comment      = "Disabled metrics server"
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_disabled")
	assert.Equal(t, false, resource.AttributeValues["enable"])

	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.NotNil(t, server.Enable)
	assert.False(t, *server.Enable)
}

// TestMetricsServerMaxBodySize tests metrics server with custom max body size
func TestMetricsServerMaxBodySize(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-bodysize")
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_bodysize" {
  name          = "%s"
  type          = "influxdb-http"
  url           = "http://%s:%d"
  organization  = "testorg"
  bucket        = "pbs-metrics"
  token         = "test-token-123456"
  max_body_size = 65536
  enable        = true
  comment       = "Metrics server with custom max body size"
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_bodysize")
	assert.Equal(t, json.Number("65536"), resource.AttributeValues["max_body_size"])

	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.NotNil(t, server.MaxBodySize)
	assert.Equal(t, 65536, *server.MaxBodySize)
}

// TestMetricsServerTimeout tests metrics server with custom timeout
// NOTE: PBS 4.0 removed the timeout parameter, so this test is disabled

// TestMetricsServerHTTPToUDPUpdate tests updating from HTTP to UDP type (should force replacement)
func TestMetricsServerTypeChange(t *testing.T) {
	t.Skip("✅ CONVERTED TO HCL: covered by other tests")

	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-typechange")
	influxHost := getInfluxDBHost()
	influxPort := getInfluxDBPort()

	// Start with HTTP
	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_typechange" {
  name         = "%s"
  type         = "influxdb-http"
  url          = "http://%s:%d"
  organization = "testorg"
  bucket       = "pbs-metrics"
  token        = "test-token-123456"
  enable       = true
}
`, serverName, influxHost, influxPort)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	// Verify HTTP configuration
	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.Equal(t, metrics.MetricsServerTypeInfluxDBHTTP, server.Type)

	// Note: Changing type should trigger replacement
	// This test verifies the resource exists and type is correct
	// In practice, changing the type would require destroy + create
}
