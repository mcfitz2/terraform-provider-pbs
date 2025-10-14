/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package jobs provides API client functionality for PBS job configurations
package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Client represents the jobs API client
type Client struct {
	api *api.Client
}

// NewClient creates a new jobs API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// Prune Job Types

// PruneJob represents a prune job configuration
type PruneJob struct {
	ID          string `json:"id"`
	Store       string `json:"store"`
	Schedule    string `json:"schedule"`
	KeepLast    *int   `json:"keep-last,omitempty"`
	KeepHourly  *int   `json:"keep-hourly,omitempty"`
	KeepDaily   *int   `json:"keep-daily,omitempty"`
	KeepWeekly  *int   `json:"keep-weekly,omitempty"`
	KeepMonthly *int   `json:"keep-monthly,omitempty"`
	KeepYearly  *int   `json:"keep-yearly,omitempty"`
	MaxDepth    *int   `json:"max-depth,omitempty"`
	NamespaceRE string `json:"ns,omitempty"`
	BackupType  string `json:"backup-type,omitempty"` // vm, ct, host
	BackupID    string `json:"backup-id,omitempty"`
	Comment     string `json:"comment,omitempty"`
	// Disable field removed in PBS 4.0
}

// ListPruneJobs lists all prune job configurations
func (c *Client) ListPruneJobs(ctx context.Context) ([]PruneJob, error) {
	resp, err := c.api.Get(ctx, "/config/prune")
	if err != nil {
		return nil, fmt.Errorf("failed to list prune jobs: %w", err)
	}

	var jobs []PruneJob
	if err := json.Unmarshal(resp.Data, &jobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prune jobs: %w", err)
	}

	return jobs, nil
}

// GetPruneJob gets a specific prune job by ID
func (c *Client) GetPruneJob(ctx context.Context, id string) (*PruneJob, error) {
	path := fmt.Sprintf("/config/prune/%s", url.PathEscape(id))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get prune job %s: %w", id, err)
	}

	var job PruneJob
	if err := json.Unmarshal(resp.Data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal prune job %s: %w", id, err)
	}

	return &job, nil
}

// CreatePruneJob creates a new prune job
func (c *Client) CreatePruneJob(ctx context.Context, job *PruneJob) error {
	if job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	if job.Store == "" {
		return fmt.Errorf("datastore is required")
	}
	if job.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}

	body := map[string]interface{}{
		"id":       job.ID,
		"store":    job.Store,
		"schedule": job.Schedule,
	}

	if job.KeepLast != nil {
		body["keep-last"] = *job.KeepLast
	}
	if job.KeepHourly != nil {
		body["keep-hourly"] = *job.KeepHourly
	}
	if job.KeepDaily != nil {
		body["keep-daily"] = *job.KeepDaily
	}
	if job.KeepWeekly != nil {
		body["keep-weekly"] = *job.KeepWeekly
	}
	if job.KeepMonthly != nil {
		body["keep-monthly"] = *job.KeepMonthly
	}
	if job.KeepYearly != nil {
		body["keep-yearly"] = *job.KeepYearly
	}
	if job.MaxDepth != nil {
		body["max-depth"] = *job.MaxDepth
	}
	if job.NamespaceRE != "" {
		body["ns"] = job.NamespaceRE
	}
	if job.BackupType != "" {
		body["backup-type"] = job.BackupType
	}
	if job.BackupID != "" {
		body["backup-id"] = job.BackupID
	}
	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	// Disable field removed in PBS 4.0

	_, err := c.api.Post(ctx, "/config/prune", body)
	if err != nil {
		return fmt.Errorf("failed to create prune job %s: %w", job.ID, err)
	}

	return nil
}

// UpdatePruneJob updates an existing prune job
func (c *Client) UpdatePruneJob(ctx context.Context, id string, job *PruneJob) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	body := map[string]interface{}{}

	if job.Store != "" {
		body["store"] = job.Store
	}
	if job.Schedule != "" {
		body["schedule"] = job.Schedule
	}
	if job.KeepLast != nil {
		body["keep-last"] = *job.KeepLast
	}
	if job.KeepHourly != nil {
		body["keep-hourly"] = *job.KeepHourly
	}
	if job.KeepDaily != nil {
		body["keep-daily"] = *job.KeepDaily
	}
	if job.KeepWeekly != nil {
		body["keep-weekly"] = *job.KeepWeekly
	}
	if job.KeepMonthly != nil {
		body["keep-monthly"] = *job.KeepMonthly
	}
	if job.KeepYearly != nil {
		body["keep-yearly"] = *job.KeepYearly
	}
	if job.MaxDepth != nil {
		body["max-depth"] = *job.MaxDepth
	}
	if job.NamespaceRE != "" {
		body["ns"] = job.NamespaceRE
	}
	if job.BackupType != "" {
		body["backup-type"] = job.BackupType
	}
	if job.BackupID != "" {
		body["backup-id"] = job.BackupID
	}
	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	// Disable field removed in PBS 4.0

	path := fmt.Sprintf("/config/prune/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update prune job %s: %w", id, err)
	}

	return nil
}

