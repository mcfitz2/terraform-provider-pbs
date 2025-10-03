/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

// Package resources provides Terraform resources for the PBS provider
package resources

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
	"github.com/micah/terraform-provider-pbs/pbs/endpoints"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &s3EndpointResource{}
	_ resource.ResourceWithConfigure   = &s3EndpointResource{}
	_ resource.ResourceWithImportState = &s3EndpointResource{}
)

// NewS3EndpointResource is a helper function to simplify the provider implementation.
func NewS3EndpointResource() resource.Resource {
	return &s3EndpointResource{}
}

// s3EndpointResource is the resource implementation.
type s3EndpointResource struct {
	client *pbs.Client
}

// s3EndpointResourceModel maps the resource schema data.
type s3EndpointResourceModel struct {
	ID             types.String `tfsdk:"id"`
	AccessKey      types.String `tfsdk:"access_key"`
	SecretKey      types.String `tfsdk:"secret_key"`
	Endpoint       types.String `tfsdk:"endpoint"`
	Region         types.String `tfsdk:"region"`
	Fingerprint    types.String `tfsdk:"fingerprint"`
	Port           types.Int64  `tfsdk:"port"`
	PathStyle      types.Bool   `tfsdk:"path_style"`
	ProviderQuirks types.Set    `tfsdk:"provider_quirks"`
}

// Metadata returns the resource type name.
func (r *s3EndpointResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_s3_endpoint"
}

