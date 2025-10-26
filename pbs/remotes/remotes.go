/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package remotes provides API client functionality for PBS remote configurations
package remotes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Client represents the remotes API client
type Client struct {
	api *api.Client
}

// NewClient creates a new remotes API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// Remote represents a PBS remote configuration
type Remote struct {
	Name        string   `json:"name"`
	Host        string   `json:"host"`
	Port        *int     `json:"port,omitempty"`
	AuthID      string   `json:"auth-id"`
	Password    string   `json:"password,omitempty"`
	Fingerprint string   `json:"fingerprint,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Delete      []string `json:"delete,omitempty"`
}

// RemoteDatastore represents a datastore on a remote PBS server
type RemoteDatastore struct {
	Name        string              `json:"name"`
	Comment     string              `json:"comment,omitempty"`
	Maintenance *MaintenanceInfo    `json:"maintenance,omitempty"`
}

// MaintenanceInfo represents datastore maintenance status
type MaintenanceInfo struct {
	Type    string `json:"type"`
	Message string `json:"message,omitempty"`
}

// RemoteNamespace represents a namespace on a remote datastore
type RemoteNamespace struct {
	Namespace string `json:"ns"`
	Comment   string `json:"comment,omitempty"`
}

// RemoteGroup represents a backup group on a remote datastore
type RemoteGroup struct {
	BackupType  string   `json:"backup-type"`
	BackupID    string   `json:"backup-id"`
	BackupCount int      `json:"backup-count,omitempty"`
	LastBackup  int64    `json:"last-backup,omitempty"`
	Owner       string   `json:"owner,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Files       []string `json:"files,omitempty"`
}

// ListRemotes lists all remote configurations
func (c *Client) ListRemotes(ctx context.Context) ([]Remote, error) {
	resp, err := c.api.Get(ctx, "/config/remote")
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	var remotes []Remote
	if err := json.Unmarshal(resp.Data, &remotes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remotes list response: %w", err)
	}

	return remotes, nil
}

// GetRemote gets a specific remote configuration by name
func (c *Client) GetRemote(ctx context.Context, name string) (*Remote, error) {
	path := fmt.Sprintf("/config/remote/%s", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get remote %s: %w", name, err)
	}

	var remote Remote
	if err := json.Unmarshal(resp.Data, &remote); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote %s: %w", name, err)
	}

	return &remote, nil
}

// CreateRemote creates a new remote configuration
func (c *Client) CreateRemote(ctx context.Context, remote *Remote) error {
	if remote.Name == "" {
		return fmt.Errorf("remote name is required")
	}
	if remote.Host == "" {
		return fmt.Errorf("remote host is required")
	}
	if remote.AuthID == "" {
		return fmt.Errorf("remote auth-id is required")
	}
	if remote.Password == "" {
		return fmt.Errorf("remote password is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{
		"name":    remote.Name,
		"host":    remote.Host,
		"auth-id": remote.AuthID,
		"password": remote.Password,
	}

	if remote.Port != nil {
		body["port"] = *remote.Port
	}
	if remote.Fingerprint != "" {
		body["fingerprint"] = remote.Fingerprint
	}
	if remote.Comment != "" {
		body["comment"] = remote.Comment
	}

	_, err := c.api.Post(ctx, "/config/remote", body)
	if err != nil {
		return fmt.Errorf("failed to create remote %s: %w", remote.Name, err)
	}

	return nil
}

// UpdateRemote updates an existing remote configuration
func (c *Client) UpdateRemote(ctx context.Context, name string, remote *Remote) error {
	if name == "" {
		return fmt.Errorf("remote name is required")
	}

	// Convert struct to map for API request
	body := map[string]interface{}{}

	// Always send these if they're set
	if remote.Host != "" {
		body["host"] = remote.Host
	}
	if remote.AuthID != "" {
		body["auth-id"] = remote.AuthID
	}
	if remote.Password != "" {
		body["password"] = remote.Password
	}
	
	// Optional fields
	if remote.Port != nil {
		body["port"] = *remote.Port
	}
	if remote.Fingerprint != "" {
		body["fingerprint"] = remote.Fingerprint
	}
	if remote.Comment != "" {
		body["comment"] = remote.Comment
	}

	// Include digest for optimistic locking
	if remote.Digest != "" {
		body["digest"] = remote.Digest
	}

	// Include delete array for clearing optional fields
	if len(remote.Delete) > 0 {
		body["delete"] = remote.Delete
	}

	path := fmt.Sprintf("/config/remote/%s", url.PathEscape(name))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update remote %s: %w", name, err)
	}

	return nil
}

// DeleteRemote deletes a remote configuration
func (c *Client) DeleteRemote(ctx context.Context, name string, digest string) error {
	if name == "" {
		return fmt.Errorf("remote name is required")
	}

	path := fmt.Sprintf("/config/remote/%s", url.PathEscape(name))
	
	// Include digest as query parameter if provided
	if digest != "" {
		path = fmt.Sprintf("%s?digest=%s", path, url.QueryEscape(digest))
	}

	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete remote %s: %w", name, err)
	}

	return nil
}

// ListRemoteStores lists all datastores available on a remote PBS server
func (c *Client) ListRemoteStores(ctx context.Context, name string) ([]RemoteDatastore, error) {
	if name == "" {
		return nil, fmt.Errorf("remote name is required")
	}

	path := fmt.Sprintf("/config/remote/%s/scan", url.PathEscape(name))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote stores for %s: %w", name, err)
	}

	var stores []RemoteDatastore
	if err := json.Unmarshal(resp.Data, &stores); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote stores for %s: %w", name, err)
	}

	return stores, nil
}

// ListRemoteNamespaces lists all namespaces in a remote datastore
func (c *Client) ListRemoteNamespaces(ctx context.Context, name string, store string) ([]RemoteNamespace, error) {
	if name == "" {
		return nil, fmt.Errorf("remote name is required")
	}
	if store == "" {
		return nil, fmt.Errorf("store name is required")
	}

	path := fmt.Sprintf("/config/remote/%s/scan/%s/namespaces", url.PathEscape(name), url.PathEscape(store))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote namespaces for %s/%s: %w", name, store, err)
	}

	var namespaces []RemoteNamespace
	if err := json.Unmarshal(resp.Data, &namespaces); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote namespaces for %s/%s: %w", name, store, err)
	}

	return namespaces, nil
}

// ListRemoteGroups lists all backup groups in a remote datastore, optionally filtered by namespace
func (c *Client) ListRemoteGroups(ctx context.Context, name string, store string, namespace string) ([]RemoteGroup, error) {
	if name == "" {
		return nil, fmt.Errorf("remote name is required")
	}
	if store == "" {
		return nil, fmt.Errorf("store name is required")
	}

	path := fmt.Sprintf("/config/remote/%s/scan/%s/groups", url.PathEscape(name), url.PathEscape(store))
	
	// Add namespace as query parameter if provided
	if namespace != "" {
		path = fmt.Sprintf("%s?namespace=%s", path, url.QueryEscape(namespace))
	}

	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to list remote groups for %s/%s: %w", name, store, err)
	}

	var groups []RemoteGroup
	if err := json.Unmarshal(resp.Data, &groups); err != nil {
		return nil, fmt.Errorf("failed to unmarshal remote groups for %s/%s: %w", name, store, err)
	}

	return groups, nil
}
