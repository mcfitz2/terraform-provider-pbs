package jobs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbsjobs "github.com/micah/terraform-provider-pbs/pbs/jobs"
)

func TestPruneJobDataSourceSchema(t *testing.T) {
	ds := &pruneJobDataSource{}
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
	computedAttrs := []string{"store", "schedule", "keep_last", "keep_hourly",
		"keep_daily", "keep_weekly", "keep_monthly", "keep_yearly",
		"max_depth", "namespace", "comment", "disable", "digest"}

	for _, attrName := range computedAttrs {
		attr, ok := resp.Schema.Attributes[attrName]
		require.True(t, ok, "%s attribute should exist", attrName)
		require.True(t, attr.IsComputed(), "%s should be computed", attrName)
	}
}

func TestPruneJobToState(t *testing.T) {
	keepLast := 5
	keepDaily := 7
	maxDepth := 2
	disabled := false

	job := &pbsjobs.PruneJob{
		ID:        "prune-job-1",
		Store:     "datastore1",
		Schedule:  "daily",
		KeepLast:  &keepLast,
		KeepDaily: &keepDaily,
		MaxDepth:  &maxDepth,
		Namespace: "ns1",
		Comment:   "Test prune job",
		Disable:   &disabled,
		Digest:    "test-digest",
	}

	var state pruneJobDataSourceModel
	pruneJobToState(job, &state)

	assert.Equal(t, "prune-job-1", state.ID.ValueString())
	assert.Equal(t, "datastore1", state.Store.ValueString())
	assert.Equal(t, "daily", state.Schedule.ValueString())
	assert.Equal(t, int64(5), state.KeepLast.ValueInt64())
	assert.Equal(t, int64(7), state.KeepDaily.ValueInt64())
	assert.Equal(t, int64(2), state.MaxDepth.ValueInt64())
	assert.Equal(t, "ns1", state.Namespace.ValueString())
	assert.Equal(t, "Test prune job", state.Comment.ValueString())
	assert.False(t, state.Disable.ValueBool())
	assert.Equal(t, "test-digest", state.Digest.ValueString())
}

func TestPruneJobToStateMinimal(t *testing.T) {
	job := &pbsjobs.PruneJob{
		ID:       "prune-job-minimal",
		Store:    "datastore1",
		Schedule: "weekly",
	}

	var state pruneJobDataSourceModel
	pruneJobToState(job, &state)

	assert.Equal(t, "prune-job-minimal", state.ID.ValueString())
	assert.Equal(t, "datastore1", state.Store.ValueString())
	assert.Equal(t, "weekly", state.Schedule.ValueString())

	// Optional fields should be null
	assert.True(t, state.KeepLast.IsNull())
	assert.True(t, state.KeepDaily.IsNull())
	assert.True(t, state.MaxDepth.IsNull())
	assert.True(t, state.Namespace.IsNull())
	assert.True(t, state.Comment.IsNull())
	assert.True(t, state.Disable.IsNull())

	// Digest will be empty string (not null) when omitted
	assert.Equal(t, "", state.Digest.ValueString())
}
