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

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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
	Name          types.String `tfsdk:"name"`
	Path          types.String `tfsdk:"path"`
	Comment       types.String `tfsdk:"comment"`
	Disabled      types.Bool   `tfsdk:"disabled"`
	GCSchedule    types.String `tfsdk:"gc_schedule"`
	PruneSchedule types.String `tfsdk:"prune_schedule"`
	KeepLast      types.Int64  `tfsdk:"keep_last"`
	KeepHourly    types.Int64  `tfsdk:"keep_hourly"`
	KeepDaily     types.Int64  `tfsdk:"keep_daily"`
	KeepWeekly    types.Int64  `tfsdk:"keep_weekly"`
	KeepMonthly   types.Int64  `tfsdk:"keep_monthly"`
	KeepYearly    types.Int64  `tfsdk:"keep_yearly"`

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
	NotifyUser       types.String          `tfsdk:"notify_user"`
	NotifyLevel      types.String          `tfsdk:"notify_level"`
	NotificationMode types.String          `tfsdk:"notification_mode"`
	Notify           *notifyModel          `tfsdk:"notify"`
	MaintenanceMode  *maintenanceModeModel `tfsdk:"maintenance_mode"`
	VerifyNew        types.Bool            `tfsdk:"verify_new"`
	ReuseDatastore   types.Bool            `tfsdk:"reuse_datastore"`
	OverwriteInUse   types.Bool            `tfsdk:"overwrite_in_use"`
	Tuning           *tuningModel          `tfsdk:"tuning"`
	TuneLevel        types.Int64           `tfsdk:"tune_level"`
	Fingerprint      types.String          `tfsdk:"fingerprint"`
	Digest           types.String          `tfsdk:"digest"`

	// S3 backend options
	S3Client types.String `tfsdk:"s3_client"`
	S3Bucket types.String `tfsdk:"s3_bucket"`
}

type maintenanceModeModel struct {
	Type    types.String `tfsdk:"type"`
	Message types.String `tfsdk:"message"`
}

type notifyModel struct {
	GC     types.String `tfsdk:"gc"`
	Prune  types.String `tfsdk:"prune"`
	Sync   types.String `tfsdk:"sync"`
	Verify types.String `tfsdk:"verify"`
}

