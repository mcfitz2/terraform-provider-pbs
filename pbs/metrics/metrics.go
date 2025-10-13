/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package metrics provides API client functionality for PBS metrics server configurations
package metrics

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/url"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Client represents the metrics API client
type Client struct {
	api *api.Client
}

// NewClient creates a new metrics API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// MetricsServerType represents the type of metrics server
type MetricsServerType string

const (
	MetricsServerTypeInfluxDBUDP  MetricsServerType = "influxdb-udp"
	MetricsServerTypeInfluxDBHTTP MetricsServerType = "influxdb-http"
)

// MetricsServer represents a metrics server configuration
type MetricsServer struct {
	Name         string            `json:"name"`
	Type         MetricsServerType `json:"type"`
	Enable       *bool             `json:"enable,omitempty"`
	// PBS 4.0 fields
	URL          string            `json:"url,omitempty"`           // InfluxDB HTTP only (PBS 4.0)
	Host         string            `json:"host,omitempty"`          // InfluxDB UDP only (PBS 4.0)
	// Backwards compatibility fields (will be converted to URL/Host)
	Server       string            `json:"server,omitempty"`
	Port         int               `json:"port,omitempty"`
	MTU          *int              `json:"mtu,omitempty"`
	Organization string            `json:"organization,omitempty"`  // InfluxDB HTTP only
	Bucket       string            `json:"bucket,omitempty"`        // InfluxDB HTTP only
	Token        string            `json:"token,omitempty"`         // InfluxDB HTTP only
	MaxBodySize  *int              `json:"max-body-size,omitempty"` // InfluxDB HTTP only
	VerifyTLS    *bool             `json:"verify-tls,omitempty"`    // InfluxDB HTTP only (renamed from verify_certificate)
	Comment      string            `json:"comment,omitempty"`
}

// ListMetricsServers lists all metrics server configurations
func (c *Client) ListMetricsServers(ctx context.Context) ([]MetricsServer, error) {
	resp, err := c.api.Get(ctx, "/config/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics servers: %w", err)
	}

	var servers []MetricsServer
	if err := json.Unmarshal(resp.Data, &servers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics servers: %w", err)
	}

	// PBS 4.0 returns URL/Host, parse them to Server+Port for backwards compatibility
	for i := range servers {
		servers[i].parseURLFields()
	}

	return servers, nil
}

// GetMetricsServer gets a specific metrics server configuration by name
func (c *Client) GetMetricsServer(ctx context.Context, serverType MetricsServerType, name string) (*MetricsServer, error) {
	path := fmt.Sprintf("/config/metrics/%s/%s", serverType, url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics server %s: %w", name, err)
	}

	var server MetricsServer
	if err := json.Unmarshal(resp.Data, &server); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics server %s: %w", name, err)
	}

	// PBS 4.0 returns URL/Host, parse them to Server+Port for backwards compatibility
	server.parseURLFields()
	server.Type = serverType

	return &server, nil
}

// parseURLFields extracts Server and Port from URL/Host fields for backwards compatibility
func (s *MetricsServer) parseURLFields() {
	if s.URL != "" {
		// Parse https://hostname:port
		if u, err := url.Parse(s.URL); err == nil {
			s.Server = u.Hostname()
			if u.Port() != "" {
				fmt.Sscanf(u.Port(), "%d", &s.Port)
			} else if u.Scheme == "https" {
				s.Port = 443
			} else if u.Scheme == "http" {
				s.Port = 80
			}
		}
	} else if s.Host != "" {
		// Parse hostname:port
		if host, portStr, err := net.SplitHostPort(s.Host); err == nil {
			s.Server = host
			fmt.Sscanf(portStr, "%d", &s.Port)
		} else {
			// No port specified
			s.Server = s.Host
		}
	}
}

