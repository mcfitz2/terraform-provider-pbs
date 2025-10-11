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
	Server       string            `json:"server"`
	Port         int               `json:"port"`
	MTU          *int              `json:"mtu,omitempty"`
	Organization string            `json:"organization,omitempty"`  // InfluxDB HTTP only
	Bucket       string            `json:"bucket,omitempty"`        // InfluxDB HTTP only
	Token        string            `json:"token,omitempty"`         // InfluxDB HTTP only
	MaxBodySize  *int              `json:"max-body-size,omitempty"` // InfluxDB HTTP only
	VerifyTLS    *bool             `json:"verify-tls,omitempty"`    // InfluxDB HTTP only
	Timeout      *int              `json:"timeout,omitempty"`       // InfluxDB HTTP only
	Protocol     string            `json:"proto,omitempty"`         // InfluxDB UDP only (udp/tcp)
	Comment      string            `json:"comment,omitempty"`
}

// ListMetricsServers lists all metrics server configurations
func (c *Client) ListMetricsServers(ctx context.Context) ([]MetricsServer, error) {
	resp, err := c.api.Get(ctx, "/config/metrics/server")
	if err != nil {
		return nil, fmt.Errorf("failed to list metrics servers: %w", err)
	}

	var servers []MetricsServer
	if err := json.Unmarshal(resp.Data, &servers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics servers: %w", err)
	}

	return servers, nil
}

// GetMetricsServer gets a specific metrics server configuration by name
func (c *Client) GetMetricsServer(ctx context.Context, serverType MetricsServerType, name string) (*MetricsServer, error) {
	path := fmt.Sprintf("/config/metrics/server/%s/%s", serverType, url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics server %s: %w", name, err)
	}

	var server MetricsServer
	if err := json.Unmarshal(resp.Data, &server); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics server %s: %w", name, err)
	}

	return &server, nil
}

// CreateMetricsServer creates a new metrics server configuration
func (c *Client) CreateMetricsServer(ctx context.Context, server *MetricsServer) error {
	if server.Name == "" {
		return fmt.Errorf("server name is required")
	}
	if server.Server == "" {
		return fmt.Errorf("server address is required")
	}
	if server.Port == 0 {
		return fmt.Errorf("server port is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{
		"name":   server.Name,
		"server": server.Server,
		"port":   server.Port,
	}

	if server.Enable != nil {
		body["enable"] = *server.Enable
	}
	if server.MTU != nil {
		body["mtu"] = *server.MTU
	}
	if server.Comment != "" {
		body["comment"] = server.Comment
	}

	// Type-specific fields
	switch server.Type {
	case MetricsServerTypeInfluxDBHTTP:
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
		if server.Timeout != nil {
			body["timeout"] = *server.Timeout
		}
	case MetricsServerTypeInfluxDBUDP:
		if server.Protocol != "" {
			body["proto"] = server.Protocol
		}
	}

	path := fmt.Sprintf("/config/metrics/server/%s", server.Type)
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

	if server.Server != "" {
		body["server"] = server.Server
	}
	if server.Port > 0 {
		body["port"] = server.Port
	}
	if server.Enable != nil {
		body["enable"] = *server.Enable
	}
	if server.MTU != nil {
		body["mtu"] = *server.MTU
	}
	if server.Comment != "" {
		body["comment"] = server.Comment
	}

	// Type-specific fields
	switch serverType {
	case MetricsServerTypeInfluxDBHTTP:
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
		if server.Timeout != nil {
			body["timeout"] = *server.Timeout
		}
	case MetricsServerTypeInfluxDBUDP:
		if server.Protocol != "" {
			body["proto"] = server.Protocol
		}
	}

	path := fmt.Sprintf("/config/metrics/server/%s/%s", serverType, url.PathEscape(name))
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

	path := fmt.Sprintf("/config/metrics/server/%s/%s", serverType, url.PathEscape(name))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete metrics server %s: %w", name, err)
	}

	return nil
}
