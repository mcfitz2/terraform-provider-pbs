/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/notifications"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &webhookNotificationResource{}
	_ resource.ResourceWithConfigure   = &webhookNotificationResource{}
	_ resource.ResourceWithImportState = &webhookNotificationResource{}
)

// NewWebhookNotificationResource is a helper function to simplify the provider implementation.
func NewWebhookNotificationResource() resource.Resource {
	return &webhookNotificationResource{}
}

// webhookNotificationResource is the resource implementation.
type webhookNotificationResource struct {
	client *pbs.Client
}

// webhookNotificationResourceModel maps the resource schema data.
type webhookNotificationResourceModel struct {
	Name    types.String `tfsdk:"name"`
	URL     types.String `tfsdk:"url"`
	Body    types.String `tfsdk:"body"`
	Method  types.String `tfsdk:"method"`
	Headers types.Map    `tfsdk:"headers"`
	Secret  types.String `tfsdk:"secret"`
	Comment types.String `tfsdk:"comment"`
	Disable types.Bool   `tfsdk:"disable"`
}

// Metadata returns the resource type name.
func (r *webhookNotificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_notification"
}

// Schema defines the schema for the resource.
func (r *webhookNotificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Webhook notification target for PBS alerts and notifications.",
		MarkdownDescription: `Manages a Webhook notification target.

Configure a webhook endpoint to receive HTTP notifications from PBS about backup jobs,
verification tasks, and system events.`,
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description:         "The unique name identifier for the Webhook target.",
				MarkdownDescription: "The unique name identifier for the Webhook notification target.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Description:         "Webhook URL.",
				MarkdownDescription: "Webhook URL where notifications will be sent (e.g., `https://hooks.example.com/notify`).",
				Required:            true,
			},
			"body": schema.StringAttribute{
				Description:         "Custom request body template.",
				MarkdownDescription: "Custom request body template. Can use template variables for notification data.",
				Optional:            true,
			},
			"method": schema.StringAttribute{
				Description:         "HTTP method for webhook requests.",
				MarkdownDescription: "HTTP method for webhook requests. Valid values: `POST`, `PUT`. Defaults to `POST`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("POST"),
				Validators: []validator.String{
					stringvalidator.OneOf("POST", "PUT"),
				},
			},
			"headers": schema.MapAttribute{
				Description:         "Custom HTTP headers.",
				MarkdownDescription: "Custom HTTP headers to include in webhook requests. Specify as key-value pairs.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"secret": schema.StringAttribute{
				Description:         "Secret for HMAC signature.",
				MarkdownDescription: "Secret for HMAC-SHA256 signature. The signature will be sent in the `X-PBS-Signature` header.",
				Optional:            true,
				Sensitive:           true,
			},
			"comment": schema.StringAttribute{
				Description:         "A comment describing this notification target.",
				MarkdownDescription: "A comment describing this notification target.",
				Optional:            true,
			},
			"disable": schema.BoolAttribute{
				Description:         "Disable this notification target.",
				MarkdownDescription: "Disable this notification target. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *webhookNotificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *webhookNotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookNotificationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := &notifications.WebhookTarget{
		Name: plan.Name.ValueString(),
		URL:  plan.URL.ValueString(),
	}

	if !plan.Body.IsNull() {
		target.Body = plan.Body.ValueString()
	}
	if !plan.Method.IsNull() {
		target.Method = plan.Method.ValueString()
	}
	if !plan.Headers.IsNull() {
		headers := make(map[string]string)
		diags = plan.Headers.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.Headers = headers
	}
	if !plan.Secret.IsNull() {
		target.Secret = plan.Secret.ValueString()
	}
	if !plan.Comment.IsNull() {
		target.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		target.Disable = &disable
	}

	err := r.client.Notifications.CreateWebhookTarget(ctx, target)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating Webhook notification target",
			fmt.Sprintf("Could not create Webhook notification target %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *webhookNotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookNotificationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	target, err := r.client.Notifications.GetWebhookTarget(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading Webhook notification target",
			fmt.Sprintf("Could not read Webhook notification target %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}

	state.URL = types.StringValue(target.URL)
	if target.Body != "" {
		state.Body = types.StringValue(target.Body)
	}
	if target.Method != "" {
		state.Method = types.StringValue(target.Method)
	}
	if len(target.Headers) > 0 {
		headers, diags := types.MapValueFrom(ctx, types.StringType, target.Headers)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Headers = headers
	}
	// Don't update secret from API (sensitive field)
	if target.Comment != "" {
		state.Comment = types.StringValue(target.Comment)
	}
	if target.Disable != nil {
		state.Disable = types.BoolValue(*target.Disable)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *webhookNotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookNotificationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	target := &notifications.WebhookTarget{
		Name: plan.Name.ValueString(),
		URL:  plan.URL.ValueString(),
	}

	if !plan.Body.IsNull() {
		target.Body = plan.Body.ValueString()
	}
	if !plan.Method.IsNull() {
		target.Method = plan.Method.ValueString()
	}
	if !plan.Headers.IsNull() {
		headers := make(map[string]string)
		diags = plan.Headers.ElementsAs(ctx, &headers, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		target.Headers = headers
	}
	if !plan.Secret.IsNull() {
		target.Secret = plan.Secret.ValueString()
	}
	if !plan.Comment.IsNull() {
		target.Comment = plan.Comment.ValueString()
	}
	if !plan.Disable.IsNull() {
		disable := plan.Disable.ValueBool()
		target.Disable = &disable
	}

	err := r.client.Notifications.UpdateWebhookTarget(ctx, plan.Name.ValueString(), target)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating Webhook notification target",
			fmt.Sprintf("Could not update Webhook notification target %s: %s", plan.Name.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *webhookNotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookNotificationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Notifications.DeleteWebhookTarget(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting Webhook notification target",
			fmt.Sprintf("Could not delete Webhook notification target %s: %s", state.Name.ValueString(), err.Error()),
		)
		return
	}
}

// ImportState imports the resource state.
func (r *webhookNotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