type tuningModel struct {
	ChunkOrder         types.String `tfsdk:"chunk_order"`
	GCAtimeCutoff      types.Int64  `tfsdk:"gc_atime_cutoff"`
	GCAtimeSafetyCheck types.Bool   `tfsdk:"gc_atime_safety_check"`
	GCCacheCapacity    types.Int64  `tfsdk:"gc_cache_capacity"`
	SyncLevel          types.String `tfsdk:"sync_level"`
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
			"path": schema.StringAttribute{
				Description:         "Path to the datastore (required for dir, optional for others).",
				MarkdownDescription: "Path to the datastore. Required for directory datastores, optional for others.",
				Optional:            true,
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
				DeprecationMessage:  "Removed in PBS 4.0+. Configure prune jobs with the pbs_prune_job resource instead.",
			},
			"keep_last": schema.Int64Attribute{
				Description:         "Number of latest backups to keep.",
				MarkdownDescription: "Number of latest backups to keep when pruning.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"keep_hourly": schema.Int64Attribute{
				Description:         "Number of hourly backups to keep.",
				MarkdownDescription: "Number of hourly backups to keep when pruning.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"keep_daily": schema.Int64Attribute{
				Description:         "Number of daily backups to keep.",
				MarkdownDescription: "Number of daily backups to keep when pruning.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"keep_weekly": schema.Int64Attribute{
				Description:         "Number of weekly backups to keep.",
				MarkdownDescription: "Number of weekly backups to keep when pruning.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"keep_monthly": schema.Int64Attribute{
				Description:         "Number of monthly backups to keep.",
				MarkdownDescription: "Number of monthly backups to keep when pruning.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"keep_yearly": schema.Int64Attribute{
				Description:         "Number of yearly backups to keep.",
				MarkdownDescription: "Number of yearly backups to keep when pruning.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
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
			"notification_mode": schema.StringAttribute{
				Description:         "Notification delivery mode.",
				MarkdownDescription: "Notification delivery mode. Valid values: `legacy-sendmail`, `notification-system`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("legacy-sendmail", "notification-system"),
				},
			},
			"notify": schema.SingleNestedAttribute{
				Description:         "Per-job notification settings overriding datastore defaults.",
				MarkdownDescription: "Per-job notification settings overriding datastore defaults.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"gc": schema.StringAttribute{
						Description:         "Garbage collection notification level.",
						MarkdownDescription: "Garbage collection notification level. Valid values: `always`, `error`, `never`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("always", "error", "never"),
						},
					},
					"prune": schema.StringAttribute{
						Description:         "Prune job notification level.",
						MarkdownDescription: "Prune job notification level. Valid values: `always`, `error`, `never`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("always", "error", "never"),
						},
					},
					"sync": schema.StringAttribute{
						Description:         "Sync job notification level.",
						MarkdownDescription: "Sync job notification level. Valid values: `always`, `error`, `never`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("always", "error", "never"),
						},
					},
					"verify": schema.StringAttribute{
						Description:         "Verification job notification level.",
						MarkdownDescription: "Verification job notification level. Valid values: `always`, `error`, `never`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("always", "error", "never"),
						},
					},
				},
			},
			"maintenance_mode": schema.SingleNestedAttribute{
				Description:         "Maintenance mode configuration.",
				MarkdownDescription: "Maintenance mode configuration allowing `offline` or `read-only` modes with optional message.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description:         "Maintenance mode type.",
						MarkdownDescription: "Maintenance mode type. Valid values: `offline`, `read-only`.",
						Required:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("offline", "read-only"),
						},
					},
					"message": schema.StringAttribute{
						Description:         "Message shown in maintenance mode.",
						MarkdownDescription: "Optional message presented for maintenance mode.",
						Optional:            true,
					},
				},
			},
			"verify_new": schema.BoolAttribute{
				Description:         "Verify newly created snapshots immediately after backup.",
				MarkdownDescription: "Verify newly created snapshots immediately after backup.",
				Optional:            true,
			},
			"reuse_datastore": schema.BoolAttribute{
				Description:         "Reuse existing datastore chunks when possible.",
				MarkdownDescription: "Reuse existing datastore chunks when possible.",
				Optional:            true,
			},
			"overwrite_in_use": schema.BoolAttribute{
				Description:         "Allow overwriting chunks that are currently in use.",
				MarkdownDescription: "Allow overwriting chunks that are currently in use.",
				Optional:            true,
			},
			"tuning": schema.SingleNestedAttribute{
				Description:         "Advanced tuning options for datastore behaviour.",
				MarkdownDescription: "Advanced tuning options for datastore behaviour such as chunk order and sync level.",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"chunk_order": schema.StringAttribute{
						Description:         "Chunk iteration order.",
						MarkdownDescription: "Chunk iteration order. Valid values: `inode`, `none`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("inode", "none"),
						},
					},
					"gc_atime_cutoff": schema.Int64Attribute{
						Description:         "Garbage collection access time cutoff (seconds).",
						MarkdownDescription: "Garbage collection access time cutoff in seconds.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"gc_atime_safety_check": schema.BoolAttribute{
						Description:         "Enable garbage collection access time safety check.",
						MarkdownDescription: "Enable garbage collection access time safety check.",
						Optional:            true,
					},
					"gc_cache_capacity": schema.Int64Attribute{
						Description:         "Garbage collection cache capacity.",
						MarkdownDescription: "Garbage collection cache capacity.",
						Optional:            true,
						Validators: []validator.Int64{
							int64validator.AtLeast(0),
						},
					},
					"sync_level": schema.StringAttribute{
						Description:         "Datastore fsync level.",
						MarkdownDescription: "Datastore fsync level. Valid values: `none`, `filesystem`, `file`.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("none", "filesystem", "file"),
						},
					},
				},
			},
			"tune_level": schema.Int64Attribute{
				Description:         "Tuning level for performance optimization.",
				MarkdownDescription: "Tuning level for performance optimization (0-4).",
				Optional:            true,
				DeprecationMessage:  "Use tuning.sync_level instead.",
				Validators: []validator.Int64{
					int64validator.Between(0, 4),
				},
			},
			"fingerprint": schema.StringAttribute{
				Description:         "Certificate fingerprint for secure connections.",
				MarkdownDescription: "Certificate fingerprint for secure connections (network datastores).",
				Optional:            true,
			},
			"digest": schema.StringAttribute{
				Description:         "Opaque digest returned by PBS for optimistic locking.",
				MarkdownDescription: "Opaque digest returned by PBS for optimistic locking.",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
	datastore, err := r.planToDatastore(&plan, nil)
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

	createdDatastore, err := r.client.Datastores.GetDatastore(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datastore",
			"Could not read datastore "+plan.Name.ValueString()+" after creation: "+err.Error(),
		)
		return
	}

	state := plan

	if err := r.datastoreToState(createdDatastore, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Datastore",
			"Could not convert datastore to state after creation: "+err.Error(),
		)
		return
	}

	if createdDatastore.Disabled == nil && state.Disabled.IsNull() {
		state.Disabled = types.BoolValue(false)
	}

	// Preserve sensitive or plan-only fields not returned by the API
	state.Password = plan.Password
	state.PruneSchedule = plan.PruneSchedule
	state.Path = plan.Path
	state.Comment = plan.Comment
	state.GCSchedule = plan.GCSchedule
	state.KeepLast = plan.KeepLast
	state.KeepHourly = plan.KeepHourly
	state.KeepDaily = plan.KeepDaily
	state.KeepWeekly = plan.KeepWeekly
	state.KeepMonthly = plan.KeepMonthly
	state.KeepYearly = plan.KeepYearly
	state.NotifyUser = plan.NotifyUser
	state.NotifyLevel = plan.NotifyLevel
	state.NotificationMode = plan.NotificationMode
	state.Notify = plan.Notify
	state.MaintenanceMode = plan.MaintenanceMode
	state.VerifyNew = plan.VerifyNew
	state.ReuseDatastore = plan.ReuseDatastore
	state.OverwriteInUse = plan.OverwriteInUse
	state.Tuning = plan.Tuning
	state.TuneLevel = plan.TuneLevel
	state.Fingerprint = plan.Fingerprint
	state.S3Client = plan.S3Client
	state.S3Bucket = plan.S3Bucket

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

	var state datastoreResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if plan.Digest.IsNull() && !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		plan.Digest = state.Digest
	}

	// Validate type-specific requirements
	if err := r.validateDatastoreConfig(&plan); err != nil {
		resp.Diagnostics.AddError("Configuration Validation Error", err.Error())
		return
	}

	// Convert plan to datastore struct
	datastore, err := r.planToDatastore(&plan, &state)
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

