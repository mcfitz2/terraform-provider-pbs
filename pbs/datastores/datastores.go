/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package datastores provides API client functionality for PBS datastore configurations
package datastores

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// isDebugEnabled checks if debug logging should be enabled
func isDebugEnabled() bool {
	return os.Getenv("PBS_DEBUG") != "" || os.Getenv("TF_LOG") != ""
}

// Client represents the datastores API client
type Client struct {
	api *api.Client
}

// NewClient creates a new datastores API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// datastoreListItem represents a datastore in list responses (minimal info)
type datastoreListItem struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	MaintenanceMode string `json:"maintenance-mode,omitempty"`
}

// Datastore represents a PBS datastore configuration
type Datastore struct {
	Name          string `json:"name"`
	Path          string `json:"path,omitempty"`
	Comment       string `json:"comment,omitempty"`
	Disabled      *bool  `json:"disable,omitempty"`
	GCSchedule    string `json:"gc-schedule,omitempty"`
	PruneSchedule string `json:"prune-schedule,omitempty"`

	// Retention windows
	KeepLast    *int `json:"keep-last,omitempty"`
	KeepHourly  *int `json:"keep-hourly,omitempty"`
	KeepDaily   *int `json:"keep-daily,omitempty"`
	KeepWeekly  *int `json:"keep-weekly,omitempty"`
	KeepMonthly *int `json:"keep-monthly,omitempty"`
	KeepYearly  *int `json:"keep-yearly,omitempty"`

	// Maintenance and notification fields
	MaintenanceModeRaw string           `json:"maintenance-mode,omitempty"`
	MaintenanceMode    *MaintenanceMode `json:"-"`
	NotifyRaw          string           `json:"notify,omitempty"`
	Notify             *DatastoreNotify `json:"-"`
	NotifyUser         string           `json:"notify-user,omitempty"`
	NotificationMode   string           `json:"notification-mode,omitempty"`
	NotifyLevel        string           `json:"notify-level,omitempty"`

	// Verification and reuse toggles
	VerifyNew      *bool `json:"verify-new,omitempty"`
	ReuseDatastore *bool `json:"reuse-datastore,omitempty"`
	OverwriteInUse *bool `json:"overwrite-in-use,omitempty"`

	// Tuning options
	TuningRaw string           `json:"tuning,omitempty"`
	Tuning    *DatastoreTuning `json:"-"`

	// Advanced options
	Fingerprint   string `json:"fingerprint,omitempty"`
	BackingDevice string `json:"backing-device,omitempty"`

	// S3 backend options (stored as backend configuration)
	Backend     string `json:"backend,omitempty"` // e.g. "type=s3,client=endpoint_id,bucket=bucket_name"
	BackendType string `json:"-"`
	S3Client    string `json:"-"` // S3 endpoint ID (for easier access in Go code)
	S3Bucket    string `json:"-"`

	Digest string   `json:"digest,omitempty"`
	Delete []string `json:"delete,omitempty"`
}

// ListDatastores lists all datastore configurations
func (c *Client) ListDatastores(ctx context.Context) ([]Datastore, error) {
	resp, err := c.api.Get(ctx, "/config/datastore")
	if err != nil {
		return nil, fmt.Errorf("failed to list datastores: %w", err)
	}

	// Parse the list response which contains minimal datastore info
	var listItems []datastoreListItem
	if err := json.Unmarshal(resp.Data, &listItems); err != nil {
		return nil, fmt.Errorf("failed to unmarshal datastores list response: %w", err)
	}

	// For the list operation, we only return basic info
	// If detailed info is needed, GetDatastore should be called for individual items
	datastores := make([]Datastore, len(listItems))
	for i, item := range listItems {
		datastores[i] = Datastore{
			Name: item.Name,
			Path: item.Path,
		}
	}

	return datastores, nil
}