// CreateMetricsServer creates a new metrics server configuration
func (c *Client) CreateMetricsServer(ctx context.Context, server *MetricsServer) error {
	if server.Name == "" {
		return fmt.Errorf("server name is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{
		"name": server.Name,
	}

	if server.Enable != nil {
		body["enable"] = *server.Enable
	}
	if server.Comment != "" {
		body["comment"] = server.Comment
	}

	// Type-specific fields for PBS 4.0
	switch server.Type {
	case MetricsServerTypeInfluxDBHTTP:
		// PBS 4.0 expects 'url' field instead of separate server+port
		if server.URL != "" {
			body["url"] = server.URL
		} else if server.Server != "" && server.Port > 0 {
			// Backwards compatibility: construct URL from server+port
			body["url"] = fmt.Sprintf("https://%s:%d", server.Server, server.Port)
		} else {
			return fmt.Errorf("either url or server+port is required for InfluxDB HTTP")
		}

		if server.Organization != "" {
			body["organization"] = server.Organization
		}
		if server.Bucket != "" {
			body["bucket"] = server.Bucket
		}
		if server.Token != "" {
			body["token"] = server.Token
		}
		if server.MaxBodySize != nil {
			body["max-body-size"] = *server.MaxBodySize
		}
		if server.VerifyTLS != nil {
			body["verify-tls"] = *server.VerifyTLS
		}

	case MetricsServerTypeInfluxDBUDP:
		// PBS 4.0 expects 'host' field in format 'hostname:port'
		if server.Host != "" {
			body["host"] = server.Host
		} else if server.Server != "" && server.Port > 0 {
			// Backwards compatibility: construct host from server+port
			body["host"] = fmt.Sprintf("%s:%d", server.Server, server.Port)
		} else {
			return fmt.Errorf("either host or server+port is required for InfluxDB UDP")
		}

		if server.MTU != nil {
			body["mtu"] = *server.MTU
		}
	}

	path := fmt.Sprintf("/config/metrics/%s", server.Type)
	_, err := c.api.Post(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to create metrics server %s: %w", server.Name, err)
	}

	return nil
}

// UpdateMetricsServer updates an existing metrics server configuration
func (c *Client) UpdateMetricsServer(ctx context.Context, serverType MetricsServerType, name string, server *MetricsServer) error {
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{}

	if server.Enable != nil {
		body["enable"] = *server.Enable
	}
	if server.Comment != "" {
		body["comment"] = server.Comment
	}

	// Type-specific fields for PBS 4.0
	switch serverType {
	case MetricsServerTypeInfluxDBHTTP:
		// PBS 4.0 expects 'url' field
		if server.URL != "" {
			body["url"] = server.URL
		} else if server.Server != "" && server.Port > 0 {
			// Backwards compatibility: construct URL from server+port
			body["url"] = fmt.Sprintf("https://%s:%d", server.Server, server.Port)
		}

		if server.Organization != "" {
			body["organization"] = server.Organization
		}
		if server.Bucket != "" {
			body["bucket"] = server.Bucket
		}
		if server.Token != "" {
			body["token"] = server.Token
		}
		if server.MaxBodySize != nil {
			body["max-body-size"] = *server.MaxBodySize
		}
		if server.VerifyTLS != nil {
			body["verify-tls"] = *server.VerifyTLS
		}

	case MetricsServerTypeInfluxDBUDP:
		// PBS 4.0 expects 'host' field in format 'hostname:port'
		if server.Host != "" {
			body["host"] = server.Host
		} else if server.Server != "" && server.Port > 0 {
			// Backwards compatibility: construct host from server+port
			body["host"] = fmt.Sprintf("%s:%d", server.Server, server.Port)
		}

		if server.MTU != nil {
			body["mtu"] = *server.MTU
		}
	}

	path := fmt.Sprintf("/config/metrics/%s/%s", serverType, url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update metrics server %s: %w", name, err)
	}

	return nil
}

// DeleteMetricsServer deletes a metrics server configuration
func (c *Client) DeleteMetricsServer(ctx context.Context, serverType MetricsServerType, name string) error {
	if name == "" {
		return fmt.Errorf("server name is required")
	}

	path := fmt.Sprintf("/config/metrics/%s/%s", serverType, url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete metrics server %s: %w", name, err)
	}

	return nil
}