// validateDatastoreConfig validates configuration requirements that span backend types
func (r *datastoreResource) validateDatastoreConfig(plan *datastoreResourceModel) error {
	if !plan.PruneSchedule.IsNull() && !plan.PruneSchedule.IsUnknown() && strings.TrimSpace(plan.PruneSchedule.ValueString()) != "" {
		return fmt.Errorf("prune_schedule was removed in PBS 4.0. Configure pruning with the pbs_prune_job resource instead")
	}

	s3ClientSet := !plan.S3Client.IsNull() && !plan.S3Client.IsUnknown() && strings.TrimSpace(plan.S3Client.ValueString()) != ""
	s3BucketSet := !plan.S3Bucket.IsNull() && !plan.S3Bucket.IsUnknown() && strings.TrimSpace(plan.S3Bucket.ValueString()) != ""

	if s3ClientSet != s3BucketSet {
		return fmt.Errorf("s3_client and s3_bucket must be provided together")
	}

	if s3ClientSet {
		if plan.Path.IsNull() || plan.Path.IsUnknown() || strings.TrimSpace(plan.Path.ValueString()) == "" {
			return fmt.Errorf("path is required for S3 datastores (local cache directory)")
		}
		return nil
	}

	isZFS := !plan.ZFSPool.IsNull() && !plan.ZFSPool.IsUnknown()
	isLVM := !plan.VolumeGroup.IsNull() && !plan.VolumeGroup.IsUnknown()
	isCIFS := !plan.Share.IsNull() && !plan.Share.IsUnknown()
	isNFS := !plan.Export.IsNull() && !plan.Export.IsUnknown()

	if isZFS {
		if strings.TrimSpace(plan.ZFSPool.ValueString()) == "" {
			return fmt.Errorf("zfs_pool is required when configuring ZFS-backed datastores")
		}
		return nil
	}

	if isLVM {
		if strings.TrimSpace(plan.VolumeGroup.ValueString()) == "" {
			return fmt.Errorf("volume_group is required when configuring LVM-backed datastores")
		}
		return nil
	}

	if isCIFS {
		if plan.Server.IsNull() || plan.Server.IsUnknown() || strings.TrimSpace(plan.Server.ValueString()) == "" {
			return fmt.Errorf("server is required when configuring CIFS-backed datastores")
		}
		if strings.TrimSpace(plan.Share.ValueString()) == "" {
			return fmt.Errorf("share is required when configuring CIFS-backed datastores")
		}
		return nil
	}

	if isNFS {
		if plan.Server.IsNull() || plan.Server.IsUnknown() || strings.TrimSpace(plan.Server.ValueString()) == "" {
			return fmt.Errorf("server is required when configuring NFS-backed datastores")
		}
		if strings.TrimSpace(plan.Export.ValueString()) == "" {
			return fmt.Errorf("export is required when configuring NFS-backed datastores")
		}
		return nil
	}

	if plan.Path.IsNull() || plan.Path.IsUnknown() || strings.TrimSpace(plan.Path.ValueString()) == "" {
		return fmt.Errorf("path is required for directory-backed datastores")
	}

	return nil
}

