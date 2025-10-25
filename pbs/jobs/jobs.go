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
	ID          string   `json:"id"`
	Store       string   `json:"store"`
	Schedule    string   `json:"schedule"`
	KeepLast    *int     `json:"keep-last,omitempty"`
	KeepHourly  *int     `json:"keep-hourly,omitempty"`
	KeepDaily   *int     `json:"keep-daily,omitempty"`
	KeepWeekly  *int     `json:"keep-weekly,omitempty"`
	KeepMonthly *int     `json:"keep-monthly,omitempty"`
	KeepYearly  *int     `json:"keep-yearly,omitempty"`
	MaxDepth    *int     `json:"max-depth,omitempty"`
	Namespace   string   `json:"ns,omitempty"`
	Comment     string   `json:"comment,omitempty"`
	Disable     *bool    `json:"disable,omitempty"`
	Digest      string   `json:"digest,omitempty"`
	Delete      []string `json:"delete,omitempty"`
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

	setInt := func(key string, value *int) {
		if value != nil {
			body[key] = *value
		}
	}

	if job.Namespace != "" {
		body["ns"] = job.Namespace
	}

	setInt("keep-last", job.KeepLast)
	setInt("keep-hourly", job.KeepHourly)
	setInt("keep-daily", job.KeepDaily)
	setInt("keep-weekly", job.KeepWeekly)
	setInt("keep-monthly", job.KeepMonthly)
	setInt("keep-yearly", job.KeepYearly)
	setInt("max-depth", job.MaxDepth)

	if job.Comment != "" {
		body["comment"] = job.Comment
	}
	if job.Disable != nil {
		body["disable"] = *job.Disable
	}

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

	setString := func(key, value string) {
		if value != "" {
			body[key] = value
		}
	}

	setInt := func(key string, value *int) {
		if value != nil {
			body[key] = *value
		}
	}

	setBool := func(key string, value *bool) {
		if value != nil {
			body[key] = *value
		}
	}

	setString("store", job.Store)
	setString("schedule", job.Schedule)
	setString("ns", job.Namespace)
	setString("comment", job.Comment)

	setInt("keep-last", job.KeepLast)
	setInt("keep-hourly", job.KeepHourly)
	setInt("keep-daily", job.KeepDaily)
	setInt("keep-weekly", job.KeepWeekly)
	setInt("keep-monthly", job.KeepMonthly)
	setInt("keep-yearly", job.KeepYearly)
	setInt("max-depth", job.MaxDepth)

	setBool("disable", job.Disable)

	if len(job.Delete) > 0 {
		body["delete"] = job.Delete
	}
	if job.Digest != "" {
		body["digest"] = job.Digest
	}

	path := fmt.Sprintf("/config/prune/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update prune job %s: %w", id, err)
	}

	return nil
}

// DeletePruneJob deletes a prune job
func (c *Client) DeletePruneJob(ctx context.Context, id, digest string) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	path := fmt.Sprintf("/config/prune/%s", url.PathEscape(id))
	if digest != "" {
		path = fmt.Sprintf("%s?digest=%s", path, url.QueryEscape(digest))
	}

	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete prune job %s: %w", id, err)
	}

	return nil
}

// Sync Job Types

