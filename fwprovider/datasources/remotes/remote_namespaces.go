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
	_ datasource.DataSource              = &remoteNamespacesDataSource{}
	_ datasource.DataSourceWithConfigure = &remoteNamespacesDataSource{}
)

// NewRemoteNamespacesDataSource is a helper function to simplify the provider implementation.
func NewRemoteNamespacesDataSource() datasource.DataSource {
	return &remoteNamespacesDataSource{}
}

// remoteNamespacesDataSource is the data source implementation.
type remoteNamespacesDataSource struct {
	client *pbs.Client
}

// remoteNamespacesDataSourceModel maps the data source schema data.
type remoteNamespacesDataSourceModel struct {
	RemoteName types.String   `tfsdk:"remote_name"`
	Store      types.String   `tfsdk:"store"`
	Namespaces []types.String `tfsdk:"namespaces"`
}

// Metadata returns the data source type name.
func (d *remoteNamespacesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_remote_namespaces"
}

// Schema defines the schema for the data source.
func (d *remoteNamespacesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available namespaces in a datastore on a remote PBS server.",
		MarkdownDescription: `Lists available namespaces in a datastore on a remote PBS server.

This data source scans the specified remote server's datastore and returns a list of namespace paths. 
Useful for configuring sync jobs with specific namespace filters or discovering the namespace structure on a remote PBS instance.`,
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
			"namespaces": schema.ListAttribute{
				Description:         "List of namespace paths available in the remote datastore.",
				MarkdownDescription: "List of namespace paths available in the remote datastore.",
				ElementType:         types.StringType,
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *remoteNamespacesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *remoteNamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state remoteNamespacesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespaces, err := d.client.Remotes.ListRemoteNamespaces(
		ctx,
		state.RemoteName.ValueString(),
		state.Store.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error scanning remote namespaces",
			fmt.Sprintf("Could not scan namespaces on remote %s, store %s: %s",
				state.RemoteName.ValueString(),
				state.Store.ValueString(),
				err.Error()),
		)
		return
	}

	// Extract namespace paths and convert to []types.String
	namespacesList := make([]types.String, 0, len(namespaces))
	for _, ns := range namespaces {
		namespacesList = append(namespacesList, types.StringValue(ns.Namespace))
	}
	state.Namespaces = namespacesList

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
