/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package datastores

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/datastores"
)

var (
	_ datasource.DataSource              = &datastoreDataSource{}
	_ datasource.DataSourceWithConfigure = &datastoreDataSource{}
)

// NewDatastoreDataSource is a helper function to simplify the provider implementation.
func NewDatastoreDataSource() datasource.DataSource {
	return &datastoreDataSource{}
}

// datastoreDataSource is the data source implementation.
type datastoreDataSource struct {
	client *pbs.Client
}

// datastoreDataSourceModel maps the data source schema data.
type datastoreDataSourceModel struct {
	Name             types.String          `tfsdk:"name"`
	Path             types.String          `tfsdk:"path"`
	Removable        types.Bool            `tfsdk:"removable"`
	BackingDevice    types.String          `tfsdk:"backing_device"`
	Comment          types.String          `tfsdk:"comment"`
	Disabled         types.Bool            `tfsdk:"disabled"`
	GCSchedule       types.String          `tfsdk:"gc_schedule"`
	PruneSchedule    types.String          `tfsdk:"prune_schedule"`
	KeepLast         types.Int64           `tfsdk:"keep_last"`
	KeepHourly       types.Int64           `tfsdk:"keep_hourly"`
	KeepDaily        types.Int64           `tfsdk:"keep_daily"`
	KeepWeekly       types.Int64           `tfsdk:"keep_weekly"`
	KeepMonthly      types.Int64           `tfsdk:"keep_monthly"`
	KeepYearly       types.Int64           `tfsdk:"keep_yearly"`
	NotifyUser       types.String          `tfsdk:"notify_user"`
	NotifyLevel      types.String          `tfsdk:"notify_level"`
	NotificationMode types.String          `tfsdk:"notification_mode"`
	Notify           *notifyModel          `tfsdk:"notify"`
	MaintenanceMode  *maintenanceModeModel `tfsdk:"maintenance_mode"`
	VerifyNew        types.Bool            `tfsdk:"verify_new"`
	Tuning           *tuningModel          `tfsdk:"tuning"`
	Fingerprint      types.String          `tfsdk:"fingerprint"`
	Digest           types.String          `tfsdk:"digest"`
	S3Client         types.String          `tfsdk:"s3_client"`
	S3Bucket         types.String          `tfsdk:"s3_bucket"`
}

type notifyModel struct {
	GC     types.String `tfsdk:"gc"`
	Prune  types.String `tfsdk:"prune"`
	Sync   types.String `tfsdk:"sync"`
	Verify types.String `tfsdk:"verify"`
}

type maintenanceModeModel struct {
	Type    types.String `tfsdk:"type"`
	Message types.String `tfsdk:"message"`
}

type tuningModel struct {
	ChunkOrder         types.String `tfsdk:"chunk_order"`
	GCAtimeCutoff      types.Int64  `tfsdk:"gc_atime_cutoff"`
	GCAtimeSafetyCheck types.Bool   `tfsdk:"gc_atime_safety_check"`
	GCCacheCapacity    types.Int64  `tfsdk:"gc_cache_capacity"`
	SyncLevel          types.String `tfsdk:"sync_level"`
}

// Metadata returns the data source type name.
func (d *datastoreDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastore"
}

