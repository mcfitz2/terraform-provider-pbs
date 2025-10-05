/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package datastores provides Terraform resources for PBS datastores
package datastores

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/datastores"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &datastoreResource{}
	_ resource.ResourceWithConfigure   = &datastoreResource{}
	_ resource.ResourceWithImportState = &datastoreResource{}
)

// NewDatastoreResource is a helper function to simplify the provider implementation.
func NewDatastoreResource() resource.Resource {
	return &datastoreResource{}
}

// datastoreResource is the resource implementation.
type datastoreResource struct {
	client *pbs.Client
}

// datastoreResourceModel maps the resource schema data.
type datastoreResourceModel struct {
	Name           types.String `tfsdk:"name"`
	Type           types.String `tfsdk:"type"`
	Path           types.String `tfsdk:"path"`
	Content        types.List   `tfsdk:"content"`
	MaxBackups     types.Int64  `tfsdk:"max_backups"`
	Comment        types.String `tfsdk:"comment"`
	Disabled       types.Bool   `tfsdk:"disabled"`
	GCSchedule     types.String `tfsdk:"gc_schedule"`
	PruneSchedule  types.String `tfsdk:"prune_schedule"`
	CreateBasePath types.Bool   `tfsdk:"create_base_path"`

	// ZFS-specific options
	ZFSPool     types.String `tfsdk:"zfs_pool"`
	ZFSDataset  types.String `tfsdk:"zfs_dataset"`
	BlockSize   types.String `tfsdk:"block_size"`
	Compression types.String `tfsdk:"compression"`

	// LVM-specific options
	VolumeGroup types.String `tfsdk:"volume_group"`
	ThinPool    types.String `tfsdk:"thin_pool"`

	// Network storage options (CIFS/NFS)
	Server   types.String `tfsdk:"server"`
	Export   types.String `tfsdk:"export"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Domain   types.String `tfsdk:"domain"`
	Share    types.String `tfsdk:"share"`
	SubDir   types.String `tfsdk:"sub_dir"`
	Options  types.String `tfsdk:"options"`

	// Advanced options
	NotifyUser  types.String `tfsdk:"notify_user"`
	NotifyLevel types.String `tfsdk:"notify_level"`
	TuneLevel   types.Int64  `tfsdk:"tune_level"`
	Fingerprint types.String `tfsdk:"fingerprint"`

	// S3 backend options
	S3Client types.String `tfsdk:"s3_client"`
	S3Bucket types.String `tfsdk:"s3_bucket"`
}

// Metadata returns the resource type name.
func (r *datastoreResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastore"
}

// Schema defines the schema for the resource.
func (r *datastoreResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a PBS datastore configuration.",
		MarkdownDescription: "Manages a Proxmox Backup Server datastore configuration supporting directory, ZFS, LVM, CIFS, NFS, and S3 backends.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "Unique identifier for the datastore.",
				MarkdownDescription: "Unique identifier for the datastore.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 32),
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9\-]*$`),
						"Name must start with a letter and contain only letters, numbers, and hyphens.",
					),
				},
			},
			"type": schema.StringAttribute{
				Description:         "Type of datastore backend (dir, zfs, lvm, cifs, nfs, s3).",
				MarkdownDescription: "Type of datastore backend. Valid values: `dir`, `zfs`, `lvm`, `cifs`, `nfs`, `s3`.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("dir", "zfs", "lvm", "cifs", "nfs", "s3"),
				},
			},
			"path": schema.StringAttribute{
				Description:         "Path to the datastore (required for dir, optional for others).",
				MarkdownDescription: "Path to the datastore. Required for directory datastores, optional for others.",
				Optional:            true,
			},
			"content": schema.ListAttribute{
				Description:         "Content types allowed on this datastore.",
				MarkdownDescription: "Content types allowed on this datastore. Valid values: `backup`, `ct`, `iso`, `vztmpl`.",
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				Default:             listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("backup")})),
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.OneOf("backup", "ct", "iso", "vztmpl")),
				},
			},
			"max_backups": schema.Int64Attribute{
				Description:         "Maximum number of backups per guest (0 = unlimited).",
				MarkdownDescription: "Maximum number of backups per guest. Set to 0 for unlimited backups.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				Description:         "Description for the datastore.",
				MarkdownDescription: "Description for the datastore.",
				Optional:            true,
			},
			"disabled": schema.BoolAttribute{
				Description:         "Disable the datastore.",
				MarkdownDescription: "Whether the datastore is disabled.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"gc_schedule": schema.StringAttribute{
				Description:         "Garbage collection schedule in cron format.",
				MarkdownDescription: "Garbage collection schedule in cron format (e.g., `daily`, `weekly`, or `0 3 * * 0`).",
				Optional:            true,
			},
			"prune_schedule": schema.StringAttribute{
				Description:         "Prune schedule in cron format.",
				MarkdownDescription: "Prune schedule in cron format (e.g., `daily`, `weekly`, or `0 2 * * *`).",
				Optional:            true,
			},
			"create_base_path": schema.BoolAttribute{
				Description:         "Create base directory if it doesn't exist (dir type only).",
				MarkdownDescription: "Create base directory if it doesn't exist. Only applicable for directory datastores.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},

			// ZFS-specific attributes
			"zfs_pool": schema.StringAttribute{
				Description:         "ZFS pool name (zfs type only).",
				MarkdownDescription: "ZFS pool name. Required for ZFS datastores.",
				Optional:            true,
			},
			"zfs_dataset": schema.StringAttribute{
				Description:         "ZFS dataset name (zfs type only).",
				MarkdownDescription: "ZFS dataset name. Optional for ZFS datastores.",
				Optional:            true,
			},
			"block_size": schema.StringAttribute{
				Description:         "Block size for ZFS (zfs type only).",
				MarkdownDescription: "Block size for ZFS datasets (e.g., `4K`, `8K`, `16K`).",
				Optional:            true,
			},
			"compression": schema.StringAttribute{
				Description:         "Compression algorithm for ZFS (zfs type only).",
				MarkdownDescription: "Compression algorithm for ZFS. Valid values: `on`, `off`, `lz4`, `zstd`, `gzip`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("on", "off", "lz4", "zstd", "gzip"),
				},
			},

			// LVM-specific attributes
			"volume_group": schema.StringAttribute{
				Description:         "LVM volume group name (lvm type only).",
				MarkdownDescription: "LVM volume group name. Required for LVM datastores.",
				Optional:            true,
			},
			"thin_pool": schema.StringAttribute{
				Description:         "LVM thin pool name (lvm type only).",
				MarkdownDescription: "LVM thin pool name. Optional for LVM datastores.",
				Optional:            true,
			},

			// Network storage attributes
			"server": schema.StringAttribute{
				Description:         "Server hostname or IP address (cifs/nfs type only).",
				MarkdownDescription: "Server hostname or IP address. Required for CIFS/NFS datastores.",
				Optional:            true,
			},
			"export": schema.StringAttribute{
				Description:         "NFS export path (nfs type only).",
				MarkdownDescription: "NFS export path. Required for NFS datastores.",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				Description:         "Username for CIFS authentication (cifs type only).",
				MarkdownDescription: "Username for CIFS authentication. Optional for CIFS datastores.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				Description:         "Password for CIFS authentication (cifs type only).",
				MarkdownDescription: "Password for CIFS authentication. Optional for CIFS datastores.",
				Optional:            true,
				Sensitive:           true,
			},
			"domain": schema.StringAttribute{
				Description:         "Domain for CIFS authentication (cifs type only).",
				MarkdownDescription: "Domain for CIFS authentication. Optional for CIFS datastores.",
				Optional:            true,
			},
			"share": schema.StringAttribute{
				Description:         "CIFS share name (cifs type only).",
				MarkdownDescription: "CIFS share name. Required for CIFS datastores.",
				Optional:            true,
			},
			"sub_dir": schema.StringAttribute{
				Description:         "Subdirectory on the remote share.",
				MarkdownDescription: "Subdirectory on the remote share. Optional for network datastores.",
				Optional:            true,
			},
			"options": schema.StringAttribute{
				Description:         "Mount options for network storage.",
				MarkdownDescription: "Mount options for network storage (e.g., `vers=3,soft`).",
				Optional:            true,
			},

			// Advanced attributes
			"notify_user": schema.StringAttribute{
				Description:         "User to send notifications to.",
				MarkdownDescription: "User to send datastore notifications to (e.g., `root@pam`).",
				Optional:            true,
			},
			"notify_level": schema.StringAttribute{
				Description:         "Notification level.",
				MarkdownDescription: "Notification level. Valid values: `info`, `notice`, `warning`, `error`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("info", "notice", "warning", "error"),
				},
			},
			"tune_level": schema.Int64Attribute{
				Description:         "Tuning level for performance optimization.",
				MarkdownDescription: "Tuning level for performance optimization (0-4).",
				Optional:            true,
				Validators:          []validator.Int64{
					// Add range validator for 0-4
				},
			},
			"fingerprint": schema.StringAttribute{
				Description:         "Certificate fingerprint for secure connections.",
				MarkdownDescription: "Certificate fingerprint for secure connections (network datastores).",
				Optional:            true,
			},
			"s3_client": schema.StringAttribute{
				Description:         "S3 endpoint ID for S3 datastores.",
				MarkdownDescription: "S3 endpoint ID for S3 datastores. Must reference an existing S3 endpoint configuration.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"s3_bucket": schema.StringAttribute{
				Description:         "S3 bucket name for S3 datastores.",
				MarkdownDescription: "S3 bucket name for S3 datastores. The bucket must be created beforehand.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *datastoreResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	resourceConfig, ok := req.ProviderData.(*config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.Resource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = resourceConfig.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *datastoreResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan datastoreResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, "error reading plan in Create method", map[string]any{"diagnostics": resp.Diagnostics})
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Terraform Create method - plan: %+v", plan))

	// Validate type-specific requirements
	if err := r.validateDatastoreConfig(&plan); err != nil {
		resp.Diagnostics.AddError("Configuration Validation Error", err.Error())
		return
	}

	// Convert plan to datastore struct
	datastore, err := r.planToDatastore(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Create the datastore with retry logic for PBS lock contention
	err = r.createDatastoreWithRetry(ctx, datastore)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Datastore",
			"Could not create datastore, unexpected error: "+err.Error(),
		)
		return
	}

	// Log that the resource was created
	tflog.Trace(ctx, "created datastore resource", map[string]any{"name": plan.Name.ValueString()})

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *datastoreResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state datastoreResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed values from API
	// PBS datastore operations are asynchronous, so we may need to retry
	var datastore *datastores.Datastore
	var err error

	// Try up to 10 times with exponential backoff for async operations
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		datastore, err = r.client.Datastores.GetDatastore(ctx, state.Name.ValueString())
		if err == nil {
			break
		}

		// If it's the last attempt, don't wait
		if i < maxRetries-1 {
			// Wait with exponential backoff: 1s, 2s, 4s, 8s, 16s, 32s, but cap at 5s
			wait := time.Duration(1<<i) * time.Second
			if wait > 5*time.Second {
				wait = 5 * time.Second
			}
			time.Sleep(wait)
		}
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datastore",
			"Could not read datastore "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state with refreshed values
	err = r.datastoreToState(datastore, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Datastore",
			"Could not convert datastore to state: "+err.Error(),
		)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *datastoreResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan datastoreResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate type-specific requirements
	if err := r.validateDatastoreConfig(&plan); err != nil {
		resp.Diagnostics.AddError("Configuration Validation Error", err.Error())
		return
	}

	// Convert plan to datastore struct
	datastore, err := r.planToDatastore(&plan)
	if err != nil {
		resp.Diagnostics.AddError("Configuration Error", err.Error())
		return
	}

	// Update the datastore
	err = r.client.Datastores.UpdateDatastore(ctx, plan.Name.ValueString(), datastore)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Datastore",
			"Could not update datastore, unexpected error: "+err.Error(),
		)
		return
	}

	// Log that the resource was updated
	tflog.Trace(ctx, "updated datastore resource", map[string]any{"name": plan.Name.ValueString()})

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *datastoreResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state datastoreResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing datastore
	// Check if we should destroy data (useful for tests)
	destroyData := os.Getenv("PBS_DESTROY_DATA_ON_DELETE") == "true"
	err := r.client.Datastores.DeleteDatastoreWithOptions(ctx, state.Name.ValueString(), destroyData)
	if err != nil {
		// Check if the datastore is already gone (desired state achieved)
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "not found") ||
			strings.Contains(errorMsg, "does not exist") ||
			strings.Contains(errorMsg, "404") {
			// Resource already deleted - this is fine, desired state achieved
			tflog.Info(ctx, "Datastore already deleted", map[string]any{"name": state.Name.ValueString()})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting Datastore",
			"Could not delete datastore, unexpected error: "+err.Error(),
		)
		return
	}

	// Log that the resource was deleted
	tflog.Trace(ctx, "deleted datastore resource", map[string]any{"name": state.Name.ValueString()})
}

