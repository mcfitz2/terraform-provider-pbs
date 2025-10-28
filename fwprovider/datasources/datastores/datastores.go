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

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
)

var (
	_ datasource.DataSource              = &datastoresDataSource{}
	_ datasource.DataSourceWithConfigure = &datastoresDataSource{}
)

// NewDatastoresDataSource is a helper function to simplify the provider implementation.
func NewDatastoresDataSource() datasource.DataSource {
	return &datastoresDataSource{}
}

// datastoresDataSource is the data source implementation.
type datastoresDataSource struct {
	client *pbs.Client
}

// datastoresDataSourceModel maps the data source schema data.
type datastoresDataSourceModel struct {
	Stores []datastoreDataSourceModel `tfsdk:"stores"`
}

// Metadata returns the data source type name.
func (d *datastoresDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_datastores"
}

// Schema defines the schema for the data source.
func (d *datastoresDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Lists all PBS datastore configurations.",
		MarkdownDescription: "Lists all Proxmox Backup Server datastore configurations.",
		Attributes: map[string]schema.Attribute{
			"stores": schema.ListNestedAttribute{
				Description:         "List of all datastores.",
				MarkdownDescription: "List of all datastores.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description:         "Unique identifier for the datastore.",
							MarkdownDescription: "Unique identifier for the datastore.",
							Computed:            true,
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
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *datastoresDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *datastoresDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datastoresDataSourceModel

	// Get all datastores from API
	datastores, err := d.client.Datastores.ListDatastores(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Datastores",
			fmt.Sprintf("Could not list datastores: %s", err.Error()),
		)
		return
	}

	// Map API response to state
	state.Stores = make([]datastoreDataSourceModel, 0, len(datastores))
	for _, ds := range datastores {
		var storeModel datastoreDataSourceModel
		// Get full details for each datastore
		fullDs, err := d.client.Datastores.GetDatastore(ctx, ds.Name)
		if err != nil {
			// Log warning but continue with other datastores
			resp.Diagnostics.AddWarning(
				"Error Reading Datastore Details",
				fmt.Sprintf("Could not read details for datastore %s: %s", ds.Name, err.Error()),
			)
			continue
		}
		
		if err := datastoreToState(fullDs, &storeModel); err != nil {
			resp.Diagnostics.AddWarning(
				"Error Converting Datastore",
				fmt.Sprintf("Could not convert datastore %s to state: %s", ds.Name, err.Error()),
			)
			continue
		}
		state.Stores = append(state.Stores, storeModel)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
