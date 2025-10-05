/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package jobs

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &syncJobResource{}
	_ resource.ResourceWithConfigure   = &syncJobResource{}
	_ resource.ResourceWithImportState = &syncJobResource{}
)

// NewSyncJobResource is a helper function to simplify the provider implementation.
func NewSyncJobResource() resource.Resource {
	return &syncJobResource{}
}

// syncJobResource is the resource implementation.
type syncJobResource struct {
	client *pbs.Client
}

// syncJobResourceModel maps the resource schema data.
type syncJobResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Store          types.String `tfsdk:"store"`
	Schedule       types.String `tfsdk:"schedule"`
	Remote         types.String `tfsdk:"remote"`
	RemoteStore    types.String `tfsdk:"remote_store"`
	Namespace      types.String `tfsdk:"namespace"`
	GroupFilter    types.List   `tfsdk:"group_filter"`
	RemoveVanished types.Bool   `tfsdk:"remove_vanished"`
	Owner          types.String `tfsdk:"owner"`
	RateLimitKbps  types.Int64  `tfsdk:"rate_limit_kbps"`
	Comment        types.String `tfsdk:"comment"`
	Disable        types.Bool   `tfsdk:"disable"`
}

// Metadata returns the resource type name.
func (r *syncJobResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sync_job"
}