// planToDatastore converts a Terraform plan to a datastore struct, applying state-aware deletions when provided.
func (r *datastoreResource) planToDatastore(plan *datastoreResourceModel, state *datastoreResourceModel) (*datastores.Datastore, error) {
	ds := &datastores.Datastore{
		Name: plan.Name.ValueString(),
	}

	// Common fields
	if !plan.Path.IsNull() && !plan.Path.IsUnknown() {
		ds.Path = plan.Path.ValueString()
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		ds.Comment = plan.Comment.ValueString()
	}
	if ptr := optionalBoolPointer(plan.Disabled); ptr != nil {
		ds.Disabled = ptr
		if ds.Disabled != nil && !*ds.Disabled {
			ds.Disabled = nil
		}
	}
	if !plan.GCSchedule.IsNull() && !plan.GCSchedule.IsUnknown() {
		ds.GCSchedule = plan.GCSchedule.ValueString()
	}
	if ptr := optionalInt64Pointer(plan.KeepLast); ptr != nil {
		ds.KeepLast = ptr
	}
	if ptr := optionalInt64Pointer(plan.KeepHourly); ptr != nil {
		ds.KeepHourly = ptr
	}
	if ptr := optionalInt64Pointer(plan.KeepDaily); ptr != nil {
		ds.KeepDaily = ptr
	}
	if ptr := optionalInt64Pointer(plan.KeepWeekly); ptr != nil {
		ds.KeepWeekly = ptr
	}
	if ptr := optionalInt64Pointer(plan.KeepMonthly); ptr != nil {
		ds.KeepMonthly = ptr
	}
	if ptr := optionalInt64Pointer(plan.KeepYearly); ptr != nil {
		ds.KeepYearly = ptr
	}

	// ZFS-specific
	if !plan.ZFSPool.IsNull() && !plan.ZFSPool.IsUnknown() {
		ds.ZFSPool = plan.ZFSPool.ValueString()
	}
	if !plan.ZFSDataset.IsNull() && !plan.ZFSDataset.IsUnknown() {
		ds.ZFSDataset = plan.ZFSDataset.ValueString()
	}
	if !plan.BlockSize.IsNull() && !plan.BlockSize.IsUnknown() {
		ds.BlockSize = plan.BlockSize.ValueString()
	}
	if !plan.Compression.IsNull() && !plan.Compression.IsUnknown() {
		ds.Compression = plan.Compression.ValueString()
	}

	// LVM-specific
	if !plan.VolumeGroup.IsNull() && !plan.VolumeGroup.IsUnknown() {
		ds.VolumeGroup = plan.VolumeGroup.ValueString()
	}
	if !plan.ThinPool.IsNull() && !plan.ThinPool.IsUnknown() {
		ds.ThinPool = plan.ThinPool.ValueString()
	}

	// Network storage
	if !plan.Server.IsNull() && !plan.Server.IsUnknown() {
		ds.Server = plan.Server.ValueString()
	}
	if !plan.Export.IsNull() && !plan.Export.IsUnknown() {
		ds.Export = plan.Export.ValueString()
	}
	if !plan.Username.IsNull() && !plan.Username.IsUnknown() {
		ds.Username = plan.Username.ValueString()
	}
	if !plan.Password.IsNull() && !plan.Password.IsUnknown() {
		ds.Password = plan.Password.ValueString()
	}
	if !plan.Domain.IsNull() && !plan.Domain.IsUnknown() {
		ds.Domain = plan.Domain.ValueString()
	}
	if !plan.Share.IsNull() && !plan.Share.IsUnknown() {
		ds.Share = plan.Share.ValueString()
	}
	if !plan.SubDir.IsNull() && !plan.SubDir.IsUnknown() {
		ds.SubDir = plan.SubDir.ValueString()
	}
	if !plan.Options.IsNull() && !plan.Options.IsUnknown() {
		ds.Options = plan.Options.ValueString()
	}

	// Advanced options & toggles
	if !plan.NotifyUser.IsNull() && !plan.NotifyUser.IsUnknown() {
		ds.NotifyUser = plan.NotifyUser.ValueString()
	}
	if !plan.NotifyLevel.IsNull() && !plan.NotifyLevel.IsUnknown() {
		ds.NotifyLevel = plan.NotifyLevel.ValueString()
	}
	if !plan.NotificationMode.IsNull() && !plan.NotificationMode.IsUnknown() {
		ds.NotificationMode = plan.NotificationMode.ValueString()
	}
	if ptr := optionalBoolPointer(plan.VerifyNew); ptr != nil {
		ds.VerifyNew = ptr
	}
	if ptr := optionalBoolPointer(plan.ReuseDatastore); ptr != nil {
		ds.ReuseDatastore = ptr
	}
	if ptr := optionalBoolPointer(plan.OverwriteInUse); ptr != nil {
		ds.OverwriteInUse = ptr
	}

	notifyBlockDefined := plan.Notify != nil
	notifyHasValues := false
	if plan.Notify != nil {
		notify := &datastores.DatastoreNotify{}
		if !plan.Notify.GC.IsNull() && !plan.Notify.GC.IsUnknown() {
			notify.GC = strings.ToLower(plan.Notify.GC.ValueString())
			notifyHasValues = notifyHasValues || notify.GC != ""
		}
		if !plan.Notify.Prune.IsNull() && !plan.Notify.Prune.IsUnknown() {
			notify.Prune = strings.ToLower(plan.Notify.Prune.ValueString())
			notifyHasValues = notifyHasValues || notify.Prune != ""
		}
		if !plan.Notify.Sync.IsNull() && !plan.Notify.Sync.IsUnknown() {
			notify.Sync = strings.ToLower(plan.Notify.Sync.ValueString())
			notifyHasValues = notifyHasValues || notify.Sync != ""
		}
		if !plan.Notify.Verify.IsNull() && !plan.Notify.Verify.IsUnknown() {
			notify.Verify = strings.ToLower(plan.Notify.Verify.ValueString())
			notifyHasValues = notifyHasValues || notify.Verify != ""
		}
		if notifyHasValues {
			ds.Notify = notify
		}
	}

	if plan.MaintenanceMode != nil {
		mmType := ""
		if !plan.MaintenanceMode.Type.IsNull() && !plan.MaintenanceMode.Type.IsUnknown() {
			mmType = strings.ToLower(plan.MaintenanceMode.Type.ValueString())
		}
		mm := &datastores.MaintenanceMode{Type: mmType}
		if !plan.MaintenanceMode.Message.IsNull() && !plan.MaintenanceMode.Message.IsUnknown() {
			mm.Message = plan.MaintenanceMode.Message.ValueString()
		}
		if mm.Type != "" || mm.Message != "" {
			ds.MaintenanceMode = mm
		}
	}

	tuningBlockDefined := plan.Tuning != nil
	tuningHasValues := false
	if plan.Tuning != nil {
		tuning := &datastores.DatastoreTuning{}
		if !plan.Tuning.ChunkOrder.IsNull() && !plan.Tuning.ChunkOrder.IsUnknown() {
			tuning.ChunkOrder = strings.ToLower(plan.Tuning.ChunkOrder.ValueString())
		}
		if ptr := optionalInt64Pointer(plan.Tuning.GCAtimeCutoff); ptr != nil {
			tuning.GCAtimeCutoff = ptr
		}
		if ptr := optionalBoolPointer(plan.Tuning.GCAtimeSafetyCheck); ptr != nil {
			tuning.GCAtimeSafetyCheck = ptr
		}
		if ptr := optionalInt64Pointer(plan.Tuning.GCCacheCapacity); ptr != nil {
			tuning.GCCacheCapacity = ptr
		}
		if !plan.Tuning.SyncLevel.IsNull() && !plan.Tuning.SyncLevel.IsUnknown() {
			tuning.SyncLevel = strings.ToLower(plan.Tuning.SyncLevel.ValueString())
		}
		if !isEmptyTuning(tuning) {
			ds.Tuning = tuning
			tuningHasValues = true
		}
	}

	if !plan.TuneLevel.IsNull() && !plan.TuneLevel.IsUnknown() {
		syncLevel, err := tuneLevelToSyncLevel(int(plan.TuneLevel.ValueInt64()))
		if err != nil {
			return nil, err
		}
		if ds.Tuning == nil {
			ds.Tuning = &datastores.DatastoreTuning{}
		}
		ds.Tuning.SyncLevel = syncLevel
		tuningHasValues = true
	}

	if !plan.Fingerprint.IsNull() && !plan.Fingerprint.IsUnknown() {
		ds.Fingerprint = plan.Fingerprint.ValueString()
	}

	if !plan.S3Client.IsNull() && !plan.S3Client.IsUnknown() {
		ds.S3Client = plan.S3Client.ValueString()
	}
	if !plan.S3Bucket.IsNull() && !plan.S3Bucket.IsUnknown() {
		ds.S3Bucket = plan.S3Bucket.ValueString()
	}
	if ds.S3Client != "" && ds.S3Bucket != "" {
		ds.Backend = fmt.Sprintf("type=s3,client=%s,bucket=%s", ds.S3Client, ds.S3Bucket)
	}

	if !plan.Digest.IsNull() && !plan.Digest.IsUnknown() {
		ds.Digest = plan.Digest.ValueString()
	}

	if state != nil {
		var deletes []string

		if shouldDeleteStringAttr(plan.NotifyUser, state.NotifyUser) {
			deletes = append(deletes, "notify-user")
		}
		if shouldDeleteStringAttr(plan.NotifyLevel, state.NotifyLevel) {
			deletes = append(deletes, "notify-level")
		}
		if shouldDeleteStringAttr(plan.NotificationMode, state.NotificationMode) {
			deletes = append(deletes, "notification-mode")
		}
		if (plan.Notify == nil && state.Notify != nil) || (notifyBlockDefined && !notifyHasValues && state.Notify != nil) {
			deletes = append(deletes, "notify")
		}
		if plan.MaintenanceMode == nil && state.MaintenanceMode != nil {
			deletes = append(deletes, "maintenance-mode")
		}
		if ((plan.Tuning == nil && plan.TuneLevel.IsNull()) || (tuningBlockDefined && !tuningHasValues && plan.TuneLevel.IsNull())) && state.Tuning != nil {
			deletes = append(deletes, "tuning")
		}
		if plan.VerifyNew.IsNull() && hasBoolValue(state.VerifyNew) {
			deletes = append(deletes, "verify-new")
		}
		if plan.ReuseDatastore.IsNull() && hasBoolValue(state.ReuseDatastore) {
			deletes = append(deletes, "reuse-datastore")
		}
		if plan.OverwriteInUse.IsNull() && hasBoolValue(state.OverwriteInUse) {
			deletes = append(deletes, "overwrite-in-use")
		}
		if len(deletes) > 0 {
			ds.Delete = deletes
		}
	}

	return ds, nil
}

