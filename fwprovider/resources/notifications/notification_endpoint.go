/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

var (
	_ resource.Resource                = &notificationEndpointResource{}
	_ resource.ResourceWithConfigure   = &notificationEndpointResource{}
	_ resource.ResourceWithImportState = &notificationEndpointResource{}
)

func NewNotificationEndpointResource() resource.Resource {
	return &notificationEndpointResource{}
}

type notificationEndpointResource struct {
	client *pbs.Client
}

type notificationEndpointResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Targets types.List   `tfsdk:"targets"`
	Comment types.String `tfsdk:"comment"`
	Disable types.Bool   `tfsdk:"disable"`
	Origin  types.String `tfsdk:"origin"`
}

func (r *notificationEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_endpoint"
}

func (r *notificationEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description:         "Manages a notification endpoint (target group) in PBS.",
		MarkdownDescription: "Manages a notification endpoint (target group) that aggregates multiple notification targets.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "Unique name of the notification endpoint.",
				MarkdownDescription: "Unique name of the notification endpoint.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"targets": schema.ListAttribute{
				Description:         "List of notification target names associated with this endpoint.",
				MarkdownDescription: "List of notification target names associated with this endpoint.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				Description:         "Optional comment describing this endpoint.",
				MarkdownDescription: "Optional comment describing this endpoint.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this notification endpoint.",
				MarkdownDescription: "Disable this notification endpoint. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"origin": schema.StringAttribute{
				Description:         "Origin of this configuration as reported by PBS.",
				MarkdownDescription: "Origin of this configuration as reported by PBS (e.g., `user`, `builtin`).",
				Computed:            true,
			},
		},
	}
}

func (r *notificationEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *notificationEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	supported, err := r.client.Notifications.SupportsNotificationEndpoints(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking notification endpoint support",
			fmt.Sprintf("Failed to verify notification endpoint capability: %s", err.Error()),
		)
		return
	}
	if !supported {
		resp.Diagnostics.AddError(
			"Notification endpoints not supported",
			"Notification endpoints require Proxmox Backup Server 4.0 or later. The connected PBS instance does not expose the /config/notifications/endpoints API. Please upgrade PBS before using this resource.",
		)
		return
	}

	endpoint := &notifications.NotificationEndpoint{
		Name: plan.Name.ValueString(),
	}

	if !plan.Targets.IsNull() && !plan.Targets.IsUnknown() {
		var targets []string
		diags := plan.Targets.ElementsAs(ctx, &targets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		endpoint.Targets = targets
	}
	if !plan.Comment.IsNull() {
		endpoint.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		endpoint.Disable = &disable
	}

	if err := r.client.Notifications.CreateNotificationEndpoint(ctx, endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Error creating notification endpoint",
			fmt.Sprintf("Could not create notification endpoint %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	created, err := r.client.Notifications.GetNotificationEndpoint(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created notification endpoint",
			fmt.Sprintf("Created notification endpoint %s but could not read it back: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	r.setModelFromAPI(ctx, &plan, created, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationEndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state notificationEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint, err := r.client.Notifications.GetNotificationEndpoint(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading notification endpoint",
			fmt.Sprintf("Could not read notification endpoint %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	r.setModelFromAPI(ctx, &state, endpoint, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *notificationEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	supported, err := r.client.Notifications.SupportsNotificationEndpoints(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error checking notification endpoint support",
			fmt.Sprintf("Failed to verify notification endpoint capability: %s", err.Error()),
		)
		return
	}
	if !supported {
		resp.Diagnostics.AddError(
			"Notification endpoints not supported",
			"Notification endpoints require Proxmox Backup Server 4.0 or later. The connected PBS instance no longer exposes the /config/notifications/endpoints API. Please upgrade PBS before updating this resource.",
		)
		return
	}

	endpoint := &notifications.NotificationEndpoint{}

	if plan.Targets.IsNull() {
		endpoint.Targets = []string{}
	} else if !plan.Targets.IsUnknown() {
		var targets []string
		diags := plan.Targets.ElementsAs(ctx, &targets, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		endpoint.Targets = targets
	}
	if plan.Comment.IsNull() {
		endpoint.Comment = ""
	} else {
		endpoint.Comment = plan.Comment.ValueString()
	}
	if plan.Disable.IsNull() {
		falseVal := false
		endpoint.Disable = &falseVal
	} else {
		disable := plan.Disable.ValueBool()
		endpoint.Disable = &disable
	}

	if err := r.client.Notifications.UpdateNotificationEndpoint(ctx, plan.Name.ValueString(), endpoint); err != nil {
		resp.Diagnostics.AddError(
			"Error updating notification endpoint",
			fmt.Sprintf("Could not update notification endpoint %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	updated, err := r.client.Notifications.GetNotificationEndpoint(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated notification endpoint",
			fmt.Sprintf("Updated notification endpoint %s but could not read it back: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	r.setModelFromAPI(ctx, &plan, updated, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *notificationEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.Notifications.DeleteNotificationEndpoint(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Error deleting notification endpoint",
			fmt.Sprintf("Could not delete notification endpoint %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

func (r *notificationEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func (r *notificationEndpointResource) setModelFromAPI(ctx context.Context, model *notificationEndpointResourceModel, endpoint *notifications.NotificationEndpoint, diagnostics *diag.Diagnostics) {
	model.Name = types.StringValue(endpoint.Name)

	if endpoint.Targets != nil {
		targets, targetDiags := types.ListValueFrom(ctx, types.StringType, endpoint.Targets)
		diagnostics.Append(targetDiags...)
		if diagnostics.HasError() {
			return
		}
		model.Targets = targets
	} else {
		model.Targets = types.ListNull(types.StringType)
	}

	if endpoint.Comment != "" {
		model.Comment = types.StringValue(endpoint.Comment)
	} else {
		model.Comment = types.StringNull()
	}

	if endpoint.Disable != nil {
		model.Disable = types.BoolValue(*endpoint.Disable)
	} else {
		model.Disable = types.BoolNull()
	}

	if endpoint.Origin != "" {
		model.Origin = types.StringValue(endpoint.Origin)
	} else {
		model.Origin = types.StringNull()
	}
}
