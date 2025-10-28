package jobs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/require"
)

func TestSyncJobDataSourceSchema(t *testing.T) {
	ds := &syncJobDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.Schema.Attributes)

	// Verify required id attribute
	idAttr, ok := resp.Schema.Attributes["id"]
	require.True(t, ok, "id attribute should exist")
	require.True(t, idAttr.IsRequired(), "id should be required")

	// Verify computed attributes
	computedAttrs := []string{"store", "remote", "remote_store", "schedule", 
		"remote_namespace", "namespace", "comment", "remove_vanished", 
		"max_depth", "group_filter", "digest"}

	for _, attrName := range computedAttrs {
		attr, ok := resp.Schema.Attributes[attrName]
		require.True(t, ok, "%s attribute should exist", attrName)
		require.True(t, attr.IsComputed(), "%s should be computed", attrName)
	}
}