// Schema defines the schema for the resource.
func (r *s3EndpointResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		Description: "Manages a Proxmox Backup Server S3 client configuration.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "ID to uniquely identify S3 client config (3-32 chars, alphanumeric with dots, dashes, underscores).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`), "must match pattern [A-Za-z0-9_][A-Za-z0-9._-]*"),
				},
			},
			"access_key": schema.StringAttribute{
				Description: "Access key for S3 object store.",
				Required:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"secret_key": schema.StringAttribute{
				Description: "S3 secret key.",
				Required:    true,
				Sensitive:   true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"endpoint": schema.StringAttribute{
				Description: "Endpoint to access S3 object store.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "Region to access S3 object store (3-32 chars, lowercase alphanumeric with dashes/underscores).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
					stringvalidator.RegexMatches(regexp.MustCompile(`^[_a-z\d][-_a-z\d]+$`), "must match pattern [_a-z\\d][-_a-z\\d]+"),
				},
			},
			"fingerprint": schema.StringAttribute{
				Description: "X509 certificate fingerprint (sha256).",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(regexp.MustCompile(`^[0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){31}$`), "must be a valid SHA256 fingerprint"),
				},
			},
			"port": schema.Int64Attribute{
				Description: "Port to access S3 object store.",
				Optional:    true,
			},
			"path_style": schema.BoolAttribute{
				Description: "Use path style bucket addressing over vhost style.",
				Optional:    true,
			},
			"provider_quirks": schema.SetAttribute{
				Description:         "S3 provider-specific quirks. Use ['skip-if-none-match-header'] for Backblaze B2 compatibility.",
				MarkdownDescription: "S3 provider-specific quirks. Use `['skip-if-none-match-header']` for Backblaze B2 compatibility to handle unsupported S3 headers.",
				Optional:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *s3EndpointResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *config.Resource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = config.Client
}

// Create creates the resource and sets the initial Terraform state.
func (r *s3EndpointResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan s3EndpointResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create the S3 endpoint
	endpoint := &endpoints.S3Endpoint{
		ID:        plan.ID.ValueString(),
		AccessKey: plan.AccessKey.ValueString(),
		SecretKey: plan.SecretKey.ValueString(),
		Endpoint:  plan.Endpoint.ValueString(),
	}

	// Set optional fields if provided
	if !plan.Region.IsNull() && !plan.Region.IsUnknown() {
		endpoint.Region = plan.Region.ValueString()
	}

	if !plan.Fingerprint.IsNull() && !plan.Fingerprint.IsUnknown() {
		endpoint.Fingerprint = plan.Fingerprint.ValueString()
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := int(plan.Port.ValueInt64())
		endpoint.Port = &port
	}

	if !plan.PathStyle.IsNull() && !plan.PathStyle.IsUnknown() {
		pathStyle := plan.PathStyle.ValueBool()
		endpoint.PathStyle = &pathStyle
	}

	// Handle provider_quirks if provided
	if !plan.ProviderQuirks.IsNull() && !plan.ProviderQuirks.IsUnknown() {
		var quirks []string
		resp.Diagnostics.Append(plan.ProviderQuirks.ElementsAs(ctx, &quirks, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		endpoint.ProviderQuirks = quirks
	}

	err := r.client.Endpoints.CreateS3Endpoint(ctx, endpoint)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating S3 Endpoint",
			"Could not create S3 endpoint, unexpected error: "+err.Error(),
		)
		return
	}

	// The ID is already set from the plan (it's required)

	// Log that the resource was created
	tflog.Trace(ctx, "created S3 endpoint resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *s3EndpointResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state s3EndpointResourceModel

	// Get current state
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed values from API
	endpoint, err := r.client.Endpoints.GetS3Endpoint(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading S3 Endpoint",
			"Could not read S3 endpoint ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(endpoint.ID)
	state.AccessKey = types.StringValue(endpoint.AccessKey)
	state.Endpoint = types.StringValue(endpoint.Endpoint)

	// Set optional fields
	if endpoint.Region != "" {
		state.Region = types.StringValue(endpoint.Region)
	} else {
		state.Region = types.StringNull()
	}

	if endpoint.Fingerprint != "" {
		state.Fingerprint = types.StringValue(endpoint.Fingerprint)
	} else {
		state.Fingerprint = types.StringNull()
	}

	if endpoint.Port != nil {
		state.Port = types.Int64Value(int64(*endpoint.Port))
	} else {
		state.Port = types.Int64Null()
	}

	if endpoint.PathStyle != nil {
		state.PathStyle = types.BoolValue(*endpoint.PathStyle)
	} else {
		state.PathStyle = types.BoolNull()
	}

	// Handle provider_quirks
	if len(endpoint.ProviderQuirks) > 0 {
		var quirks []attr.Value
		for _, quirk := range endpoint.ProviderQuirks {
			quirks = append(quirks, types.StringValue(quirk))
		}
		quirksSet, diags := types.SetValue(types.StringType, quirks)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.ProviderQuirks = quirksSet
	} else {
		state.ProviderQuirks = types.SetNull(types.StringType)
	}

	// Note: SecretKey is not returned by the API for security reasons
	// so we keep the existing value in the state

	// Set refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *s3EndpointResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan s3EndpointResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the S3 endpoint
	endpoint := &endpoints.S3Endpoint{
		ID:        plan.ID.ValueString(),
		AccessKey: plan.AccessKey.ValueString(),
		SecretKey: plan.SecretKey.ValueString(),
		Endpoint:  plan.Endpoint.ValueString(),
	}

	// Set optional fields if provided
	if !plan.Region.IsNull() && !plan.Region.IsUnknown() {
		endpoint.Region = plan.Region.ValueString()
	}

	if !plan.Fingerprint.IsNull() && !plan.Fingerprint.IsUnknown() {
		endpoint.Fingerprint = plan.Fingerprint.ValueString()
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := int(plan.Port.ValueInt64())
		endpoint.Port = &port
	}

	if !plan.PathStyle.IsNull() && !plan.PathStyle.IsUnknown() {
		pathStyle := plan.PathStyle.ValueBool()
		endpoint.PathStyle = &pathStyle
	}

	// Handle provider_quirks if provided
	if !plan.ProviderQuirks.IsNull() && !plan.ProviderQuirks.IsUnknown() {
		var quirks []string
		resp.Diagnostics.Append(plan.ProviderQuirks.ElementsAs(ctx, &quirks, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		endpoint.ProviderQuirks = quirks
	}

	err := r.client.Endpoints.UpdateS3Endpoint(ctx, plan.ID.ValueString(), endpoint)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating S3 Endpoint",
			"Could not update S3 endpoint, unexpected error: "+err.Error(),
		)
		return
	}

	// Log that the resource was updated
	tflog.Trace(ctx, "updated S3 endpoint resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *s3EndpointResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state s3EndpointResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete existing endpoint
	err := r.client.Endpoints.DeleteS3Endpoint(ctx, state.ID.ValueString())
	if err != nil {
		// Check if the endpoint is already gone (desired state achieved)
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "not found") ||
			strings.Contains(errorMsg, "does not exist") ||
			strings.Contains(errorMsg, "no such s3-endpoint") ||
			strings.Contains(errorMsg, "404") {
			// Resource already deleted - this is fine, desired state achieved
			tflog.Info(ctx, "S3 endpoint already deleted", map[string]any{"id": state.ID.ValueString()})
			return
		}

		resp.Diagnostics.AddError(
			"Error Deleting S3 Endpoint",
			"Could not delete S3 endpoint, unexpected error: "+err.Error(),
		)
		return
	}

	// Log that the resource was deleted
	tflog.Trace(ctx, "deleted S3 endpoint resource")
}

// ImportState imports an existing resource into Terraform state.
func (r *s3EndpointResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