// Schema defines the schema for the resource.
func (r *syncJobResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a PBS sync job for automated remote datastore synchronization.",
		MarkdownDescription: `Manages a PBS sync job.

Sync jobs pull backups from a remote PBS server to the local datastore, enabling off-site 
backup replication. You can filter which backup groups to sync and control bandwidth usage 
with rate limiting. The ` + "`remove_vanished`" + ` option keeps the local copy synchronized 
by removing backups that no longer exist on the remote.`,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description:         "The unique identifier for the sync job.",
				MarkdownDescription: "The unique identifier for the sync job.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"store": schema.StringAttribute{
				Description:         "The local datastore name where backups will be synced to.",
				MarkdownDescription: "The local datastore name where backups will be synced to.",
				Required:            true,
			},
			"schedule": schema.StringAttribute{
				Description:         "When to run the sync job (systemd calendar event format).",
				MarkdownDescription: "When to run the sync job. Uses systemd calendar event format (e.g., `hourly`, `*:00/15`, `Mon,Wed,Fri 02:00`).",
				Required:            true,
			},
			"remote": schema.StringAttribute{
				Description:         "The remote server name (configured in PBS remotes).",
				MarkdownDescription: "The remote server name (configured in PBS remotes).",
				Required:            true,
			},
			"remote_store": schema.StringAttribute{
				Description:         "The datastore name on the remote server.",
				MarkdownDescription: "The datastore name on the remote server.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				Description:         "Namespace to sync from (optional, can use depth/pattern matching).",
				MarkdownDescription: "Namespace to sync from. Optional, supports depth and pattern matching (e.g., `ns1`, `ns1/sub`).",
				Optional:            true,
			},
			"group_filter": schema.ListAttribute{
				Description:         "List of backup group filters (regex patterns).",
				MarkdownDescription: "List of backup group filters using regex patterns. Only groups matching these patterns will be synced.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"remove_vanished": schema.BoolAttribute{
				Description:         "Remove backups that no longer exist on the remote.",
				MarkdownDescription: "Remove backups locally that no longer exist on the remote. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"owner": schema.StringAttribute{
				Description:         "Owner of the synced backups (user ID).",
				MarkdownDescription: "Owner user ID for the synced backups. Optional.",
				Optional:            true,
			},
			"rate_limit_kbps": schema.Int64Attribute{
				Description:         "Transfer rate limit in KiB/s.",
				MarkdownDescription: "Transfer rate limit in KiB/s. Defaults to unlimited (0).",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(0),
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this sync job.",
				MarkdownDescription: "A comment describing this sync job.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this sync job.",
				MarkdownDescription: "Disable this sync job. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *syncJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*pbs.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *pbs.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *syncJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan syncJobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := &jobs.SyncJob{
		ID:          plan.ID.ValueString(),
		Store:       plan.Store.ValueString(),
		Schedule:    plan.Schedule.ValueString(),
		Remote:      plan.Remote.ValueString(),
		RemoteStore: plan.RemoteStore.ValueString(),
	}

	if !plan.Namespace.IsNull() {
		job.NamespaceRE = plan.Namespace.ValueString()
	}
	if !plan.GroupFilter.IsNull() {
		var filters []string
		diags := plan.GroupFilter.ElementsAs(ctx, &filters, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		job.GroupFilter = filters
	}
	if !plan.RemoveVanished.IsNull() {
		removeVanished := plan.RemoveVanished.ValueBool()
		job.RemoveVanished = &removeVanished
	}
	if !plan.Owner.IsNull() {
		job.Owner = plan.Owner.ValueString()
	}
	if !plan.RateLimitKbps.IsNull() {
		rateLimit := int(plan.RateLimitKbps.ValueInt64())
		job.RateLimitKBPS = &rateLimit
	}
	if !plan.Comment.IsNull() {
		job.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		job.Disable = &disable
	}

	err := r.client.Jobs.CreateSyncJob(ctx, job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating sync job",
			fmt.Sprintf("Could not create sync job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *syncJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state syncJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, err := r.client.Jobs.GetSyncJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sync job",
			fmt.Sprintf("Could not read sync job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}

	// Update state
	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)
	state.Remote = types.StringValue(job.Remote)
	state.RemoteStore = types.StringValue(job.RemoteStore)

	if job.NamespaceRE != "" {
		state.Namespace = types.StringValue(job.NamespaceRE)
	} else {
		state.Namespace = types.StringNull()
	}
	if len(job.GroupFilter) > 0 {
		filters, diags := types.ListValueFrom(ctx, types.StringType, job.GroupFilter)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.GroupFilter = filters
	} else {
		state.GroupFilter = types.ListNull(types.StringType)
	}
	if job.RemoveVanished != nil {
		state.RemoveVanished = types.BoolValue(*job.RemoveVanished)
	} else {
		state.RemoveVanished = types.BoolValue(false)
	}
	if job.Owner != "" {
		state.Owner = types.StringValue(job.Owner)
	} else {
		state.Owner = types.StringNull()
	}
	if job.RateLimitKBPS != nil {
		state.RateLimitKbps = types.Int64Value(int64(*job.RateLimitKBPS))
	} else {
		state.RateLimitKbps = types.Int64Value(0)
	}
	if job.Comment != "" {
		state.Comment = types.StringValue(job.Comment)
	} else {
		state.Comment = types.StringNull()
	}
	if job.Disable != nil {
		state.Disable = types.BoolValue(*job.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *syncJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan syncJobResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job := &jobs.SyncJob{
		ID:          plan.ID.ValueString(),
		Store:       plan.Store.ValueString(),
		Schedule:    plan.Schedule.ValueString(),
		Remote:      plan.Remote.ValueString(),
		RemoteStore: plan.RemoteStore.ValueString(),
	}

	if !plan.Namespace.IsNull() {
		job.NamespaceRE = plan.Namespace.ValueString()
	}
	if !plan.GroupFilter.IsNull() {
		var filters []string
		diags := plan.GroupFilter.ElementsAs(ctx, &filters, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		job.GroupFilter = filters
	}
	if !plan.RemoveVanished.IsNull() {
		removeVanished := plan.RemoveVanished.ValueBool()
		job.RemoveVanished = &removeVanished
	}
	if !plan.Owner.IsNull() {
		job.Owner = plan.Owner.ValueString()
	}
	if !plan.RateLimitKbps.IsNull() {
		rateLimit := int(plan.RateLimitKbps.ValueInt64())
		job.RateLimitKBPS = &rateLimit
	}
	if !plan.Comment.IsNull() {
		job.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		job.Disable = &disable
	}

	err := r.client.Jobs.UpdateSyncJob(ctx, plan.ID.ValueString(), job)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating sync job",
			fmt.Sprintf("Could not update sync job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *syncJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state syncJobResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Jobs.DeleteSyncJob(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting sync job",
			fmt.Sprintf("Could not delete sync job %s: %s", state.ID.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *syncJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
