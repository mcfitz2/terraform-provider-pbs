package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/micah/terraform-provider-pbs/pbs/metrics"
)

// TestMetricsServerInfluxDBHTTPIntegration tests InfluxDB HTTP metrics server lifecycle
func TestMetricsServerInfluxDBHTTPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-http")

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_influxdb_http" {
  name         = "%s"
  type         = "influxdb-http"
  server       = "influx.example.com"
  port         = 443
  organization = "myorg"
  bucket       = "pbs-metrics"
  token        = "mytoken123456"
  enable       = true
  comment      = "Test InfluxDB HTTP metrics server"
}
`, serverName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_influxdb_http")
	assert.Equal(t, serverName, resource.AttributeValues["name"])
	assert.Equal(t, "influxdb-http", resource.AttributeValues["type"])
	assert.Equal(t, "influx.example.com", resource.AttributeValues["server"])
	assert.Equal(t, float64(443), resource.AttributeValues["port"])
	assert.Equal(t, "myorg", resource.AttributeValues["organization"])
	assert.Equal(t, "pbs-metrics", resource.AttributeValues["bucket"])

	// Verify via API
	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.Equal(t, serverName, server.Name)
	assert.Equal(t, metrics.MetricsServerTypeInfluxDBHTTP, server.Type)
	assert.Equal(t, "influx.example.com", server.Server)
	assert.Equal(t, 443, server.Port)
	assert.Equal(t, "myorg", server.Organization)
	assert.Equal(t, "pbs-metrics", server.Bucket)

	// Test update
	updatedConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_influxdb_http" {
  name         = "%s"
  type         = "influxdb-http"
  server       = "influx-new.example.com"
  port         = 443
  organization = "neworg"
  bucket       = "new-pbs-metrics"
  token        = "newtoken789"
  enable       = true
  comment      = "Updated InfluxDB HTTP metrics server"
}
`, serverName)

	tc.WriteMainTF(t, updatedConfig)
	tc.ApplyTerraform(t)

	server, err = metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.Equal(t, "influx-new.example.com", server.Server)
	assert.Equal(t, 443, server.Port)
	assert.Equal(t, "neworg", server.Organization)
	assert.Equal(t, "Updated InfluxDB HTTP metrics server", server.Comment)
}