// GetDatastore gets a specific datastore configuration by name
func (c *Client) GetDatastore(ctx context.Context, name string) (*Datastore, error) {
	if name == "" {
		return nil, fmt.Errorf("datastore name is required")
	}

	// Try to get individual datastore details first
	escapedName := url.PathEscape(name)
	path := fmt.Sprintf("/config/datastore/%s", escapedName)

	// CRITICAL DEBUG: Log when we attempt to get datastore
	fmt.Fprintf(os.Stderr, "[PBS-DEBUG] GetDatastore: Attempting GET '%s' at %s\n", name, time.Now().Format(time.RFC3339Nano))

	// Log the GET request for debugging
	if isDebugEnabled() {
		tflog.Debug(ctx, "GetDatastore: Attempting GET", map[string]interface{}{
			"path":          path,
			"original_name": name,
		})
	}

	resp, getErr := c.api.Get(ctx, path)

	// CRITICAL DEBUG: Log the result
	if getErr != nil {
		fmt.Fprintf(os.Stderr, "[PBS-DEBUG] GetDatastore: GET '%s' FAILED at %s: %v\n", name, time.Now().Format(time.RFC3339Nano), getErr)
	} else {
		fmt.Fprintf(os.Stderr, "[PBS-DEBUG] GetDatastore: GET '%s' SUCCEEDED at %s\n", name, time.Now().Format(time.RFC3339Nano))
	}
	if getErr == nil {
		if isDebugEnabled() {
			tflog.Debug(ctx, "GetDatastore: GET succeeded", map[string]interface{}{
				"response_size": len(resp.Data),
			})
		}

		var ds Datastore
		if unmarshalErr := json.Unmarshal(resp.Data, &ds); unmarshalErr == nil {
			ds.Name = name // Ensure name is set
			if isDebugEnabled() {
				tflog.Debug(ctx, "GetDatastore: Successfully unmarshaled datastore", map[string]interface{}{
					"name": name,
				})
			}

			// Parse backend configuration if present
			c.parseBackendConfig(&ds)

			// Parse property string fields into typed structs
			ds.MaintenanceMode = parseMaintenanceMode(ds.MaintenanceModeRaw)
			if ds.MaintenanceMode != nil {
				ds.MaintenanceModeRaw = formatMaintenanceMode(ds.MaintenanceMode)
			}

			ds.Notify = parseNotify(ds.NotifyRaw)
			if ds.Notify != nil {
				ds.NotifyRaw = formatNotify(ds.Notify)
			}

			ds.Tuning = parseTuning(ds.TuningRaw)
			if ds.Tuning != nil {
				ds.TuningRaw = formatTuning(ds.Tuning)
			}

			return &ds, nil
		} else {
			if isDebugEnabled() {
				tflog.Debug(ctx, "GetDatastore: Unmarshal failed", map[string]interface{}{
					"error": unmarshalErr.Error(),
				})
			}
		}
	} else {
		if isDebugEnabled() {
			tflog.Debug(ctx, "GetDatastore: GET failed", map[string]interface{}{
				"error": getErr.Error(),
			})
		}
	}

	// If individual get fails, fall back to list and find
	// This handles cases where the direct endpoint might not work but the datastore exists
	if isDebugEnabled() {
		tflog.Debug(ctx, "GetDatastore: Falling back to list operation")
	}
	datastores, listErr := c.ListDatastores(ctx)
	if listErr != nil {
		// Both GET and LIST failed - return detailed error
		if isDebugEnabled() {
			tflog.Debug(ctx, "GetDatastore: List also failed", map[string]interface{}{
				"error": listErr.Error(),
			})
		}
		return nil, fmt.Errorf("failed to get datastore %s (GET error: %v, LIST error: %w)", name, getErr, listErr)
	}

	if isDebugEnabled() {
		tflog.Debug(ctx, "GetDatastore: List returned datastores", map[string]interface{}{
			"count": len(datastores),
		})
		for i, ds := range datastores {
			tflog.Debug(ctx, "GetDatastore: Datastore in list", map[string]interface{}{
				"index": i,
				"name":  ds.Name,
			})
		}
	}

	for _, ds := range datastores {
		if ds.Name == name {
			if isDebugEnabled() {
				tflog.Debug(ctx, "GetDatastore: Found in list (minimal data)", map[string]interface{}{
					"name": name,
				})
			}
			// NOTE: This returns incomplete data - only Name and Path are populated from list
			return &ds, nil
		}
	}

	// Datastore not found in list - include the original GET error for debugging
	if isDebugEnabled() {
		tflog.Debug(ctx, "GetDatastore: Datastore not found in list", map[string]interface{}{
			"name":       name,
			"list_count": len(datastores),
		})
	}
	if getErr != nil {
		return nil, fmt.Errorf("datastore %s not found (GET error: %v)", name, getErr)
	}
	return nil, fmt.Errorf("datastore %s not found", name)
}

