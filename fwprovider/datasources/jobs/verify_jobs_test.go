package jobs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/require"
)

func TestVerifyJobsDataSourceSchema(t *testing.T) {
	ds := &verifyJobsDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.Schema.Attributes)

	// Verify jobs attribute exists and is computed
	jobsAttr, ok := resp.Schema.Attributes["jobs"]
	require.True(t, ok, "jobs attribute should exist")
	require.True(t, jobsAttr.IsComputed(), "jobs should be computed")

	// Verify optional store filter attribute
	storeAttr, ok := resp.Schema.Attributes["store"]
	require.True(t, ok, "store attribute should exist")
	require.True(t, storeAttr.IsOptional(), "store should be optional")
}