// ImportState imports an existing resource into Terraform state.
func (r *datastoreResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to name attribute
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

// Helper functions

// validateDatastoreConfig validates type-specific configuration requirements
func (r *datastoreResource) validateDatastoreConfig(plan *datastoreResourceModel) error {
	dsType := plan.Type.ValueString()

	switch dsType {
	case "dir":
		if plan.Path.IsNull() || plan.Path.ValueString() == "" {
			return fmt.Errorf("path is required for directory datastores")
		}
	case "zfs":
		if plan.ZFSPool.IsNull() || plan.ZFSPool.ValueString() == "" {
			return fmt.Errorf("zfs_pool is required for ZFS datastores")
		}
	case "lvm":
		if plan.VolumeGroup.IsNull() || plan.VolumeGroup.ValueString() == "" {
			return fmt.Errorf("volume_group is required for LVM datastores")
		}
	case "cifs":
		if plan.Server.IsNull() || plan.Server.ValueString() == "" {
			return fmt.Errorf("server is required for CIFS datastores")
		}
		if plan.Share.IsNull() || plan.Share.ValueString() == "" {
			return fmt.Errorf("share is required for CIFS datastores")
		}
	case "nfs":
		if plan.Server.IsNull() || plan.Server.ValueString() == "" {
			return fmt.Errorf("server is required for NFS datastores")
		}
		if plan.Export.IsNull() || plan.Export.ValueString() == "" {
			return fmt.Errorf("export is required for NFS datastores")
		}
	case "s3":
		if plan.Path.IsNull() || plan.Path.ValueString() == "" {
			return fmt.Errorf("path is required for S3 datastores (local cache directory)")
		}
		if plan.S3Client.IsNull() || plan.S3Client.ValueString() == "" {
			return fmt.Errorf("s3_client is required for S3 datastores")
		}
		if plan.S3Bucket.IsNull() || plan.S3Bucket.ValueString() == "" {
			return fmt.Errorf("s3_bucket is required for S3 datastores")
		}
	}

	return nil
}

// planToDatastore converts a Terraform plan to a datastore struct
func (r *datastoreResource) planToDatastore(plan *datastoreResourceModel) (*datastores.Datastore, error) {
	ds := &datastores.Datastore{
		Name: plan.Name.ValueString(),
		Type: datastores.DatastoreType(plan.Type.ValueString()),
	}

	// Common fields
	if !plan.Path.IsNull() {
		ds.Path = plan.Path.ValueString()
	}
	if !plan.Content.IsNull() {
		var content []string
		for _, item := range plan.Content.Elements() {
			content = append(content, item.(types.String).ValueString())
		}
		ds.Content = content
	}
	if !plan.MaxBackups.IsNull() {
		maxBackups := int(plan.MaxBackups.ValueInt64())
		ds.MaxBackups = &maxBackups
	}
	if !plan.Comment.IsNull() {
		ds.Comment = plan.Comment.ValueString()
	}
	if !plan.Disabled.IsNull() {
		disabled := plan.Disabled.ValueBool()
		ds.Disabled = &disabled
	}
	if !plan.GCSchedule.IsNull() {
		ds.GCSchedule = plan.GCSchedule.ValueString()
	}
	if !plan.PruneSchedule.IsNull() {
		ds.PruneSchedule = plan.PruneSchedule.ValueString()
	}

	// Directory-specific
	if !plan.CreateBasePath.IsNull() {
		createBasePath := plan.CreateBasePath.ValueBool()
		ds.CreateBasePath = &createBasePath
	}

	// ZFS-specific
	if !plan.ZFSPool.IsNull() {
		ds.ZFSPool = plan.ZFSPool.ValueString()
	}
	if !plan.ZFSDataset.IsNull() {
		ds.ZFSDataset = plan.ZFSDataset.ValueString()
	}
	if !plan.BlockSize.IsNull() {
		ds.BlockSize = plan.BlockSize.ValueString()
	}
	if !plan.Compression.IsNull() {
		ds.Compression = plan.Compression.ValueString()
	}

	// LVM-specific
	if !plan.VolumeGroup.IsNull() {
		ds.VolumeGroup = plan.VolumeGroup.ValueString()
	}
	if !plan.ThinPool.IsNull() {
		ds.ThinPool = plan.ThinPool.ValueString()
	}

	// Network storage
	if !plan.Server.IsNull() {
		ds.Server = plan.Server.ValueString()
	}
	if !plan.Export.IsNull() {
		ds.Export = plan.Export.ValueString()
	}
	if !plan.Username.IsNull() {
		ds.Username = plan.Username.ValueString()
	}
	if !plan.Password.IsNull() {
		ds.Password = plan.Password.ValueString()
	}
	if !plan.Domain.IsNull() {
		ds.Domain = plan.Domain.ValueString()
	}
	if !plan.Share.IsNull() {
		ds.Share = plan.Share.ValueString()
	}
	if !plan.SubDir.IsNull() {
		ds.SubDir = plan.SubDir.ValueString()
	}
	if !plan.Options.IsNull() {
		ds.Options = plan.Options.ValueString()
	}

	// Advanced options
	if !plan.NotifyUser.IsNull() {
		ds.NotifyUser = plan.NotifyUser.ValueString()
	}
	if !plan.NotifyLevel.IsNull() {
		ds.NotifyLevel = plan.NotifyLevel.ValueString()
	}
	if !plan.TuneLevel.IsNull() {
		tuneLevel := int(plan.TuneLevel.ValueInt64())
		ds.TuneLevel = &tuneLevel
	}
	if !plan.Fingerprint.IsNull() {
		ds.Fingerprint = plan.Fingerprint.ValueString()
	}

	// S3-specific options
	if !plan.S3Client.IsNull() {
		ds.S3Client = plan.S3Client.ValueString()
	}
	if !plan.S3Bucket.IsNull() {
		ds.S3Bucket = plan.S3Bucket.ValueString()
	}

	return ds, nil
}

// datastoreToState converts a datastore struct to Terraform state
func (r *datastoreResource) datastoreToState(ds *datastores.Datastore, state *datastoreResourceModel) error {
	state.Name = types.StringValue(ds.Name)
	state.Type = types.StringValue(string(ds.Type))

	// Common fields
	if ds.Path != "" {
		state.Path = types.StringValue(ds.Path)
	}
	if len(ds.Content) > 0 {
		contentValues := make([]attr.Value, len(ds.Content))
		for i, content := range ds.Content {
			contentValues[i] = types.StringValue(content)
		}
		state.Content = types.ListValueMust(types.StringType, contentValues)
	}
	if ds.MaxBackups != nil {
		state.MaxBackups = types.Int64Value(int64(*ds.MaxBackups))
	}
	if ds.Comment != "" {
		state.Comment = types.StringValue(ds.Comment)
	}
	if ds.Disabled != nil {
		state.Disabled = types.BoolValue(*ds.Disabled)
	}
	if ds.GCSchedule != "" {
		state.GCSchedule = types.StringValue(ds.GCSchedule)
	}
	if ds.PruneSchedule != "" {
		state.PruneSchedule = types.StringValue(ds.PruneSchedule)
	}

	// Directory-specific
	if ds.CreateBasePath != nil {
		state.CreateBasePath = types.BoolValue(*ds.CreateBasePath)
	}

	// ZFS-specific
	if ds.ZFSPool != "" {
		state.ZFSPool = types.StringValue(ds.ZFSPool)
	}
	if ds.ZFSDataset != "" {
		state.ZFSDataset = types.StringValue(ds.ZFSDataset)
	}
	if ds.BlockSize != "" {
		state.BlockSize = types.StringValue(ds.BlockSize)
	}
	if ds.Compression != "" {
		state.Compression = types.StringValue(ds.Compression)
	}

	// LVM-specific
	if ds.VolumeGroup != "" {
		state.VolumeGroup = types.StringValue(ds.VolumeGroup)
	}
	if ds.ThinPool != "" {
		state.ThinPool = types.StringValue(ds.ThinPool)
	}

	// Network storage
	if ds.Server != "" {
		state.Server = types.StringValue(ds.Server)
	}
	if ds.Export != "" {
		state.Export = types.StringValue(ds.Export)
	}
	if ds.Username != "" {
		state.Username = types.StringValue(ds.Username)
	}
	// Note: Password is sensitive and not returned by the API
	if ds.Domain != "" {
		state.Domain = types.StringValue(ds.Domain)
	}
	if ds.Share != "" {
		state.Share = types.StringValue(ds.Share)
	}
	if ds.SubDir != "" {
		state.SubDir = types.StringValue(ds.SubDir)
	}
	if ds.Options != "" {
		state.Options = types.StringValue(ds.Options)
	}

	// Advanced options
	if ds.NotifyUser != "" {
		state.NotifyUser = types.StringValue(ds.NotifyUser)
	}
	if ds.NotifyLevel != "" {
		state.NotifyLevel = types.StringValue(ds.NotifyLevel)
	}
	if ds.TuneLevel != nil {
		state.TuneLevel = types.Int64Value(int64(*ds.TuneLevel))
	}
	if ds.Fingerprint != "" {
		state.Fingerprint = types.StringValue(ds.Fingerprint)
	}

	// S3-specific
	if ds.S3Client != "" {
		state.S3Client = types.StringValue(ds.S3Client)
	}
	if ds.S3Bucket != "" {
		state.S3Bucket = types.StringValue(ds.S3Bucket)
	}

	return nil
}

// createDatastoreWithRetry attempts to create a datastore with retry logic for PBS lock contention
func (r *datastoreResource) createDatastoreWithRetry(ctx context.Context, datastore *datastores.Datastore) error {
	maxRetries := 3
	baseDelay := 2 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := r.client.Datastores.CreateDatastore(ctx, datastore)
		if err == nil {
			return nil
		}

		// Check if this is a lock contention error or task failure
		errorMsg := err.Error()

		// Log detailed information for task failures
		if strings.Contains(errorMsg, "task failed") {
			// Try to extract UPID from error message
			// PBS task errors often contain format like "UPID:node:00001234:..."
			upid := "unknown"
			if strings.Contains(errorMsg, "UPID:") {
				// Extract UPID from error message
				parts := strings.Split(errorMsg, "UPID:")
				if len(parts) > 1 {
					upidPart := strings.Split(parts[1], " ")[0]
					if upidPart != "" {
						upid = "UPID:" + upidPart
					}
				}
			}

			// Check for known compatibility issues
			isBackblazeCompatIssue := strings.Contains(errorMsg, "501") &&
				strings.Contains(errorMsg, "Not Implemented") &&
				strings.Contains(errorMsg, "access time safety check")

			logLevel := "Error"
			if isBackblazeCompatIssue {
				logLevel = "Warn" // Known issue, not provider error
			}

			tflog.Error(ctx, fmt.Sprintf("PBS task failed (%s)", logLevel), map[string]any{
				"error":                     errorMsg,
				"upid":                      upid,
				"attempt":                   attempt,
				"datastore":                 datastore.Name,
				"known_compatibility_issue": isBackblazeCompatIssue,
			})

			// For known Backblaze compatibility issues, don't retry
			if isBackblazeCompatIssue {
				return fmt.Errorf("known compatibility issue: %s", errorMsg)
			}
		}

		isLockError := strings.Contains(errorMsg, "Unable to acquire lock") ||
			strings.Contains(errorMsg, "Interrupted system call") ||
			strings.Contains(errorMsg, ".datastore.lck")

		if isLockError && attempt < maxRetries {
			// Exponential backoff with jitter
			delay := baseDelay * time.Duration(attempt)
			time.Sleep(delay)
			continue
		}

		// Not a lock error or final attempt - return error
		return err
	}

	return fmt.Errorf("failed to create datastore after %d attempts", maxRetries)
}