func optionalInt64Pointer(value types.Int64) *int {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := int(value.ValueInt64())
	return &v
}

func optionalBoolPointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

func hasBoolValue(value types.Bool) bool {
	return !value.IsNull() && !value.IsUnknown()
}

func isEmptyTuning(t *datastores.DatastoreTuning) bool {
	if t == nil {
		return true
	}
	return t.ChunkOrder == "" && t.GCAtimeCutoff == nil && t.GCAtimeSafetyCheck == nil && t.GCCacheCapacity == nil && t.SyncLevel == ""
}

func shouldDeleteStringAttr(plan types.String, state types.String) bool {
	if state.IsNull() || state.IsUnknown() {
		return false
	}
	if plan.IsNull() || plan.IsUnknown() {
		return state.ValueString() != ""
	}
	return false
}

func tuneLevelToSyncLevel(level int) (string, error) {
	switch level {
	case 0:
		return "none", nil
	case 1:
		return "filesystem", nil
	case 2:
		return "file", nil
	default:
		return "", fmt.Errorf("unsupported tune_level %d; valid values are 0-2", level)
	}
}

func syncLevelToTuneLevel(syncLevel string) (int, bool) {
	switch strings.ToLower(syncLevel) {
	case "none":
		return 0, true
	case "filesystem":
		return 1, true
	case "file":
		return 2, true
	default:
		return 0, false
	}
}