// DeletePruneJob deletes a prune job
func (c *Client) DeletePruneJob(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	path := fmt.Sprintf("/config/prune/%s", url.PathEscape(id))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete prune job %s: %w", id, err)
	}

	return nil
}

// Sync Job Types

// SyncJob represents a sync job configuration
type SyncJob struct {
	ID             string   `json:"id"`
	Store          string   `json:"store"`
	Remote         string   `json:"remote"`
	RemoteStore    string   `json:"remote-store"`
	Schedule       string   `json:"schedule"`
	NamespaceRE    string   `json:"ns,omitempty"`
	MaxDepth       *int     `json:"max-depth,omitempty"`
	GroupFilter    []string `json:"group-filter,omitempty"`
	RemoveVanished *bool    `json:"remove-vanished,omitempty"`
	Comment        string   `json:"comment,omitempty"`
	// Disable field removed in PBS 4.0
	Owner      string  `json:"owner,omitempty"`
	RateLimitIn string `json:"rate-in,omitempty"` // PBS 4.0 expects byte size string format (e.g., "10M", "1G")
}

// ListSyncJobs lists all sync job configurations
func (c *Client) ListSyncJobs(ctx context.Context) ([]SyncJob, error) {
	resp, err := c.api.Get(ctx, "/config/sync")
	if err != nil {
		return nil, fmt.Errorf("failed to list sync jobs: %w", err)
	}

	var jobs []SyncJob
	if err := json.Unmarshal(resp.Data, &jobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync jobs: %w", err)
	}

	return jobs, nil
}

// GetSyncJob gets a specific sync job by ID
func (c *Client) GetSyncJob(ctx context.Context, id string) (*SyncJob, error) {
	path := fmt.Sprintf("/config/sync/%s", url.PathEscape(id))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get sync job %s: %w", id, err)
	}

	var job SyncJob
	if err := json.Unmarshal(resp.Data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sync job %s: %w", id, err)
	}

	return &job, nil
}

// CreateSyncJob creates a new sync job
func (c *Client) CreateSyncJob(ctx context.Context, job *SyncJob) error {
	if job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	if job.Store == "" {
		return fmt.Errorf("datastore is required")
	}
	if job.Remote == "" {
		return fmt.Errorf("remote is required")
	}
	if job.RemoteStore == "" {
		return fmt.Errorf("remote datastore is required")
	}
	if job.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}

	body := map[string]interface{}{
		"id":           job.ID,
		"store":        job.Store,
		"remote":       job.Remote,
		"remote-store": job.RemoteStore,
		"schedule":     job.Schedule,
	}

	if job.NamespaceRE != "" {
		body["ns"] = job.NamespaceRE
	}
	if job.MaxDepth != nil {
		body["max-depth"] = *job.MaxDepth
	}
	if len(job.GroupFilter) > 0 {
		body["group-filter"] = job.GroupFilter
	}
	if job.RemoveVanished != nil {
		body["remove-vanished"] = *job.RemoveVanished
	}
	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	// Disable field removed in PBS 4.0
	if job.Owner != "" {
		body["owner"] = job.Owner
	}
	if job.RateLimitIn != "" {
		body["rate-in"] = job.RateLimitIn
	}

	_, err := c.api.Post(ctx, "/config/sync", body)
	if err != nil {
		return fmt.Errorf("failed to create sync job %s: %w", job.ID, err)
	}

	return nil
}

// UpdateSyncJob updates an existing sync job
func (c *Client) UpdateSyncJob(ctx context.Context, id string, job *SyncJob) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	body := map[string]interface{}{}

	if job.Store != "" {
		body["store"] = job.Store
	}
	if job.Remote != "" {
		body["remote"] = job.Remote
	}
	if job.RemoteStore != "" {
		body["remote-store"] = job.RemoteStore
	}
	if job.Schedule != "" {
		body["schedule"] = job.Schedule
	}
	if job.NamespaceRE != "" {
		body["ns"] = job.NamespaceRE
	}
	if job.MaxDepth != nil {
		body["max-depth"] = *job.MaxDepth
	}
	if len(job.GroupFilter) > 0 {
		body["group-filter"] = job.GroupFilter
	}
	if job.RemoveVanished != nil {
		body["remove-vanished"] = *job.RemoveVanished
	}
	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	// Disable field removed in PBS 4.0
	if job.Owner != "" {
		body["owner"] = job.Owner
	}
	if job.RateLimitIn != "" {
		body["rate-in"] = job.RateLimitIn
	}

	path := fmt.Sprintf("/config/sync/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update sync job %s: %w", id, err)
	}

	return nil
}

