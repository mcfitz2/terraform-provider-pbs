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
	"strings"
	"time"

	"github.com/micah/terraform-provider-pbs/pbs/api"
)

// Client represents the datastores API client
type Client struct {
	api *api.Client
}

// NewClient creates a new datastores API client
func NewClient(apiClient *api.Client) *Client {
	return &Client{api: apiClient}
}

// DatastoreType represents the type of datastore backend
type DatastoreType string

const (
	DatastoreTypeDirectory DatastoreType = "dir"
	DatastoreTypeZFS       DatastoreType = "zfs"
	DatastoreTypeLVM       DatastoreType = "lvm"
	DatastoreTypeCIFS      DatastoreType = "cifs"
	DatastoreTypeNFS       DatastoreType = "nfs"
	DatastoreTypeS3        DatastoreType = "s3"
)

// datastoreListItem represents a datastore in list responses (minimal info)
type datastoreListItem struct {
	Name            string `json:"name"`
	Path            string `json:"path"`
	MaintenanceMode string `json:"maintenance-mode,omitempty"`
}

// Datastore represents a PBS datastore configuration
type Datastore struct {
	Name          string        `json:"name"`
	Type          DatastoreType `json:"type"`
	Path          string        `json:"path,omitempty"`
	Content       []string      `json:"content,omitempty"`
	MaxBackups    *int          `json:"max-backups,omitempty"`
	Comment       string        `json:"comment,omitempty"`
	Disabled      *bool         `json:"disable,omitempty"`
	GCSchedule    string        `json:"gc-schedule,omitempty"`
	PruneSchedule string        `json:"prune-schedule,omitempty"`

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

	// ZFS-specific options
	ZFSPool     string `json:"pool,omitempty"`
	ZFSDataset  string `json:"dataset,omitempty"`
	BlockSize   string `json:"blocksize,omitempty"`
	Compression string `json:"compression,omitempty"`

	// LVM-specific options
	VolumeGroup string `json:"vg,omitempty"`
	ThinPool    string `json:"thinpool,omitempty"`

	// Network storage options (CIFS/NFS)
	Server   string `json:"server,omitempty"`
	Export   string `json:"export,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Share    string `json:"share,omitempty"`
	SubDir   string `json:"subdir,omitempty"`
	Options  string `json:"options,omitempty"`

	// S3 backend options (stored as backend configuration)
	Backend  string `json:"backend,omitempty"` // e.g. "type=s3,client=endpoint_id,bucket=bucket_name"
	S3Client string `json:"-"`                 // S3 endpoint ID (for easier access in Go code)
	S3Bucket string `json:"-"`

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
		// Try to infer type from path or use directory as default
		datastoreType := c.inferDatastoreType(item.Path)
		datastores[i] = Datastore{
			Name: item.Name,
			Path: item.Path,
			Type: datastoreType,
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
	resp, err := c.api.Get(ctx, fmt.Sprintf("/config/datastore/%s", escapedName))
	if err == nil {
		var ds Datastore
		if unmarshalErr := json.Unmarshal(resp.Data, &ds); unmarshalErr == nil {
			ds.Name = name // Ensure name is set

			// Parse S3 backend configuration if present
			if ds.Backend != "" {
				c.parseS3Backend(&ds)
			}

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

			// If type is not set, infer it
			if ds.Type == "" {
				ds.Type = c.inferDatastoreType(ds.Path)
			}
			return &ds, nil
		}
	}

	// If individual get fails, fall back to list and find
	datastores, err := c.ListDatastores(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list datastores to find %s: %w", name, err)
	}

	for _, ds := range datastores {
		if ds.Name == name {
			// Ensure type is set properly
			if ds.Type == "" {
				ds.Type = c.inferDatastoreType(ds.Path)
			}
			return &ds, nil
		}
	}

	return nil, fmt.Errorf("datastore %s not found", name)
}

// CreateDatastore creates a new datastore configuration
func (c *Client) CreateDatastore(ctx context.Context, datastore *Datastore) error {
	if datastore.Name == "" {
		return fmt.Errorf("datastore name is required")
	}

	// Convert struct to map for API request
	body := c.datastoreToMap(datastore)

	// Creating datastore with PBS API
	resp, err := c.api.Post(ctx, "/config/datastore", body)
	if err != nil {
		return fmt.Errorf("failed to create datastore %s: %w", datastore.Name, err)
	}

	// Parse the UPID from the response
	var upid string
	if err := json.Unmarshal(resp.Data, &upid); err != nil {
		return fmt.Errorf("failed to parse UPID from response: %w", err)
	}

	// Parse successful, proceeding with task monitoring

	// Get the node name from the UPID or by querying nodes
	node, err := c.getNodeForTask(ctx, upid)
	if err != nil {
		return fmt.Errorf("failed to determine node for task: %w", err)
	}

	// Wait for the task to complete with a reasonable timeout
	// For S3 datastores, this involves file I/O which can take time on slow connections
	// Wait for task completion
	if err := c.api.WaitForTask(ctx, node, upid, 5*time.Minute); err != nil {
		return fmt.Errorf("datastore creation task failed (UPID: %s): %w", upid, err)
	}

	// Give PBS a moment to register the datastore internally after task completion
	// This prevents race conditions where the datastore exists but isn't yet visible via API
	// Increased to 3 seconds to account for S3 network operations and PBS internal state updates
	time.Sleep(3 * time.Second)

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

// DeleteDatastoreWithOptions deletes a datastore configuration with additional options
func (c *Client) DeleteDatastoreWithOptions(ctx context.Context, name string, destroyData bool) error {
	if name == "" {
		return fmt.Errorf("datastore name is required")
	}

	escapedName := url.PathEscape(name)
	path := fmt.Sprintf("/config/datastore/%s", escapedName)

	// Add destroy-data parameter if requested
	if destroyData {
		path += "?destroy-data=1"
	}

	_, err := c.api.Delete(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to delete datastore %s: %w", name, err)
	}

	return nil
}

// datastoreToMap converts a Datastore struct to a map for API requests
func (c *Client) datastoreToMap(ds *Datastore) map[string]interface{} {
	body := map[string]interface{}{
		"name": ds.Name,
	}

	if ds.Type != "" {
		body["type"] = string(ds.Type)
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

	if len(ds.Content) > 0 {
		body["content"] = ds.Content
	}

	setInt("max-backups", ds.MaxBackups)
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

	setString("pool", ds.ZFSPool)
	setString("dataset", ds.ZFSDataset)
	setString("blocksize", ds.BlockSize)
	setString("compression", ds.Compression)

	setString("vg", ds.VolumeGroup)
	setString("thinpool", ds.ThinPool)

	setString("server", ds.Server)
	setString("export", ds.Export)
	setString("username", ds.Username)
	setString("password", ds.Password)
	setString("domain", ds.Domain)
	setString("share", ds.Share)
	setString("subdir", ds.SubDir)
	setString("options", ds.Options)

	if ds.Type == DatastoreTypeS3 || strings.HasPrefix(ds.Backend, "type=s3") {
		if ds.Backend != "" {
			body["backend"] = ds.Backend
		} else if ds.S3Client != "" && ds.S3Bucket != "" {
			body["backend"] = fmt.Sprintf("type=s3,client=%s,bucket=%s", ds.S3Client, ds.S3Bucket)
		}
	} else if ds.Backend != "" {
		body["backend"] = ds.Backend
	}
}

// inferDatastoreType attempts to determine datastore type from available information
func (c *Client) inferDatastoreType(path string) DatastoreType {
	// For now, assume all datastores are directory type since that's what we support
	// In the future, this could be enhanced to detect ZFS paths, LVM paths, etc.
	// based on path patterns or additional API calls
	// Note: S3 type should be determined by parseS3Backend when backend field is present
	return DatastoreTypeDirectory
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

// parseS3Backend parses the S3 backend configuration string and populates the datastore fields
// Backend format: "type=s3,client=endpoint_id,bucket=bucket_name"
func (c *Client) parseS3Backend(ds *Datastore) {
	if ds.Backend == "" {
		return
	}

	// Parse the backend configuration string
	parts := strings.Split(ds.Backend, ",")
	for _, part := range parts {
		keyValue := strings.SplitN(part, "=", 2)
		if len(keyValue) != 2 {
			continue
		}

		key := strings.TrimSpace(keyValue[0])
		value := strings.TrimSpace(keyValue[1])

		switch key {
		case "type":
			if value == "s3" {
				ds.Type = DatastoreTypeS3
			}
		case "client":
			ds.S3Client = value
		case "bucket":
			ds.S3Bucket = value
		}
	}
}
