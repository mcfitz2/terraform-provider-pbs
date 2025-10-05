/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &notificationEndpointResource{}
	_ resource.ResourceWithConfigure   = &notificationEndpointResource{}
	_ resource.ResourceWithImportState = &notificationEndpointResource{}
)

// NewNotificationEndpointResource is a helper function to simplify the provider implementation.
func NewNotificationEndpointResource() resource.Resource {
	return &notificationEndpointResource{}
}

// notificationEndpointResource is the resource implementation.
type notificationEndpointResource struct {
	client *pbs.Client
}

// notificationEndpointResourceModel maps the resource schema data.
type notificationEndpointResourceModel struct {
	Name    types.String `tfsdk:"name"`
	Targets types.List   `tfsdk:"targets"`
	Comment types.String `tfsdk:"comment"`
	Disable types.Bool   `tfsdk:"disable"`
}

// Metadata returns the resource type name.
func (r *notificationEndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification_endpoint"
}

// Schema defines the schema for the resource.
func (r *notificationEndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a notification endpoint (group of notification targets).",
		MarkdownDescription: `Manages a notification endpoint.

Notification endpoints group multiple notification targets together, allowing you to send 
notifications to several destinations (SMTP, Gotify, Webhook, etc.) at once. This is useful 
for routing system alerts to multiple channels simultaneously.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique name identifier for the notification endpoint.",
				MarkdownDescription: "The unique name identifier for the notification endpoint.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"targets": schema.ListAttribute{
				Description:         "List of notification target names to include in this endpoint.",
				MarkdownDescription: "List of notification target names (SMTP, Gotify, Sendmail, Webhook) to include in this endpoint.",
				ElementType:         types.StringType,
				Optional:            true,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this notification endpoint.",
				MarkdownDescription: "A comment describing this notification endpoint.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this notification endpoint.",
				MarkdownDescription: "Disable this notification endpoint. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *notificationEndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *notificationEndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan notificationEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := &notifications.NotificationEndpoint{
		Name: plan.Name.ValueString(),
	}

	// Convert targets list
	if !plan.Targets.IsNull() {
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

	err := r.client.Notifications.CreateNotificationEndpoint(ctx, endpoint)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating notification endpoint",
			fmt.Sprintf("Could not create notification endpoint %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
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

	// Update state with values from API
	state.Name = types.StringValue(endpoint.Name)

	if len(endpoint.Targets) > 0 {
		targets, diags := types.ListValueFrom(ctx, types.StringType, endpoint.Targets)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Targets = targets
	} else {
		state.Targets = types.ListNull(types.StringType)
	}

	if endpoint.Comment != "" {
		state.Comment = types.StringValue(endpoint.Comment)
	} else {
		state.Comment = types.StringNull()
	}

	if endpoint.Disable != nil {
		state.Disable = types.BoolValue(*endpoint.Disable)
	} else {
		state.Disable = types.BoolValue(false)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *notificationEndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan notificationEndpointResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := &notifications.NotificationEndpoint{
		Name: plan.Name.ValueString(),
	}

	// Convert targets list
	if !plan.Targets.IsNull() {
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

	err := r.client.Notifications.UpdateNotificationEndpoint(ctx, plan.Name.ValueString(), endpoint)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating notification endpoint",
			fmt.Sprintf("Could not update notification endpoint %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *notificationEndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state notificationEndpointResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Notifications.DeleteNotificationEndpoint(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting notification endpoint",
			fmt.Sprintf("Could not delete notification endpoint %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource into Terraform state.
func (r *notificationEndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