// SyncJob represents a sync job configuration
type SyncJob struct {
	ID              string   `json:"id"`
	Store           string   `json:"store"`
	Remote          string   `json:"remote"`
	RemoteStore     string   `json:"remote-store"`
	RemoteNamespace string   `json:"remote-ns,omitempty"`
	Schedule        string   `json:"schedule"`
	Namespace       string   `json:"ns,omitempty"`
	MaxDepth        *int     `json:"max-depth,omitempty"`
	GroupFilter     []string `json:"group-filter,omitempty"`
	RemoveVanished  *bool    `json:"remove-vanished,omitempty"`
	ResyncCorrupt   *bool    `json:"resync-corrupt,omitempty"`
	EncryptedOnly   *bool    `json:"encrypted-only,omitempty"`
	VerifiedOnly    *bool    `json:"verified-only,omitempty"`
	RunOnMount      *bool    `json:"run-on-mount,omitempty"`
	TransferLast    *int     `json:"transfer-last,omitempty"`
	SyncDirection   string   `json:"sync-direction,omitempty"`
	Comment         string   `json:"comment,omitempty"`
	Owner           string   `json:"owner,omitempty"`
	RateIn          string   `json:"rate-in,omitempty"`
	RateOut         string   `json:"rate-out,omitempty"`
	BurstIn         string   `json:"burst-in,omitempty"`
	BurstOut        string   `json:"burst-out,omitempty"`
	Digest          string   `json:"digest,omitempty"`
	Delete          []string `json:"delete,omitempty"`
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

	setInt := func(key string, value *int) {
		if value != nil {
			body[key] = *value
		}
	}

	setBool := func(key string, value *bool) {
		if value != nil {
			body[key] = *value
		}
	}

	setString := func(key, value string) {
		if value != "" {
			body[key] = value
		}
	}

	setString("remote-ns", job.RemoteNamespace)
	setString("ns", job.Namespace)
	setInt("max-depth", job.MaxDepth)

	if len(job.GroupFilter) > 0 {
		body["group-filter"] = job.GroupFilter
	}

	setBool("remove-vanished", job.RemoveVanished)
	setBool("resync-corrupt", job.ResyncCorrupt)
	setBool("encrypted-only", job.EncryptedOnly)
	setBool("verified-only", job.VerifiedOnly)
	setBool("run-on-mount", job.RunOnMount)

	setInt("transfer-last", job.TransferLast)

	setString("sync-direction", job.SyncDirection)
	setString("comment", job.Comment)
	setString("owner", job.Owner)
	setString("rate-in", job.RateIn)
	setString("rate-out", job.RateOut)
	setString("burst-in", job.BurstIn)
	setString("burst-out", job.BurstOut)

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

	setString := func(key, value string) {
		if value != "" {
			body[key] = value
		}
	}

	setInt := func(key string, value *int) {
		if value != nil {
			body[key] = *value
		}
	}

	setBool := func(key string, value *bool) {
		if value != nil {
			body[key] = *value
		}
	}

	setString("store", job.Store)
	setString("remote", job.Remote)
	setString("remote-store", job.RemoteStore)
	setString("remote-ns", job.RemoteNamespace)
	setString("schedule", job.Schedule)
	setString("ns", job.Namespace)
	setString("comment", job.Comment)
	setString("owner", job.Owner)
	setString("sync-direction", job.SyncDirection)
	setString("rate-in", job.RateIn)
	setString("rate-out", job.RateOut)
	setString("burst-in", job.BurstIn)
	setString("burst-out", job.BurstOut)

	setInt("max-depth", job.MaxDepth)
	setInt("transfer-last", job.TransferLast)

	if len(job.GroupFilter) > 0 {
		body["group-filter"] = job.GroupFilter
	}

	setBool("remove-vanished", job.RemoveVanished)
	setBool("resync-corrupt", job.ResyncCorrupt)
	setBool("encrypted-only", job.EncryptedOnly)
	setBool("verified-only", job.VerifiedOnly)
	setBool("run-on-mount", job.RunOnMount)

	if len(job.Delete) > 0 {
		body["delete"] = job.Delete
	}
	if job.Digest != "" {
		body["digest"] = job.Digest
	}

	path := fmt.Sprintf("/config/sync/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update sync job %s: %w", id, err)
	}

	return nil
}

// DeleteSyncJob deletes a sync job
func (c *Client) DeleteSyncJob(ctx context.Context, id, digest string) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	path := fmt.Sprintf("/config/sync/%s", url.PathEscape(id))
	if digest != "" {
		path = fmt.Sprintf("%s?digest=%s", path, url.QueryEscape(digest))
	}

	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete sync job %s: %w", id, err)
	}

	return nil
}

// Verify Job Types

// VerifyJob represents a verification job configuration
type VerifyJob struct {
	ID             string   `json:"id"`
	Store          string   `json:"store"`
	Schedule       string   `json:"schedule"`
	IgnoreVerified *bool    `json:"ignore-verified,omitempty"`
	OutdatedAfter  *int     `json:"outdated-after,omitempty"`
	Namespace      string   `json:"ns,omitempty"`
	MaxDepth       *int     `json:"max-depth,omitempty"`
	Comment        string   `json:"comment,omitempty"`
	Digest         string   `json:"digest,omitempty"`
	Delete         []string `json:"delete,omitempty"`
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

	setBool := func(key string, value *bool) {
		if value != nil {
			body[key] = *value
		}
	}

	setInt := func(key string, value *int) {
		if value != nil {
			body[key] = *value
		}
	}

	if job.Namespace != "" {
		body["ns"] = job.Namespace
	}

	setBool("ignore-verified", job.IgnoreVerified)
	setInt("outdated-after", job.OutdatedAfter)
	setInt("max-depth", job.MaxDepth)

	if job.Comment != "" {
		body["comment"] = job.Comment
	}

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

	setString := func(key, value string) {
		if value != "" {
			body[key] = value
		}
	}

	setBool := func(key string, value *bool) {
		if value != nil {
			body[key] = *value
		}
	}

	setInt := func(key string, value *int) {
		if value != nil {
			body[key] = *value
		}
	}

	setString("store", job.Store)
	setString("schedule", job.Schedule)
	setString("ns", job.Namespace)
	setString("comment", job.Comment)

	setBool("ignore-verified", job.IgnoreVerified)

	setInt("outdated-after", job.OutdatedAfter)
	setInt("max-depth", job.MaxDepth)

	if len(job.Delete) > 0 {
		body["delete"] = job.Delete
	}
	if job.Digest != "" {
		body["digest"] = job.Digest
	}

	path := fmt.Sprintf("/config/verify/%s", url.PathEscape(id))
	_, err := c.api.Put(ctx, path, body)
	if err != nil {
		return fmt.Errorf("failed to update verify job %s: %w", id, err)
	}

	return nil
}

// DeleteVerifyJob deletes a verify job
func (c *Client) DeleteVerifyJob(ctx context.Context, id, digest string) error {
	if id == "" {
		return fmt.Errorf("job ID is required")
	}

	path := fmt.Sprintf("/config/verify/%s", url.PathEscape(id))
	if digest != "" {
		path = fmt.Sprintf("%s?digest=%s", path, url.QueryEscape(digest))
	}

	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete verify job %s: %w", id, err)
	}

	return nil
}