// CreateDatastore creates a new datastore configuration
func (c *Client) CreateDatastore(ctx context.Context, datastore *Datastore) error {
	if datastore.Name == "" {
		return fmt.Errorf("datastore name is required")
	}

	if isDebugEnabled() {
		tflog.Debug(ctx, "CreateDatastore: Starting creation", map[string]interface{}{
			"name": datastore.Name,
		})
	}

	// Convert struct to map for API request
	body := c.datastoreToMap(datastore)

	// Creating datastore with PBS API
	resp, err := c.api.Post(ctx, "/config/datastore", body)
	if err != nil {
		if isDebugEnabled() {
			tflog.Debug(ctx, "CreateDatastore: POST failed", map[string]interface{}{
				"name":  datastore.Name,
				"error": err.Error(),
			})
		}
		return fmt.Errorf("failed to create datastore %s: %w", datastore.Name, err)
	}

	// Parse the UPID from the response
	var upid string
	if err := json.Unmarshal(resp.Data, &upid); err != nil {
		if isDebugEnabled() {
			tflog.Debug(ctx, "CreateDatastore: Failed to parse UPID", map[string]interface{}{
				"error": err.Error(),
			})
		}
		return fmt.Errorf("failed to parse UPID from response: %w", err)
	}

	if isDebugEnabled() {
		tflog.Debug(ctx, "CreateDatastore: Got UPID", map[string]interface{}{
			"name": datastore.Name,
			"upid": upid,
		})
	}

	// Get the node name from the UPID or by querying nodes
	node, err := c.getNodeForTask(ctx, upid)
	if err != nil {
		if isDebugEnabled() {
			tflog.Debug(ctx, "CreateDatastore: Failed to get node", map[string]interface{}{
				"upid":  upid,
				"error": err.Error(),
			})
		}
		return fmt.Errorf("failed to determine node for task: %w", err)
	}

	if isDebugEnabled() {
		tflog.Debug(ctx, "CreateDatastore: Waiting for task", map[string]interface{}{
			"node": node,
			"upid": upid,
		})
	}

	// Wait for the task to complete with a reasonable timeout
	// For S3 datastores, this involves file I/O which can take time on slow connections
	// Wait for task completion
	if err := c.api.WaitForTask(ctx, node, upid, 5*time.Minute); err != nil {
		if isDebugEnabled() {
			tflog.Debug(ctx, "CreateDatastore: Task failed", map[string]interface{}{
				"upid":  upid,
				"error": err.Error(),
			})
		}
		return fmt.Errorf("datastore creation task failed (UPID: %s): %w", upid, err)
	}

	// CRITICAL DEBUG: Log immediately when task completes
	fmt.Fprintf(os.Stderr, "[PBS-DEBUG] CreateDatastore: Task completed for '%s' at %s\n", datastore.Name, time.Now().Format(time.RFC3339Nano))

	if isDebugEnabled() {
		tflog.Debug(ctx, "CreateDatastore: Task completed successfully, sleeping 3s")
	}

	// Give PBS more time to register the datastore internally after task completion
	// CI VMs are slower than local machines and need more time for PBS to complete
	// internal registration after the async task finishes
	// The resource layer still has retry logic for additional eventual consistency handling
	time.Sleep(3 * time.Second)

	fmt.Fprintf(os.Stderr, "[PBS-DEBUG] CreateDatastore: Completed 3s sleep for '%s' at %s\n", datastore.Name, time.Now().Format(time.RFC3339Nano))

	if isDebugEnabled() {
		tflog.Debug(ctx, "CreateDatastore: Successfully created datastore", map[string]interface{}{
			"name": datastore.Name,
		})
	}

	// Datastore created successfully
	return nil
}

// UpdateDatastore updates an existing datastore configuration
func (c *Client) UpdateDatastore(ctx context.Context, name string, datastore *Datastore) error {
	if name == "" {
		return fmt.Errorf("datastore name is required")
	}

	// Convert struct to map for API request (excluding read-only fields for updates)
	body := c.datastoreToMapForUpdate(datastore)

	escapedName := url.PathEscape(name)
	_, err := c.api.Put(ctx, fmt.Sprintf("/config/datastore/%s", escapedName), body)
	if err != nil {
		return fmt.Errorf("failed to update datastore %s: %w", name, err)
	}

	return nil
}

// DeleteDatastore deletes a datastore configuration
func (c *Client) DeleteDatastore(ctx context.Context, name string) error {
	return c.DeleteDatastoreWithOptions(ctx, name, false)
}

// DeleteDatastoreWithOptions removes a datastore with optional parameters
// This is an asynchronous operation that returns a UPID task identifier.
// The function waits for the deletion task to complete before returning.
func (c *Client) DeleteDatastoreWithOptions(ctx context.Context, name string, destroyData bool) error {
	if name == "" {
		return fmt.Errorf("datastore name is required")
	}

	escapedName := url.PathEscape(name)
	path := fmt.Sprintf("/config/datastore/%s", escapedName)

	if destroyData {
		path += "?destroy-data=1"
	}

	resp, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete datastore %s: %w", name, err)
	}

	// Parse the UPID from the response
	var upid string
	if err := json.Unmarshal(resp.Data, &upid); err != nil {
		return fmt.Errorf("failed to parse UPID from delete response: %w", err)
	}

	// Get the node name from the UPID
	node, err := c.getNodeForTask(ctx, upid)
	if err != nil {
		return fmt.Errorf("failed to extract node from UPID: %w", err)
	}

	// Wait for the deletion task to complete (5 minute timeout like CreateDatastore)
	if err := c.api.WaitForTask(ctx, node, upid, 5*time.Minute); err != nil {
		return fmt.Errorf("datastore deletion task failed: %w", err)
	}

	// Sleep briefly to ensure eventual consistency (like we do for creation)
	time.Sleep(3 * time.Second)

	return nil
}

