package jobs

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbsjobs "github.com/micah/terraform-provider-pbs/pbs/jobs"
)

func TestVerifyJobDataSourceSchema(t *testing.T) {
	ds := &verifyJobDataSource{}
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
	computedAttrs := []string{"store", "schedule", "max_depth", 
		"namespace", "outdated_after", "comment", "digest"}

	for _, attrName := range computedAttrs {
		attr, ok := resp.Schema.Attributes[attrName]
		require.True(t, ok, "%s attribute should exist", attrName)
		require.True(t, attr.IsComputed(), "%s should be computed", attrName)
	}
}

func TestVerifyJobToState(t *testing.T) {
	maxDepth := 2
	outdatedAfter := 30

	job := &pbsjobs.VerifyJob{
		ID:            "verify-job-1",
		Store:         "datastore1",
		Schedule:      "weekly",
		MaxDepth:      &maxDepth,
		Namespace:     "production",
		OutdatedAfter: &outdatedAfter,
		Comment:       "Test verify job",
		Digest:        "test-digest",
	}

	var state verifyJobDataSourceModel
	verifyJobToState(job, &state)

	assert.Equal(t, "verify-job-1", state.ID.ValueString())
	assert.Equal(t, "datastore1", state.Store.ValueString())
	assert.Equal(t, "weekly", state.Schedule.ValueString())
	assert.Equal(t, int64(2), state.MaxDepth.ValueInt64())
	assert.Equal(t, "production", state.Namespace.ValueString())
	assert.Equal(t, int64(30), state.OutdatedAfter.ValueInt64())
	assert.Equal(t, "Test verify job", state.Comment.ValueString())
	assert.Equal(t, "test-digest", state.Digest.ValueString())
}

func TestVerifyJobToStateMinimal(t *testing.T) {
	job := &pbsjobs.VerifyJob{
		ID:       "verify-job-minimal",
		Store:    "datastore1",
		Schedule: "monthly",
	}

	var state verifyJobDataSourceModel
	verifyJobToState(job, &state)

	assert.Equal(t, "verify-job-minimal", state.ID.ValueString())
	assert.Equal(t, "datastore1", state.Store.ValueString())
	assert.Equal(t, "monthly", state.Schedule.ValueString())

	// Optional fields should be null
	assert.True(t, state.MaxDepth.IsNull())
	assert.True(t, state.Namespace.IsNull())
	assert.True(t, state.OutdatedAfter.IsNull())
	assert.True(t, state.Comment.IsNull())
	
	// Digest will be empty string (not null) when omitted
	assert.Equal(t, "", state.Digest.ValueString())
}