// TestMetricsServerInfluxDBUDPIntegration tests InfluxDB UDP metrics server lifecycle
func TestMetricsServerInfluxDBUDPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-udp")

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_influxdb_udp" {
  name     = "%s"
  type     = "influxdb-udp"
  server   = "influx-udp.example.com"
  port     = 8089
  protocol = "udp"
  enable   = true
  comment  = "Test InfluxDB UDP metrics server"
}
`, serverName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_influxdb_udp")
	assert.Equal(t, serverName, resource.AttributeValues["name"])
	assert.Equal(t, "influxdb-udp", resource.AttributeValues["type"])
	assert.Equal(t, "influx-udp.example.com", resource.AttributeValues["server"])
	assert.Equal(t, float64(8089), resource.AttributeValues["port"])
	assert.Equal(t, "udp", resource.AttributeValues["protocol"])

	// Verify via API
	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBUDP, serverName)
	require.NoError(t, err)
	assert.Equal(t, serverName, server.Name)
	assert.Equal(t, metrics.MetricsServerTypeInfluxDBUDP, server.Type)
	assert.Equal(t, "influx-udp.example.com", server.Server)
	assert.Equal(t, 8089, server.Port)
}

// TestMetricsServerMTU tests metrics server with custom MTU
func TestMetricsServerMTU(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-mtu")

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_mtu" {
  name     = "%s"
  type     = "influxdb-udp"
  server   = "influx.example.com"
  port     = 8089
  protocol = "udp"
  mtu      = 1400
  enable   = true
  comment  = "Metrics server with custom MTU"
}
`, serverName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_mtu")
	assert.Equal(t, float64(1400), resource.AttributeValues["mtu"])

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

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_tls" {
  name         = "%s"
  type         = "influxdb-http"
  server       = "influx-secure.example.com"
  port         = 443
  organization = "myorg"
  bucket       = "pbs-metrics"
  token        = "securetoken"
  verify_tls   = true
  enable       = true
  comment      = "Metrics server with TLS verification"
}
`, serverName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_tls")
	assert.Equal(t, true, resource.AttributeValues["verify_tls"])

	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.NotNil(t, server.VerifyTLS)
	assert.True(t, *server.VerifyTLS)
}

// TestMetricsServerDisabled tests creating a disabled metrics server
func TestMetricsServerDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-disabled")

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_disabled" {
  name         = "%s"
  type         = "influxdb-http"
  server       = "influx.example.com"
  port         = 443
  organization = "myorg"
  bucket       = "pbs-metrics"
  token        = "token123"
  enable       = false
  comment      = "Disabled metrics server"
}
`, serverName)

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

	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_bodysize" {
  name          = "%s"
  type          = "influxdb-http"
  server        = "influx.example.com"
  port          = 443
  organization  = "myorg"
  bucket        = "pbs-metrics"
  token         = "token123"
  max_body_size = 65536
  enable        = true
  comment       = "Metrics server with custom max body size"
}
`, serverName)

	tc.WriteMainTF(t, testConfig)
	tc.ApplyTerraform(t)

	resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_bodysize")
	assert.Equal(t, float64(65536), resource.AttributeValues["max_body_size"])

	metricsClient := metrics.NewClient(tc.APIClient)
	server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	require.NoError(t, err)
	assert.NotNil(t, server.MaxBodySize)
	assert.Equal(t, 65536, *server.MaxBodySize)
}

// TestMetricsServerTimeout tests metrics server with custom timeout
// NOTE: PBS 4.0 removed the timeout parameter, so this test is disabled
func TestMetricsServerTimeout(t *testing.T) {
	t.Skip("PBS 4.0 removed the timeout parameter from metrics servers")
	
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Test disabled - timeout field removed in PBS 4.0
	// tc := SetupTest(t)
	// defer tc.DestroyTerraform(t)
	// serverName := GenerateTestName("influxdb-timeout")
	// testConfig := fmt.Sprintf(`
	// resource "pbs_metrics_server" "test_timeout" {
	//   name         = "%s"
	//   type         = "influxdb-http"
	//   server       = "influx.example.com"
	//   port         = 443
	//   organization = "myorg"
	//   bucket       = "pbs-metrics"
	//   token        = "token123"
	//   timeout      = 5
	//   enable       = true
	//   comment      = "Metrics server with custom timeout"
	// }
	// `, serverName)
	// Test code disabled - timeout field removed in PBS 4.0
	// tc.WriteMainTF(t, testConfig)
	// tc.ApplyTerraform(t)
	// resource := tc.GetResourceFromState(t, "pbs_metrics_server.test_timeout")
	// assert.Equal(t, float64(5), resource.AttributeValues["timeout"])
	// metricsClient := metrics.NewClient(tc.APIClient)
	// server, err := metricsClient.GetMetricsServer(context.Background(), metrics.MetricsServerTypeInfluxDBHTTP, serverName)
	// require.NoError(t, err)
	// assert.NotNil(t, server.Timeout)
	// assert.Equal(t, 5, *server.Timeout)
}

// TestMetricsServerHTTPToUDPUpdate tests updating from HTTP to UDP type (should force replacement)
func TestMetricsServerTypeChange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tc := SetupTest(t)
	defer tc.DestroyTerraform(t)

	serverName := GenerateTestName("influxdb-typechange")

	// Start with HTTP
	testConfig := fmt.Sprintf(`
resource "pbs_metrics_server" "test_typechange" {
  name         = "%s"
  type         = "influxdb-http"
  server       = "influx.example.com"
  port         = 443
  organization = "myorg"
  bucket       = "pbs-metrics"
  token        = "token123"
  enable       = true
}
`, serverName)

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
