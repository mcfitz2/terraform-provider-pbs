package datastores

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	pbssdk "github.com/micah/terraform-provider-pbs/pbs/datastores"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanToDatastoreAdvancedFields(t *testing.T) {
	resource := &datastoreResource{}

	plan := datastoreResourceModel{
		Name:             types.StringValue("test-datastore"),
		Type:             types.StringValue("dir"),
		Path:             types.StringValue("/datastore/test"),
		NotifyUser:       types.StringValue("root@pam"),
		NotifyLevel:      types.StringValue("warning"),
		NotificationMode: types.StringValue("notification-system"),
		Notify: &notifyModel{
			GC:     types.StringValue("Always"),
			Prune:  types.StringValue("ERROR"),
			Sync:   types.StringValue("never"),
			Verify: types.StringValue("Always"),
		},
		MaintenanceMode: &maintenanceModeModel{
			Type:    types.StringValue("OFFLINE"),
			Message: types.StringValue("Planned maintenance"),
		},
		VerifyNew:      types.BoolValue(true),
		ReuseDatastore: types.BoolValue(true),
		OverwriteInUse: types.BoolValue(false),
		Tuning: &tuningModel{
			ChunkOrder:         types.StringValue("inode"),
			GCAtimeCutoff:      types.Int64Value(3600),
			GCAtimeSafetyCheck: types.BoolValue(true),
			GCCacheCapacity:    types.Int64Value(512),
			SyncLevel:          types.StringValue("filesystem"),
		},
		TuneLevel:   types.Int64Value(2),
		Fingerprint: types.StringValue("aa:bb"),
		Digest:      types.StringValue("12345"),
	}

	ds, err := resource.planToDatastore(&plan, nil)
	require.NoError(t, err)

	require.NotNil(t, ds.Tuning)
	assert.Equal(t, "inode", ds.Tuning.ChunkOrder)
	require.NotNil(t, ds.Tuning.GCAtimeCutoff)
	assert.Equal(t, 3600, *ds.Tuning.GCAtimeCutoff)
	require.NotNil(t, ds.Tuning.GCAtimeSafetyCheck)
	assert.True(t, *ds.Tuning.GCAtimeSafetyCheck)
	require.NotNil(t, ds.Tuning.GCCacheCapacity)
	assert.Equal(t, 512, *ds.Tuning.GCCacheCapacity)
	assert.Equal(t, "file", ds.Tuning.SyncLevel, "tune_level should override sync_level with PBS value")

	require.NotNil(t, ds.Notify)
	assert.Equal(t, "always", ds.Notify.GC)
	assert.Equal(t, "error", ds.Notify.Prune)
	assert.Equal(t, "never", ds.Notify.Sync)
	assert.Equal(t, "always", ds.Notify.Verify)

	require.NotNil(t, ds.MaintenanceMode)
	assert.Equal(t, "offline", ds.MaintenanceMode.Type)
	assert.Equal(t, "Planned maintenance", ds.MaintenanceMode.Message)

	require.NotNil(t, ds.VerifyNew)
	assert.True(t, *ds.VerifyNew)
	require.NotNil(t, ds.ReuseDatastore)
	assert.True(t, *ds.ReuseDatastore)

	require.NotNil(t, ds.OverwriteInUse)
	assert.False(t, *ds.OverwriteInUse)

	assert.Equal(t, "aa:bb", ds.Fingerprint)
	assert.Equal(t, "12345", ds.Digest)
	assert.Nil(t, ds.Delete, "no deletes should be requested when plan supplies values")
}

func TestPlanToDatastoreDeleteSections(t *testing.T) {
	resource := &datastoreResource{}

	plan := datastoreResourceModel{
		Name:             types.StringValue("test"),
		Type:             types.StringValue("dir"),
		Path:             types.StringValue("/datastore/test"),
		NotifyUser:       types.StringNull(),
		NotifyLevel:      types.StringNull(),
		NotificationMode: types.StringNull(),
		VerifyNew:        types.BoolNull(),
		ReuseDatastore:   types.BoolNull(),
		OverwriteInUse:   types.BoolNull(),
		TuneLevel:        types.Int64Null(),
	}

	state := datastoreResourceModel{
		NotifyUser:       types.StringValue("root@pam"),
		NotifyLevel:      types.StringValue("warning"),
		NotificationMode: types.StringValue("legacy-sendmail"),
		VerifyNew:        types.BoolValue(true),
		ReuseDatastore:   types.BoolValue(true),
		OverwriteInUse:   types.BoolValue(true),
		Notify: &notifyModel{
			GC: types.StringValue("always"),
		},
		MaintenanceMode: &maintenanceModeModel{
			Type: types.StringValue("offline"),
		},
		Tuning: &tuningModel{
			SyncLevel: types.StringValue("filesystem"),
		},
	}

	ds, err := resource.planToDatastore(&plan, &state)
	require.NoError(t, err)

	require.NotNil(t, ds.Delete)
	assert.ElementsMatch(t, []string{
		"notify-user",
		"notify-level",
		"notification-mode",
		"notify",
		"maintenance-mode",
		"tuning",
		"verify-new",
		"reuse-datastore",
		"overwrite-in-use",
	}, ds.Delete)
}

