package datastores

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pbssdk "github.com/micah/terraform-provider-pbs/pbs/datastores"
)

func TestDatastoreDataSourceSchema(t *testing.T) {
	ds := &datastoreDataSource{}
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	ds.Schema(context.Background(), req, resp)

	require.False(t, resp.Diagnostics.HasError())
	require.NotNil(t, resp.Schema.Attributes)

	// Verify required name attribute
	nameAttr, ok := resp.Schema.Attributes["name"]
	require.True(t, ok, "name attribute should exist")
	require.True(t, nameAttr.IsRequired(), "name should be required")

	// Verify computed attributes exist
	computedAttrs := []string{"path", "comment", "gc_schedule", "prune_schedule",
		"keep_last", "keep_hourly", "keep_daily", "keep_weekly", "keep_monthly", "keep_yearly",
		"s3_client", "s3_bucket", "fingerprint", "digest"}

	for _, attrName := range computedAttrs {
		attr, ok := resp.Schema.Attributes[attrName]
		require.True(t, ok, "%s attribute should exist", attrName)
		require.True(t, attr.IsComputed(), "%s should be computed", attrName)
	}

	// Verify nested blocks
	require.NotNil(t, resp.Schema.Attributes["notify"])
	require.NotNil(t, resp.Schema.Attributes["maintenance_mode"])
	require.NotNil(t, resp.Schema.Attributes["tuning"])
}

func TestDatastoreToState(t *testing.T) {
	gcAtimeCutoff := 7200
	gcCacheCapacity := 256
	verifyNew := true
	keepLast := 3
	keepDaily := 7

	datastore := &pbssdk.Datastore{
		Name:          "test-datastore",
		Path:          "/mnt/datastore/test",
		Comment:       "Test datastore",
		GCSchedule:    "daily",
		PruneSchedule: "weekly",
		KeepLast:      &keepLast,
		KeepDaily:     &keepDaily,
		S3Client:      "test-endpoint",
		S3Bucket:      "test-bucket",
		Notify: &pbssdk.DatastoreNotify{
			GC:     "always",
			Prune:  "error",
			Sync:   "never",
			Verify: "warning",
		},
		MaintenanceMode: &pbssdk.MaintenanceMode{
			Type:    "read-only",
			Message: "Under maintenance",
		},
		VerifyNew: &verifyNew,
		Tuning: &pbssdk.DatastoreTuning{
			ChunkOrder:         "inode",
			GCAtimeCutoff:      &gcAtimeCutoff,
			GCAtimeSafetyCheck: &verifyNew,
			GCCacheCapacity:    &gcCacheCapacity,
			SyncLevel:          "file",
		},
		Fingerprint: "aa:bb:cc",
		Digest:      "test-digest",
	}

	var state datastoreDataSourceModel
	err := datastoreToState(datastore, &state)
	require.NoError(t, err)

	// Verify basic attributes
	assert.Equal(t, "test-datastore", state.Name.ValueString())
	assert.Equal(t, "/mnt/datastore/test", state.Path.ValueString())
	assert.Equal(t, "Test datastore", state.Comment.ValueString())
	assert.Equal(t, "daily", state.GCSchedule.ValueString())
	assert.Equal(t, "weekly", state.PruneSchedule.ValueString())
	assert.Equal(t, int64(3), state.KeepLast.ValueInt64())
	assert.Equal(t, int64(7), state.KeepDaily.ValueInt64())
	assert.Equal(t, "test-endpoint", state.S3Client.ValueString())
	assert.Equal(t, "test-bucket", state.S3Bucket.ValueString())
	assert.Equal(t, "aa:bb:cc", state.Fingerprint.ValueString())
	assert.Equal(t, "test-digest", state.Digest.ValueString())

	// Verify notify block
	require.NotNil(t, state.Notify)
	assert.Equal(t, "always", state.Notify.GC.ValueString())
	assert.Equal(t, "error", state.Notify.Prune.ValueString())
	assert.Equal(t, "never", state.Notify.Sync.ValueString())
	assert.Equal(t, "warning", state.Notify.Verify.ValueString())

	// Verify maintenance mode block
	require.NotNil(t, state.MaintenanceMode)
	assert.Equal(t, "read-only", state.MaintenanceMode.Type.ValueString())
	assert.Equal(t, "Under maintenance", state.MaintenanceMode.Message.ValueString())

	// Verify verify_new
	assert.True(t, state.VerifyNew.ValueBool())

	// Verify tuning block
	require.NotNil(t, state.Tuning)
	assert.Equal(t, "inode", state.Tuning.ChunkOrder.ValueString())
	assert.Equal(t, int64(7200), state.Tuning.GCAtimeCutoff.ValueInt64())
	assert.True(t, state.Tuning.GCAtimeSafetyCheck.ValueBool())
	assert.Equal(t, int64(256), state.Tuning.GCCacheCapacity.ValueInt64())
	assert.Equal(t, "file", state.Tuning.SyncLevel.ValueString())
}

func TestDatastoreToStateMinimal(t *testing.T) {
	datastore := &pbssdk.Datastore{
		Name: "minimal-datastore",
		Path: "/mnt/datastore/minimal",
	}

	var state datastoreDataSourceModel
	err := datastoreToState(datastore, &state)
	require.NoError(t, err)

	assert.Equal(t, "minimal-datastore", state.Name.ValueString())
	assert.Equal(t, "/mnt/datastore/minimal", state.Path.ValueString())

	// Optional fields should be null
	assert.True(t, state.Comment.IsNull())
	assert.True(t, state.GCSchedule.IsNull())
	assert.True(t, state.PruneSchedule.IsNull())
	assert.True(t, state.KeepLast.IsNull())
	assert.True(t, state.S3Client.IsNull())
	assert.True(t, state.S3Bucket.IsNull())
	assert.True(t, state.Fingerprint.IsNull())

	// Optional nested blocks should be nil
	assert.Nil(t, state.Notify)
	assert.Nil(t, state.MaintenanceMode)
	assert.Nil(t, state.Tuning)

	// VerifyNew should be false when nil
	assert.False(t, state.VerifyNew.ValueBool())
}

func TestDatastoreToStateRemovable(t *testing.T) {
	uuid := "01234567-89ab-cdef-0123-456789abcdef"
	datastore := &pbssdk.Datastore{
		Name:          "removable-ds",
		Path:          "/mnt/datastore/removable",
		Backend:       "type=removable",
		BackingDevice: uuid,
	}

	var state datastoreDataSourceModel
	err := datastoreToState(datastore, &state)
	require.NoError(t, err)

	assert.Equal(t, "removable-ds", state.Name.ValueString())
	assert.True(t, state.Removable.ValueBool())
	assert.Equal(t, uuid, state.BackingDevice.ValueString())
}