// datastoreToMap converts a Datastore struct to a map for API requests
func (c *Client) datastoreToMap(ds *Datastore) map[string]interface{} {
	body := map[string]interface{}{
		"name": ds.Name,
	}

	if ds.Path != "" {
		body["path"] = ds.Path
	}

	c.populateDatastoreMutableFields(body, ds)

	return body
}

// datastoreToMapForUpdate converts a Datastore struct to a map for update API requests
// Excludes fields that cannot be updated (like path, name, type)
func (c *Client) datastoreToMapForUpdate(ds *Datastore) map[string]interface{} {
	body := map[string]interface{}{}
	c.populateDatastoreMutableFields(body, ds)

	// Include digest for optimistic locking if present
	if ds.Digest != "" {
		body["digest"] = ds.Digest
	}

	if len(ds.Delete) > 0 {
		body["delete"] = ds.Delete
	}

	return body
}

func (c *Client) populateDatastoreMutableFields(body map[string]interface{}, ds *Datastore) {
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

	setString("comment", ds.Comment)
	setBool("disable", ds.Disabled)
	setString("gc-schedule", ds.GCSchedule)
	setString("prune-schedule", ds.PruneSchedule)

	setInt("keep-last", ds.KeepLast)
	setInt("keep-hourly", ds.KeepHourly)
	setInt("keep-daily", ds.KeepDaily)
	setInt("keep-weekly", ds.KeepWeekly)
	setInt("keep-monthly", ds.KeepMonthly)
	setInt("keep-yearly", ds.KeepYearly)

	if ds.MaintenanceMode != nil {
		body["maintenance-mode"] = formatMaintenanceMode(ds.MaintenanceMode)
	} else if ds.MaintenanceModeRaw != "" {
		body["maintenance-mode"] = ds.MaintenanceModeRaw
	}

	if ds.Notify != nil {
		body["notify"] = formatNotify(ds.Notify)
	} else if ds.NotifyRaw != "" {
		body["notify"] = ds.NotifyRaw
	}

	setString("notify-user", ds.NotifyUser)
	setString("notify-level", ds.NotifyLevel)
	setString("notification-mode", ds.NotificationMode)
	setBool("verify-new", ds.VerifyNew)
	setBool("reuse-datastore", ds.ReuseDatastore)
	setBool("overwrite-in-use", ds.OverwriteInUse)

	if ds.Tuning != nil {
		body["tuning"] = formatTuning(ds.Tuning)
	} else if ds.TuningRaw != "" {
		body["tuning"] = ds.TuningRaw
	}

	setString("fingerprint", ds.Fingerprint)
	setString("backing-device", ds.BackingDevice)

	backendIsS3 := strings.HasPrefix(ds.Backend, "type=s3") || (ds.S3Client != "" && ds.S3Bucket != "")
	if backendIsS3 {
		if ds.Backend != "" {
			body["backend"] = ds.Backend
		} else if ds.S3Client != "" && ds.S3Bucket != "" {
			body["backend"] = fmt.Sprintf("type=s3,client=%s,bucket=%s", ds.S3Client, ds.S3Bucket)
		}
	} else if ds.Backend != "" {
		body["backend"] = ds.Backend
	}
}

// getNodeForTask determines the node name for a given task
func (c *Client) getNodeForTask(ctx context.Context, upid string) (string, error) {
	// UPID format: "UPID:node:pid:starttime:type:id:user:status"
	// Extract node name from UPID
	parts := strings.Split(upid, ":")
	if len(parts) >= 2 && parts[0] == "UPID" {
		return parts[1], nil
	}

	// If UPID parsing fails, get the first available node
	nodes, err := c.api.GetNodes(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get nodes: %w", err)
	}

	if len(nodes) == 0 {
		return "", fmt.Errorf("no nodes available")
	}

	// Return the first available node
	return nodes[0].Node, nil
}

// parseBackendConfig parses backend configuration strings for supported backend types.
func (c *Client) parseBackendConfig(ds *Datastore) {
	if strings.TrimSpace(ds.Backend) == "" {
		return
	}

	backendType, params := ParseBackendString(ds.Backend)
	ds.BackendType = strings.ToLower(strings.TrimSpace(backendType))
	switch ds.BackendType {
	case "s3":
		if client, ok := params["client"]; ok {
			ds.S3Client = client
		}
		if bucket, ok := params["bucket"]; ok {
			ds.S3Bucket = bucket
		}
	case "removable":
		// No additional parameters to capture beyond backing-device field
	}
}
