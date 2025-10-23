/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package jobs

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/jobs"
)

var (
	_ resource.Resource                = &syncJobResource{}
	_ resource.ResourceWithConfigure   = &syncJobResource{}
	_ resource.ResourceWithImportState = &syncJobResource{}
)

var groupFilterRegex = regexp.MustCompile(`^[^/\s]+/[^/\s]+(?:/[^/\s]+)?$`)

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
	ID              types.String `tfsdk:"id"`
	Store           types.String `tfsdk:"store"`
	Schedule        types.String `tfsdk:"schedule"`
	Remote          types.String `tfsdk:"remote"`
	RemoteStore     types.String `tfsdk:"remote_store"`
	RemoteNamespace types.String `tfsdk:"remote_namespace"`
	Namespace       types.String `tfsdk:"namespace"`
	MaxDepth        types.Int64  `tfsdk:"max_depth"`
	GroupFilter     types.List   `tfsdk:"group_filter"`
	RemoveVanished  types.Bool   `tfsdk:"remove_vanished"`
	ResyncCorrupt   types.Bool   `tfsdk:"resync_corrupt"`
	EncryptedOnly   types.Bool   `tfsdk:"encrypted_only"`
	VerifiedOnly    types.Bool   `tfsdk:"verified_only"`
	RunOnMount      types.Bool   `tfsdk:"run_on_mount"`
	TransferLast    types.Int64  `tfsdk:"transfer_last"`
	SyncDirection   types.String `tfsdk:"sync_direction"`
	Owner           types.String `tfsdk:"owner"`
	RateIn          types.String `tfsdk:"rate_in"`
	RateOut         types.String `tfsdk:"rate_out"`
	BurstIn         types.String `tfsdk:"burst_in"`
	BurstOut        types.String `tfsdk:"burst_out"`
	Comment         types.String `tfsdk:"comment"`
	Disable         types.Bool   `tfsdk:"disable"`
	Digest          types.String `tfsdk:"digest"`
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
			"remote_namespace": schema.StringAttribute{
				Description:         "Remote namespace to sync from (optional).",
				MarkdownDescription: "Remote namespace to sync from. Optional; leave empty to sync from the remote root namespace.",
				Optional:            true,
			},
			"namespace": schema.StringAttribute{
				Description:         "Local namespace where backups will be stored (optional).",
				MarkdownDescription: "Local namespace where backups will be stored. Optional; supports hierarchical namespaces such as `ns1/sub`.",
				Optional:            true,
			},
			"max_depth": schema.Int64Attribute{
				Description:         "Maximum namespace depth that will be traversed when syncing.",
				MarkdownDescription: "Maximum namespace depth that will be traversed when syncing. Must be greater than or equal to 0.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"group_filter": schema.ListAttribute{
				Description:         "List of backup group selectors using `<type>/<id>[/<namespace>]` syntax.",
				MarkdownDescription: "List of backup group selectors using `<type>/<id>[/<namespace>]` syntax. Only matching groups will be synced.",
				ElementType:         types.StringType,
				Optional:            true,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(stringvalidator.RegexMatches(groupFilterRegex, "must match `<type>/<id>[/<namespace>]`")),
				},
			},
			"remove_vanished": schema.BoolAttribute{
				Description:         "Remove backups that no longer exist on the remote.",
				MarkdownDescription: "Remove backups locally that no longer exist on the remote. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"resync_corrupt": schema.BoolAttribute{
				Description:         "Resync snapshots whose data is corrupt.",
				MarkdownDescription: "Resync snapshots whose data is corrupt. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"encrypted_only": schema.BoolAttribute{
				Description:         "Only sync encrypted backups.",
				MarkdownDescription: "Only sync encrypted backups. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"verified_only": schema.BoolAttribute{
				Description:         "Only sync backups that were verified successfully.",
				MarkdownDescription: "Only sync backups that were verified successfully. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"run_on_mount": schema.BoolAttribute{
				Description:         "Run the job immediately after the datastore is mounted.",
				MarkdownDescription: "Run the job immediately after the datastore is mounted. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"transfer_last": schema.Int64Attribute{
				Description:         "Only transfer backups newer than the last N seconds (0 disables).",
				MarkdownDescription: "Only transfer backups newer than the last N seconds. Set to 0 to disable.",
				Optional:            true,
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"sync_direction": schema.StringAttribute{
				Description:         "Direction of synchronization (`pull` or `push`).",
				MarkdownDescription: "Direction of synchronization. Must be either `pull` (default) or `push`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("pull", "push"),
				},
			},
			"owner": schema.StringAttribute{
				Description:         "Owner of the synced backups (user ID).",
				MarkdownDescription: "Owner user ID for the synced backups. Optional.",
				Optional:            true,
			},
			"rate_in": schema.StringAttribute{
				Description:         "Inbound transfer rate limit (PBS byte size format).",
				MarkdownDescription: "Inbound transfer rate limit in PBS byte size format (e.g., `10M` for 10 MiB/s). Leave empty for unlimited.",
				Optional:            true,
			},
			"rate_out": schema.StringAttribute{
				Description:         "Outbound transfer rate limit (PBS byte size format).",
				MarkdownDescription: "Outbound transfer rate limit in PBS byte size format (e.g., `10M`). Leave empty for unlimited.",
				Optional:            true,
			},
			"burst_in": schema.StringAttribute{
				Description:         "Inbound burst rate limit (PBS byte size format).",
				MarkdownDescription: "Inbound burst rate limit in PBS byte size format (e.g., `20M`). Leave empty for unlimited.",
				Optional:            true,
			},
			"burst_out": schema.StringAttribute{
				Description:         "Outbound burst rate limit (PBS byte size format).",
				MarkdownDescription: "Outbound burst rate limit in PBS byte size format (e.g., `20M`). Leave empty for unlimited.",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this sync job.",
				MarkdownDescription: "A comment describing this sync job.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this sync job without deleting it.",
				MarkdownDescription: "Disable this sync job without deleting it.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"digest": schema.StringAttribute{
				Description:         "Opaque digest returned by PBS for optimistic locking.",
				MarkdownDescription: "Opaque digest returned by PBS for optimistic locking.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *syncJobResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.Resource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = cfg.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *syncJobResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan syncJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	job, diags := buildSyncJobFromPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if job.Disable != nil && !*job.Disable {
		job.Disable = nil
	}

	if err := r.client.Jobs.CreateSyncJob(ctx, job); err != nil {
		resp.Diagnostics.AddError(
			"Error creating sync job",
			fmt.Sprintf("Could not create sync job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	createdJob, err := r.client.Jobs.GetSyncJob(ctx, job.ID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sync job",
			fmt.Sprintf("Could not read sync job %s after creation: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	var state syncJobResourceModel
	resp.Diagnostics.Append(setSyncStateFromAPI(ctx, createdJob, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *syncJobResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state syncJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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

	resp.Diagnostics.Append(setSyncStateFromAPI(ctx, job, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *syncJobResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan syncJobResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state syncJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if (plan.Digest.IsNull() || plan.Digest.IsUnknown()) && !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		plan.Digest = state.Digest
	}

	job, diags := buildSyncJobFromPlan(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	job.Delete = computeSyncDeletes(&plan, &state)

	if err := r.client.Jobs.UpdateSyncJob(ctx, plan.ID.ValueString(), job); err != nil {
		resp.Diagnostics.AddError(
			"Error updating sync job",
			fmt.Sprintf("Could not update sync job %s: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	updatedJob, err := r.client.Jobs.GetSyncJob(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading sync job",
			fmt.Sprintf("Could not read sync job %s after update: %s", plan.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(setSyncStateFromAPI(ctx, updatedJob, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *syncJobResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state syncJobResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	digest := ""
	if !state.Digest.IsNull() && !state.Digest.IsUnknown() {
		digest = state.Digest.ValueString()
	}

	if err := r.client.Jobs.DeleteSyncJob(ctx, state.ID.ValueString(), digest); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting sync job",
			fmt.Sprintf("Could not delete sync job %s: %s", state.ID.ValueString(), err.Error()),
		)
	}
}

// ImportState imports the resource into Terraform state.
func (r *syncJobResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildSyncJobFromPlan(ctx context.Context, plan *syncJobResourceModel) (*jobs.SyncJob, diag.Diagnostics) {
	var diags diag.Diagnostics

	job := &jobs.SyncJob{
		ID:          plan.ID.ValueString(),
		Store:       plan.Store.ValueString(),
		Schedule:    plan.Schedule.ValueString(),
		Remote:      plan.Remote.ValueString(),
		RemoteStore: plan.RemoteStore.ValueString(),
	}

	if !plan.RemoteNamespace.IsNull() && !plan.RemoteNamespace.IsUnknown() {
		job.RemoteNamespace = plan.RemoteNamespace.ValueString()
	}
	if !plan.Namespace.IsNull() && !plan.Namespace.IsUnknown() {
		job.Namespace = plan.Namespace.ValueString()
	}

	job.MaxDepth = intPointerFromAttr(plan.MaxDepth)
	job.TransferLast = intPointerFromAttr(plan.TransferLast)
	job.RemoveVanished = boolPointerFromAttr(plan.RemoveVanished)
	job.ResyncCorrupt = boolPointerFromAttr(plan.ResyncCorrupt)
	job.EncryptedOnly = boolPointerFromAttr(plan.EncryptedOnly)
	job.VerifiedOnly = boolPointerFromAttr(plan.VerifiedOnly)
	job.RunOnMount = boolPointerFromAttr(plan.RunOnMount)
	job.Disable = boolPointerFromAttr(plan.Disable)

	if !plan.SyncDirection.IsNull() && !plan.SyncDirection.IsUnknown() {
		job.SyncDirection = plan.SyncDirection.ValueString()
	}
	if !plan.Owner.IsNull() && !plan.Owner.IsUnknown() {
		job.Owner = plan.Owner.ValueString()
	}
	if !plan.RateIn.IsNull() && !plan.RateIn.IsUnknown() {
		job.RateIn = normalizeRateString(plan.RateIn.ValueString())
	}
	if !plan.RateOut.IsNull() && !plan.RateOut.IsUnknown() {
		job.RateOut = normalizeRateString(plan.RateOut.ValueString())
	}
	if !plan.BurstIn.IsNull() && !plan.BurstIn.IsUnknown() {
		job.BurstIn = normalizeRateString(plan.BurstIn.ValueString())
	}
	if !plan.BurstOut.IsNull() && !plan.BurstOut.IsUnknown() {
		job.BurstOut = normalizeRateString(plan.BurstOut.ValueString())
	}
	if !plan.Comment.IsNull() && !plan.Comment.IsUnknown() {
		job.Comment = plan.Comment.ValueString()
	}

	filters, filterDiags := stringListFromAttribute(ctx, plan.GroupFilter)
	diags.Append(filterDiags...)
	if filterDiags.HasError() {
		return nil, diags
	}
	if len(filters) > 0 {
		job.GroupFilter = filters
	}

	if !plan.Digest.IsNull() && !plan.Digest.IsUnknown() {
		job.Digest = plan.Digest.ValueString()
	}

	return job, diags
}

func computeSyncDeletes(plan, state *syncJobResourceModel) []string {
	if state == nil {
		return nil
	}

	var deletes []string

	if shouldDeleteStringAttr(plan.RemoteNamespace, state.RemoteNamespace) {
		deletes = append(deletes, "remote-ns")
	}
	if shouldDeleteStringAttr(plan.Namespace, state.Namespace) {
		deletes = append(deletes, "ns")
	}
	if shouldDeleteIntAttr(plan.MaxDepth, state.MaxDepth) {
		deletes = append(deletes, "max-depth")
	}
	if shouldDeleteListAttr(plan.GroupFilter, state.GroupFilter) {
		deletes = append(deletes, "group-filter")
	}
	if shouldDeleteBoolAttr(plan.RemoveVanished, state.RemoveVanished) {
		deletes = append(deletes, "remove-vanished")
	}
	if shouldDeleteBoolAttr(plan.ResyncCorrupt, state.ResyncCorrupt) {
		deletes = append(deletes, "resync-corrupt")
	}
	if shouldDeleteBoolAttr(plan.EncryptedOnly, state.EncryptedOnly) {
		deletes = append(deletes, "encrypted-only")
	}
	if shouldDeleteBoolAttr(plan.VerifiedOnly, state.VerifiedOnly) {
		deletes = append(deletes, "verified-only")
	}
	if shouldDeleteBoolAttr(plan.RunOnMount, state.RunOnMount) {
		deletes = append(deletes, "run-on-mount")
	}
	if shouldDeleteIntAttr(plan.TransferLast, state.TransferLast) {
		deletes = append(deletes, "transfer-last")
	}
	if shouldDeleteStringAttr(plan.SyncDirection, state.SyncDirection) {
		deletes = append(deletes, "sync-direction")
	}
	if shouldDeleteStringAttr(plan.Owner, state.Owner) {
		deletes = append(deletes, "owner")
	}
	if shouldDeleteStringAttr(plan.RateIn, state.RateIn) {
		deletes = append(deletes, "rate-in")
	}
	if shouldDeleteStringAttr(plan.RateOut, state.RateOut) {
		deletes = append(deletes, "rate-out")
	}
	if shouldDeleteStringAttr(plan.BurstIn, state.BurstIn) {
		deletes = append(deletes, "burst-in")
	}
	if shouldDeleteStringAttr(plan.BurstOut, state.BurstOut) {
		deletes = append(deletes, "burst-out")
	}
	if shouldDeleteStringAttr(plan.Comment, state.Comment) {
		deletes = append(deletes, "comment")
	}
	if shouldDeleteBoolAttr(plan.Disable, state.Disable) {
		deletes = append(deletes, "disable")
	}

	return deletes
}

func normalizeRateString(value string) string {
	v := strings.TrimSpace(value)
	if v == "" {
		return ""
	}

	v = strings.ReplaceAll(v, " ", "")
	v = strings.ToUpper(v)

	if strings.HasSuffix(v, "IB") && len(v) > 2 {
		v = v[:len(v)-2]
	}

	if len(v) > 1 && strings.HasSuffix(v, "B") {
		prev := v[len(v)-2]
		if (prev >= 'A' && prev <= 'Z') || prev == 'I' {
			v = v[:len(v)-1]
		}
	}

	return v
}

func setSyncStateFromAPI(ctx context.Context, job *jobs.SyncJob, state *syncJobResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	state.ID = types.StringValue(job.ID)
	state.Store = types.StringValue(job.Store)
	state.Schedule = types.StringValue(job.Schedule)
	state.Remote = types.StringValue(job.Remote)
	state.RemoteStore = types.StringValue(job.RemoteStore)
	state.RemoteNamespace = stringValueOrNull(job.RemoteNamespace)
	state.Namespace = stringValueOrNull(job.Namespace)
	state.MaxDepth = int64ValueOrNull(job.MaxDepth)

	groupFilter := types.ListNull(types.StringType)
	if len(job.GroupFilter) > 0 {
		listValue, listDiags := types.ListValueFrom(ctx, types.StringType, job.GroupFilter)
		diags.Append(listDiags...)
		if !listDiags.HasError() {
			groupFilter = listValue
		}
	}
	state.GroupFilter = groupFilter

	if job.RemoveVanished != nil {
		state.RemoveVanished = types.BoolValue(*job.RemoveVanished)
	} else {
		state.RemoveVanished = types.BoolValue(false)
	}
	if job.ResyncCorrupt != nil {
		state.ResyncCorrupt = types.BoolValue(*job.ResyncCorrupt)
	} else {
		state.ResyncCorrupt = types.BoolValue(false)
	}
	if job.EncryptedOnly != nil {
		state.EncryptedOnly = types.BoolValue(*job.EncryptedOnly)
	} else {
		state.EncryptedOnly = types.BoolValue(false)
	}
	if job.VerifiedOnly != nil {
		state.VerifiedOnly = types.BoolValue(*job.VerifiedOnly)
	} else {
		state.VerifiedOnly = types.BoolValue(false)
	}
	if job.RunOnMount != nil {
		state.RunOnMount = types.BoolValue(*job.RunOnMount)
	} else {
		state.RunOnMount = types.BoolValue(false)
	}

	state.TransferLast = int64ValueOrNull(job.TransferLast)
	state.SyncDirection = stringValueOrNull(job.SyncDirection)
	state.Owner = stringValueOrNull(job.Owner)
	state.RateIn = stringValueOrNull(normalizeRateString(job.RateIn))
	state.RateOut = stringValueOrNull(normalizeRateString(job.RateOut))
	state.BurstIn = stringValueOrNull(normalizeRateString(job.BurstIn))
	state.BurstOut = stringValueOrNull(normalizeRateString(job.BurstOut))
	state.Comment = stringValueOrNull(job.Comment)
	if job.Disable != nil {
		state.Disable = types.BoolValue(*job.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}
	state.Digest = stringValueOrNull(job.Digest)

	return diags
}
