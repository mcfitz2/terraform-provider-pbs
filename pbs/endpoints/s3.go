/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package endpoints provides API client functionality for PBS endpoint configurations
package endpoints

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Client represents the endpoints API client
type Client struct {
	api *api.Client
}

// NewClient creates a new endpoints API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// S3Endpoint represents an S3 endpoint configuration
type S3Endpoint struct {
	ID             string   `json:"id"`
	AccessKey      string   `json:"access-key"`
	SecretKey      string   `json:"secret-key,omitempty"`
	Endpoint       string   `json:"endpoint"`
	Region         string   `json:"region,omitempty"`
	Fingerprint    string   `json:"fingerprint,omitempty"`
	Port           *int     `json:"port,omitempty"`
	PathStyle      *bool    `json:"path-style,omitempty"`
	ProviderQuirks []string `json:"provider-quirks,omitempty"`
	PutRateLimit   *int     `json:"put-rate-limit,omitempty"`
}

// ListS3Endpoints lists all S3 endpoint configurations
func (c *Client) ListS3Endpoints(ctx context.Context) ([]S3Endpoint, error) {
	resp, err := c.api.Get(ctx, "/config/s3")
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 endpoints: %w", err)
	}

	var endpoints []S3Endpoint
	if err := json.Unmarshal(resp.Data, &endpoints); err != nil {
		return nil, fmt.Errorf("failed to unmarshal S3 endpoints: %w", err)
	}

	return endpoints, nil
}

// GetS3Endpoint gets a specific S3 endpoint configuration by ID
func (c *Client) GetS3Endpoint(ctx context.Context, id string) (*S3Endpoint, error) {
	path := fmt.Sprintf("/config/s3/%s", url.PathEscape(id))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 endpoint %s: %w", id, err)
	}

	var endpoint S3Endpoint
	if err := json.Unmarshal(resp.Data, &endpoint); err != nil {
		return nil, fmt.Errorf("failed to unmarshal S3 endpoint %s: %w", id, err)
	}

	return &endpoint, nil
}

// CreateS3Endpoint creates a new S3 endpoint configuration
func (c *Client) CreateS3Endpoint(ctx context.Context, endpoint *S3Endpoint) error {
	if endpoint.ID == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{
		"id":         endpoint.ID,
		"access-key": endpoint.AccessKey,
		"secret-key": endpoint.SecretKey,
		"endpoint":   endpoint.Endpoint,
	}

	if endpoint.Region != "" {
		body["region"] = endpoint.Region
	}
	if endpoint.Fingerprint != "" {
		body["fingerprint"] = endpoint.Fingerprint
	}
	if endpoint.Port != nil {
		body["port"] = *endpoint.Port
	}
	if endpoint.PathStyle != nil {
		body["path-style"] = *endpoint.PathStyle
	}
	if len(endpoint.ProviderQuirks) > 0 {
		body["provider-quirks"] = endpoint.ProviderQuirks
	}
	if endpoint.PutRateLimit != nil {
		body["put-rate-limit"] = *endpoint.PutRateLimit
	}

	_, err := c.api.Post(ctx, "/config/s3", body)
	if err != nil {
		return fmt.Errorf("failed to create S3 endpoint %s: %w", endpoint.ID, err)
	}

	return nil
}

// UpdateS3Endpoint updates an existing S3 endpoint configuration
func (c *Client) UpdateS3Endpoint(ctx context.Context, id string, endpoint *S3Endpoint) error {
	if id == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{}

	if endpoint.AccessKey != "" {
		body["access-key"] = endpoint.AccessKey
	}
	if endpoint.SecretKey != "" {
		body["secret-key"] = endpoint.SecretKey
	}
	if endpoint.Endpoint != "" {
		body["endpoint"] = endpoint.Endpoint
	}
	if endpoint.Region != "" {
		body["region"] = endpoint.Region
	}
	if endpoint.Fingerprint != "" {
		body["fingerprint"] = endpoint.Fingerprint
	}
	if endpoint.Port != nil {
		body["port"] = *endpoint.Port
	}
	if endpoint.PathStyle != nil {
		body["path-style"] = *endpoint.PathStyle
	}
	if len(endpoint.ProviderQuirks) > 0 {
		body["provider-quirks"] = endpoint.ProviderQuirks
	}
	if endpoint.PutRateLimit != nil {
		body["put-rate-limit"] = *endpoint.PutRateLimit
	}

	path := fmt.Sprintf("/config/s3/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update S3 endpoint %s: %w", id, err)
	}

	return nil
}

// DeleteS3Endpoint deletes an S3 endpoint configuration
func (c *Client) DeleteS3Endpoint(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("endpoint ID is required")
	}

	path := fmt.Sprintf("/config/s3/%s", url.PathEscape(id))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete S3 endpoint %s: %w", id, err)
	}

	return nil
}