// Schema defines the schema for the data source.
func (d *datastoreDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Reads information about an existing PBS datastore configuration.",
		MarkdownDescription: "Reads information about an existing Proxmox Backup Server datastore configuration.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "Unique identifier for the datastore.",
				MarkdownDescription: "Unique identifier for the datastore.",
				Required:            true,
			},
			"path": schema.StringAttribute{
				Description:         "Filesystem path to the datastore data.",
				MarkdownDescription: "Filesystem path to the datastore data.",
				Computed:            true,
			},
			"removable": schema.BoolAttribute{
				Description:         "Whether the datastore is backed by a removable device.",
				MarkdownDescription: "Whether the datastore is backed by a removable device.",
				Computed:            true,
			},
			"backing_device": schema.StringAttribute{
				Description:         "UUID of the filesystem partition for a removable datastore.",
				MarkdownDescription: "UUID of the filesystem partition for a removable datastore.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "Description for the datastore.",
				MarkdownDescription: "Description for the datastore.",
				Computed:            true,
			},
			"disabled": schema.BoolAttribute{
				Description:         "Whether the datastore is disabled.",
				MarkdownDescription: "Whether the datastore is disabled.",
				Computed:            true,
			},
			"gc_schedule": schema.StringAttribute{
				Description:         "Garbage collection schedule in cron format.",
				MarkdownDescription: "Garbage collection schedule in cron format.",
				Computed:            true,
			},
			"prune_schedule": schema.StringAttribute{
				Description:         "Prune schedule in cron format (deprecated in PBS 4.0+).",
				MarkdownDescription: "Prune schedule in cron format (deprecated in PBS 4.0+).",
				Computed:            true,
			},
			"keep_last": schema.Int64Attribute{
				Description:         "Number of latest backups to keep.",
				MarkdownDescription: "Number of latest backups to keep.",
				Computed:            true,
			},
			"keep_hourly": schema.Int64Attribute{
				Description:         "Number of hourly backups to keep.",
				MarkdownDescription: "Number of hourly backups to keep.",
				Computed:            true,
			},
			"keep_daily": schema.Int64Attribute{
				Description:         "Number of daily backups to keep.",
				MarkdownDescription: "Number of daily backups to keep.",
				Computed:            true,
			},
			"keep_weekly": schema.Int64Attribute{
				Description:         "Number of weekly backups to keep.",
				MarkdownDescription: "Number of weekly backups to keep.",
				Computed:            true,
			},
			"keep_monthly": schema.Int64Attribute{
				Description:         "Number of monthly backups to keep.",
				MarkdownDescription: "Number of monthly backups to keep.",
				Computed:            true,
			},
			"keep_yearly": schema.Int64Attribute{
				Description:         "Number of yearly backups to keep.",
				MarkdownDescription: "Number of yearly backups to keep.",
				Computed:            true,
			},
			"notify_user": schema.StringAttribute{
				Description:         "User to send notifications to.",
				MarkdownDescription: "User to send notifications to.",
				Computed:            true,
			},
			"notify_level": schema.StringAttribute{
				Description:         "Notification level.",
				MarkdownDescription: "Notification level.",
				Computed:            true,
			},
			"notification_mode": schema.StringAttribute{
				Description:         "Notification delivery mode.",
				MarkdownDescription: "Notification delivery mode.",
				Computed:            true,
			},
			"notify": schema.SingleNestedAttribute{
				Description:         "Per-job notification settings.",
				MarkdownDescription: "Per-job notification settings.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"gc": schema.StringAttribute{
						Description:         "Garbage collection notification level.",
						MarkdownDescription: "Garbage collection notification level.",
						Computed:            true,
					},
					"prune": schema.StringAttribute{
						Description:         "Prune job notification level.",
						MarkdownDescription: "Prune job notification level.",
						Computed:            true,
					},
					"sync": schema.StringAttribute{
						Description:         "Sync job notification level.",
						MarkdownDescription: "Sync job notification level.",
						Computed:            true,
					},
					"verify": schema.StringAttribute{
						Description:         "Verification job notification level.",
						MarkdownDescription: "Verification job notification level.",
						Computed:            true,
					},
				},
			},
			"maintenance_mode": schema.SingleNestedAttribute{
				Description:         "Maintenance mode configuration.",
				MarkdownDescription: "Maintenance mode configuration.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						Description:         "Maintenance mode type.",
						MarkdownDescription: "Maintenance mode type.",
						Computed:            true,
					},
					"message": schema.StringAttribute{
						Description:         "Message shown in maintenance mode.",
						MarkdownDescription: "Message shown in maintenance mode.",
						Computed:            true,
					},
				},
			},
			"verify_new": schema.BoolAttribute{
				Description:         "Verify newly created snapshots immediately after backup.",
				MarkdownDescription: "Verify newly created snapshots immediately after backup.",
				Computed:            true,
			},
			"tuning": schema.SingleNestedAttribute{
				Description:         "Advanced tuning options for datastore behaviour.",
				MarkdownDescription: "Advanced tuning options for datastore behaviour.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"chunk_order": schema.StringAttribute{
						Description:         "Chunk iteration order.",
						MarkdownDescription: "Chunk iteration order.",
						Computed:            true,
					},
					"gc_atime_cutoff": schema.Int64Attribute{
						Description:         "Garbage collection access time cutoff (seconds).",
						MarkdownDescription: "Garbage collection access time cutoff (seconds).",
						Computed:            true,
					},
					"gc_atime_safety_check": schema.BoolAttribute{
						Description:         "Enable garbage collection access time safety check.",
						MarkdownDescription: "Enable garbage collection access time safety check.",
						Computed:            true,
					},
					"gc_cache_capacity": schema.Int64Attribute{
						Description:         "Garbage collection cache capacity.",
						MarkdownDescription: "Garbage collection cache capacity.",
						Computed:            true,
					},
					"sync_level": schema.StringAttribute{
						Description:         "Datastore fsync level.",
						MarkdownDescription: "Datastore fsync level.",
						Computed:            true,
					},
				},
			},
			"fingerprint": schema.StringAttribute{
				Description:         "Certificate fingerprint for secure connections.",
				MarkdownDescription: "Certificate fingerprint for secure connections.",
				Computed:            true,
			},
			"digest": schema.StringAttribute{
				Description:         "Opaque digest returned by PBS.",
				MarkdownDescription: "Opaque digest returned by PBS.",
				Computed:            true,
			},
			"s3_client": schema.StringAttribute{
				Description:         "S3 endpoint ID for S3 datastores.",
				MarkdownDescription: "S3 endpoint ID for S3 datastores.",
				Computed:            true,
			},
			"s3_bucket": schema.StringAttribute{
				Description:         "S3 bucket name for S3 datastores.",
				MarkdownDescription: "S3 bucket name for S3 datastores.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *datastoreDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.DataSource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.DataSource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = cfg.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *datastoreDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datastoreDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get datastore from API
	ds, err := d.client.Datastores.GetDatastore(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datastore",
			fmt.Sprintf("Could not read datastore %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	// Map API response to state
	if err := datastoreToState(ds, &state); err != nil {
		resp.Diagnostics.AddError(
			"Error Converting Datastore",
			fmt.Sprintf("Could not convert datastore to state: %s", err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// datastoreToState converts a datastore struct to Terraform state
func datastoreToState(ds *datastores.Datastore, state *datastoreDataSourceModel) error {
	state.Name = types.StringValue(ds.Name)
	state.Path = stringValueOrNull(ds.Path)
	state.BackingDevice = stringValueOrNull(ds.BackingDevice)
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
	state.NotifyUser = stringValueOrNull(ds.NotifyUser)
	state.NotifyLevel = stringValueOrNull(ds.NotifyLevel)
	state.NotificationMode = stringValueOrNull(ds.NotificationMode)
	state.VerifyNew = boolValueOrNull(ds.VerifyNew)
	state.Fingerprint = stringValueOrNull(ds.Fingerprint)
	state.Digest = types.StringValue(ds.Digest)

	// Determine if removable
	isRemovable := ds.BackendType == "removable" || ds.BackingDevice != ""
	state.Removable = types.BoolValue(isRemovable)

	// Notify settings
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

	// Maintenance mode
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

	// Tuning settings
	if ds.Tuning != nil && !isEmptyTuning(ds.Tuning) {
		tuning := &tuningModel{
			ChunkOrder:         stringValueOrNull(ds.Tuning.ChunkOrder),
			GCAtimeCutoff:      intValueOrNull(ds.Tuning.GCAtimeCutoff),
			GCAtimeSafetyCheck: boolValueOrNull(ds.Tuning.GCAtimeSafetyCheck),
			GCCacheCapacity:    intValueOrNull(ds.Tuning.GCCacheCapacity),
			SyncLevel:          stringValueOrNull(ds.Tuning.SyncLevel),
		}
		state.Tuning = tuning
	} else {
		state.Tuning = nil
	}

	// S3 backend
	state.S3Client = stringValueOrNull(ds.S3Client)
	state.S3Bucket = stringValueOrNull(ds.S3Bucket)

	return nil
}

// Helper functions

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

func isEmptyTuning(t *datastores.DatastoreTuning) bool {
	if t == nil {
		return true
	}
	return t.ChunkOrder == "" && t.GCAtimeCutoff == nil && t.GCAtimeSafetyCheck == nil && t.GCCacheCapacity == nil && t.SyncLevel == ""
}