func TestDatastoreToStateRoundTrip(t *testing.T) {
	resource := &datastoreResource{}

	maxBackups := 5
	gcAtimeCutoff := 7200
	gcCacheCapacity := 256
	verifyNew := true
	reuse := true
	overwrite := false

	ds := &pbssdk.Datastore{
		Name:          "example",
		Type:          pbssdk.DatastoreTypeDirectory,
		Path:          "/datastore/example",
		Content:       []string{"backup", "iso"},
		MaxBackups:    &maxBackups,
		Comment:       "round-trip",
		Disabled:      boolPtr(false),
		GCSchedule:    "daily",
		PruneSchedule: "weekly",
		KeepDaily:     intPtr(3),
		Notify: &pbssdk.DatastoreNotify{
			GC:     "always",
			Verify: "error",
		},
		MaintenanceMode: &pbssdk.MaintenanceMode{
			Type:    "offline",
			Message: "maintenance",
		},
		VerifyNew:      &verifyNew,
		ReuseDatastore: &reuse,
		OverwriteInUse: &overwrite,
		Tuning: &pbssdk.DatastoreTuning{
			ChunkOrder:         "inode",
			GCAtimeCutoff:      &gcAtimeCutoff,
			GCAtimeSafetyCheck: &verifyNew,
			GCCacheCapacity:    &gcCacheCapacity,
			SyncLevel:          "file",
		},
		Fingerprint: "ff:ee",
		Digest:      "digest-value",
	}

	state := datastoreResourceModel{
		Password: types.StringValue("keepme"),
	}

	err := resource.datastoreToState(ds, &state)
	require.NoError(t, err)

	assert.Equal(t, types.StringValue("example"), state.Name)
	assert.Equal(t, types.StringValue("dir"), state.Type)
	assert.Equal(t, types.StringValue("/datastore/example"), state.Path)
	assert.Equal(t, "round-trip", state.Comment.ValueString())
	assert.Equal(t, "daily", state.GCSchedule.ValueString())
	assert.Equal(t, "weekly", state.PruneSchedule.ValueString())
	assert.Equal(t, int64(3), state.KeepDaily.ValueInt64())
	assert.Equal(t, "ff:ee", state.Fingerprint.ValueString())
	assert.Equal(t, "digest-value", state.Digest.ValueString())

	// Content list should be populated with both entries
	require.False(t, state.Content.IsNull())
	elems := state.Content.Elements()
	require.Len(t, elems, 2)
	assert.Equal(t, "backup", elems[0].(types.String).ValueString())
	assert.Equal(t, "iso", elems[1].(types.String).ValueString())

	// Notify block populated
	require.NotNil(t, state.Notify)
	assert.Equal(t, "always", state.Notify.GC.ValueString())
	assert.Equal(t, "error", state.Notify.Verify.ValueString())

	// Maintenance mode
	require.NotNil(t, state.MaintenanceMode)
	assert.Equal(t, "offline", state.MaintenanceMode.Type.ValueString())
	assert.Equal(t, "maintenance", state.MaintenanceMode.Message.ValueString())

	// Tuning block and tune level derived from sync level
	require.NotNil(t, state.Tuning)
	assert.Equal(t, "inode", state.Tuning.ChunkOrder.ValueString())
	assert.Equal(t, int64(gcAtimeCutoff), state.Tuning.GCAtimeCutoff.ValueInt64())
	assert.True(t, state.Tuning.GCAtimeSafetyCheck.ValueBool())
	assert.Equal(t, int64(gcCacheCapacity), state.Tuning.GCCacheCapacity.ValueInt64())
	assert.Equal(t, "file", state.Tuning.SyncLevel.ValueString())
	assert.Equal(t, int64(2), state.TuneLevel.ValueInt64())

	// Booleans preserved
	assert.True(t, state.VerifyNew.ValueBool())
	assert.True(t, state.ReuseDatastore.ValueBool())
	assert.False(t, state.OverwriteInUse.ValueBool())

	// Password should remain untouched (not overwritten by API payload)
	assert.Equal(t, "keepme", state.Password.ValueString())
}

func boolPtr(v bool) *bool {
	return &v
}

func intPtr(v int) *int {
	return &v
}