func stringValueOrNull(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func intValueOrNull(ptr *int) types.Int64 {
	if ptr == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*ptr))
}

func boolValueOrNull(ptr *bool) types.Bool {
	if ptr == nil {
		return types.BoolNull()
	}
	return types.BoolValue(*ptr)
}

// datastoreToState converts a datastore struct to Terraform state
func (r *datastoreResource) datastoreToState(ds *datastores.Datastore, state *datastoreResourceModel) error {
	state.Name = types.StringValue(ds.Name)

	// Common fields
	state.Path = stringValueOrNull(ds.Path)
	state.Comment = stringValueOrNull(ds.Comment)
	state.Disabled = boolValueOrNull(ds.Disabled)
	state.GCSchedule = stringValueOrNull(ds.GCSchedule)
	state.PruneSchedule = stringValueOrNull(ds.PruneSchedule)
	state.KeepLast = intValueOrNull(ds.KeepLast)
	state.KeepHourly = intValueOrNull(ds.KeepHourly)
	state.KeepDaily = intValueOrNull(ds.KeepDaily)
	state.KeepWeekly = intValueOrNull(ds.KeepWeekly)
	state.KeepMonthly = intValueOrNull(ds.KeepMonthly)
	state.KeepYearly = intValueOrNull(ds.KeepYearly)

	// ZFS-specific
	state.ZFSPool = stringValueOrNull(ds.ZFSPool)
	state.ZFSDataset = stringValueOrNull(ds.ZFSDataset)
	state.BlockSize = stringValueOrNull(ds.BlockSize)
	state.Compression = stringValueOrNull(ds.Compression)

	// LVM-specific
	state.VolumeGroup = stringValueOrNull(ds.VolumeGroup)
	state.ThinPool = stringValueOrNull(ds.ThinPool)

	// Network storage
	state.Server = stringValueOrNull(ds.Server)
	state.Export = stringValueOrNull(ds.Export)
	state.Username = stringValueOrNull(ds.Username)
	// Password is not returned by API; keep whatever is currently in state
	state.Domain = stringValueOrNull(ds.Domain)
	state.Share = stringValueOrNull(ds.Share)
	state.SubDir = stringValueOrNull(ds.SubDir)
	state.Options = stringValueOrNull(ds.Options)

	// Advanced options
	state.NotifyUser = stringValueOrNull(ds.NotifyUser)
	state.NotifyLevel = stringValueOrNull(ds.NotifyLevel)
	state.NotificationMode = stringValueOrNull(ds.NotificationMode)
	state.VerifyNew = boolValueOrNull(ds.VerifyNew)
	state.ReuseDatastore = boolValueOrNull(ds.ReuseDatastore)
	state.OverwriteInUse = boolValueOrNull(ds.OverwriteInUse)
	state.Fingerprint = stringValueOrNull(ds.Fingerprint)
	state.Digest = types.StringValue(ds.Digest)

	if ds.Notify != nil {
		notify := &notifyModel{
			GC:     stringValueOrNull(ds.Notify.GC),
			Prune:  stringValueOrNull(ds.Notify.Prune),
			Sync:   stringValueOrNull(ds.Notify.Sync),
			Verify: stringValueOrNull(ds.Notify.Verify),
		}
		if notify.GC.IsNull() && notify.Prune.IsNull() && notify.Sync.IsNull() && notify.Verify.IsNull() {
			state.Notify = nil
		} else {
			state.Notify = notify
		}
	} else {
		state.Notify = nil
	}

	if ds.MaintenanceMode != nil {
		mm := &maintenanceModeModel{
			Type:    stringValueOrNull(ds.MaintenanceMode.Type),
			Message: stringValueOrNull(ds.MaintenanceMode.Message),
		}
		if mm.Type.IsNull() && mm.Message.IsNull() {
			state.MaintenanceMode = nil
		} else {
			state.MaintenanceMode = mm
		}
	} else {
		state.MaintenanceMode = nil
	}

	if ds.Tuning != nil && !isEmptyTuning(ds.Tuning) {
		tuning := &tuningModel{
			ChunkOrder:         stringValueOrNull(ds.Tuning.ChunkOrder),
			GCAtimeCutoff:      intValueOrNull(ds.Tuning.GCAtimeCutoff),
			GCAtimeSafetyCheck: boolValueOrNull(ds.Tuning.GCAtimeSafetyCheck),
			GCCacheCapacity:    intValueOrNull(ds.Tuning.GCCacheCapacity),
			SyncLevel:          stringValueOrNull(ds.Tuning.SyncLevel),
		}
		state.Tuning = tuning

		if level, ok := syncLevelToTuneLevel(ds.Tuning.SyncLevel); ok {
			state.TuneLevel = types.Int64Value(int64(level))
		} else {
			state.TuneLevel = types.Int64Null()
		}
	} else {
		state.Tuning = nil
		state.TuneLevel = types.Int64Null()
	}

	// S3-specific
	state.S3Client = stringValueOrNull(ds.S3Client)
	state.S3Bucket = stringValueOrNull(ds.S3Bucket)

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
