package datastores

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/require"
)

func TestDatastoresDataSourceSchema(t *testing.T) {
	ds := &datastoresDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.Schema.Attributes)

	// Verify stores attribute exists and is computed
	storesAttr, ok := resp.Schema.Attributes["stores"]
	require.True(t, ok, "stores attribute should exist")
	require.True(t, storesAttr.IsComputed(), "stores should be computed")
}