// DeleteSyncJob deletes a sync job
func (c *Client) DeleteSyncJob(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	path := fmt.Sprintf("/config/sync/%s", url.PathEscape(id))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete sync job %s: %w", id, err)
	}

	return nil
}

// Verify Job Types

// VerifyJob represents a verification job configuration
type VerifyJob struct {
	ID             string `json:"id"`
	Store          string `json:"store"`
	Schedule       string `json:"schedule"`
	IgnoreVerified *bool  `json:"ignore-verified,omitempty"`
	OutdatedAfter  *int   `json:"outdated-after,omitempty"` // days
	NamespaceRE    string `json:"ns,omitempty"`
	MaxDepth       *int   `json:"max-depth,omitempty"`
	Comment        string `json:"comment,omitempty"`
	// Disable field removed in PBS 4.0
}

// ListVerifyJobs lists all verify job configurations
func (c *Client) ListVerifyJobs(ctx context.Context) ([]VerifyJob, error) {
	resp, err := c.api.Get(ctx, "/config/verify")
	if err != nil {
		return nil, fmt.Errorf("failed to list verify jobs: %w", err)
	}

	var jobs []VerifyJob
	if err := json.Unmarshal(resp.Data, &jobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify jobs: %w", err)
	}

	return jobs, nil
}

// GetVerifyJob gets a specific verify job by ID
func (c *Client) GetVerifyJob(ctx context.Context, id string) (*VerifyJob, error) {
	path := fmt.Sprintf("/config/verify/%s", url.PathEscape(id))
	resp, err := c.api.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get verify job %s: %w", id, err)
	}

	var job VerifyJob
	if err := json.Unmarshal(resp.Data, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal verify job %s: %w", id, err)
	}

	return &job, nil
}

// CreateVerifyJob creates a new verify job
func (c *Client) CreateVerifyJob(ctx context.Context, job *VerifyJob) error {
	if job.ID == "" {
		return fmt.Errorf("job ID is required")
	}
	if job.Store == "" {
		return fmt.Errorf("datastore is required")
	}
	if job.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}

	body := map[string]interface{}{
		"id":       job.ID,
		"store":    job.Store,
		"schedule": job.Schedule,
	}

	if job.IgnoreVerified != nil {
		body["ignore-verified"] = *job.IgnoreVerified
	}
	if job.OutdatedAfter != nil {
		body["outdated-after"] = *job.OutdatedAfter
	}
	if job.NamespaceRE != "" {
		body["ns"] = job.NamespaceRE
	}
	if job.MaxDepth != nil {
		body["max-depth"] = *job.MaxDepth
	}
	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	// Disable field removed in PBS 4.0

	_, err := c.api.Post(ctx, "/config/verify", body)
	if err != nil {
		return fmt.Errorf("failed to create verify job %s: %w", job.ID, err)
	}

	return nil
}

// UpdateVerifyJob updates an existing verify job
func (c *Client) UpdateVerifyJob(ctx context.Context, id string, job *VerifyJob) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	body := map[string]interface{}{}

	if job.Store != "" {
		body["store"] = job.Store
	}
	if job.Schedule != "" {
		body["schedule"] = job.Schedule
	}
	if job.IgnoreVerified != nil {
		body["ignore-verified"] = *job.IgnoreVerified
	}
	if job.OutdatedAfter != nil {
		body["outdated-after"] = *job.OutdatedAfter
	}
	if job.NamespaceRE != "" {
		body["ns"] = job.NamespaceRE
	}
	if job.MaxDepth != nil {
		body["max-depth"] = *job.MaxDepth
	}
	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	// Disable field removed in PBS 4.0

	path := fmt.Sprintf("/config/verify/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update verify job %s: %w", id, err)
	}

	return nil
}

// DeleteVerifyJob deletes a verify job
func (c *Client) DeleteVerifyJob(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	path := fmt.Sprintf("/config/verify/%s", url.PathEscape(id))
	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete verify job %s: %w", id, err)
	}

	return nil
}
