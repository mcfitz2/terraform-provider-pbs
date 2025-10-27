/*
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/.
 */

package remotes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/micah/terraform-provider-pbs/fwprovider/config"
	"github.com/micah/terraform-provider-pbs/pbs"
)

var (
	_ datasource.DataSource              = &remoteGroupsDataSource{}
	_ datasource.DataSourceWithConfigure = &remoteGroupsDataSource{}
)

// NewRemoteGroupsDataSource is a helper function to simplify the provider implementation.
func NewRemoteGroupsDataSource() datasource.DataSource {
	return &remoteGroupsDataSource{}
}

// remoteGroupsDataSource is the data source implementation.
type remoteGroupsDataSource struct {
	client *pbs.Client
}

// remoteGroupsDataSourceModel maps the data source schema data.
type remoteGroupsDataSourceModel struct {
	RemoteName types.String   `tfsdk:"remote_name"`
	Store      types.String   `tfsdk:"store"`
	Namespace  types.String   `tfsdk:"namespace"`
	Groups     []types.String `tfsdk:"groups"`
}

// Metadata returns the data source type name.
func (d *remoteGroupsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_groups"
}

// Schema defines the schema for the data source.
func (d *remoteGroupsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available backup groups in a namespace on a remote PBS server.",
		MarkdownDescription: `Lists available backup groups in a namespace on a remote PBS server.

This data source scans the specified remote server's datastore namespace and returns a list of backup group identifiers. 
Useful for configuring sync jobs with specific group filters or discovering the backup structure on a remote PBS instance.`,
		Attributes: map[string]schema.Attribute{
			"remote_name": schema.StringAttribute{
				Description:         "The name of the configured remote to scan.",
				MarkdownDescription: "The name of the configured remote to scan.",
				Required:            true,
			},
			"store": schema.StringAttribute{
				Description:         "The name of the datastore on the remote server to scan.",
				MarkdownDescription: "The name of the datastore on the remote server to scan.",
				Required:            true,
			},
			"namespace": schema.StringAttribute{
				Description:         "The namespace path in the datastore to scan (optional, defaults to root namespace).",
				MarkdownDescription: "The namespace path in the datastore to scan (optional, defaults to root namespace).",
				Optional:            true,
			},
			"groups": schema.ListAttribute{
				Description:         "List of backup group identifiers available in the remote namespace.",
				MarkdownDescription: "List of backup group identifiers available in the remote namespace.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *remoteGroupsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	cfg, ok := req.ProviderData.(*config.Resource)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *config.Resource, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = cfg.Client
}

// Read refreshes the Terraform state with the latest data.
func (d *remoteGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state remoteGroupsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace := ""
	if !state.Namespace.IsNull() && !state.Namespace.IsUnknown() {
		namespace = state.Namespace.ValueString()
	}

	groups, err := d.client.Remotes.ListRemoteGroups(
		ctx,
		state.RemoteName.ValueString(),
		state.Store.ValueString(),
		namespace,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error scanning remote groups",
			fmt.Sprintf("Could not scan groups on remote %s, store %s, namespace %s: %s",
				state.RemoteName.ValueString(),
				state.Store.ValueString(),
				namespace,
				err.Error()),
		)
		return
	}

	// Extract group identifiers (format: "type/id") and convert to []types.String
	groupsList := make([]types.String, 0, len(groups))
	for _, group := range groups {
		groupID := fmt.Sprintf("%s/%s", group.BackupType, group.BackupID)
		groupsList = append(groupsList, types.StringValue(groupID))
	}
	state.Groups = groupsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
