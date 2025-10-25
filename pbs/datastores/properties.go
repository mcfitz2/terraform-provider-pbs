package datastores

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// MaintenanceMode represents the datastore maintenance-mode property string
// with a typed representation that is easier to work with in Go and Terraform.
type MaintenanceMode struct {
	Type    string
	Message string
}

// DatastoreNotify represents the per-job notification settings for a datastore.
type DatastoreNotify struct {
	GC     string
	Prune  string
	Sync   string
	Verify string
}

// DatastoreTuning captures the advanced tuning options supported by PBS 4.0.
type DatastoreTuning struct {
	ChunkOrder         string
	GCAtimeCutoff      *int
	GCAtimeSafetyCheck *bool
	GCCacheCapacity    *int
	SyncLevel          string
}

// formatMaintenanceMode converts a MaintenanceMode struct into the PBS property string format.
func formatMaintenanceMode(mm *MaintenanceMode) string {
	if mm == nil || mm.Type == "" {
		return ""
	}

	parts := []string{fmt.Sprintf("type=%s", mm.Type)}
	if mm.Message != "" {
		parts = append(parts, fmt.Sprintf("message=\"%s\"", escapeQuotes(mm.Message)))
	}

	return strings.Join(parts, ",")
}

// parseMaintenanceMode parses the PBS maintenance-mode property string into a typed struct.
func parseMaintenanceMode(raw string) *MaintenanceMode {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	props := parsePropertyString(raw)
	if len(props) == 0 {
		return nil
	}

	mm := &MaintenanceMode{Type: props["type"]}
	if msg, ok := props["message"]; ok {
		mm.Message = msg
	}

	if mm.Type == "" && mm.Message == "" {
		return nil
	}

	return mm
}

// formatNotify converts the notification settings into a property string understood by PBS.
func formatNotify(notify *DatastoreNotify) string {
	if notify == nil {
		return ""
	}

	entries := make(map[string]string)
	if notify.GC != "" {
		entries["gc"] = notify.GC
	}
	if notify.Prune != "" {
		entries["prune"] = notify.Prune
	}
	if notify.Sync != "" {
		entries["sync"] = notify.Sync
	}
	if notify.Verify != "" {
		entries["verify"] = notify.Verify
	}

	return formatPropertyString(entries)
}

// parseNotify transforms the PBS property string representation into a DatastoreNotify struct.
func parseNotify(raw string) *DatastoreNotify {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	props := parsePropertyString(raw)
	if len(props) == 0 {
		return nil
	}

	notify := &DatastoreNotify{
		GC:     props["gc"],
		Prune:  props["prune"],
		Sync:   props["sync"],
		Verify: props["verify"],
	}

	if notify.GC == "" && notify.Prune == "" && notify.Sync == "" && notify.Verify == "" {
		return nil
	}

	return notify
}

// formatTuning converts the tuning struct into the PBS property string representation.
func formatTuning(t *DatastoreTuning) string {
	if t == nil {
		return ""
	}

	entries := make(map[string]string)
	if t.ChunkOrder != "" {
		entries["chunk-order"] = t.ChunkOrder
	}
	if t.GCAtimeCutoff != nil {
		entries["gc-atime-cutoff"] = strconv.Itoa(*t.GCAtimeCutoff)
	}
	if t.GCAtimeSafetyCheck != nil {
		if *t.GCAtimeSafetyCheck {
			entries["gc-atime-safety-check"] = "1"
		} else {
			entries["gc-atime-safety-check"] = "0"
		}
	}
	if t.GCCacheCapacity != nil {
		entries["gc-cache-capacity"] = strconv.Itoa(*t.GCCacheCapacity)
	}
	if t.SyncLevel != "" {
		entries["sync-level"] = t.SyncLevel
	}

	return formatPropertyString(entries)
}

// parseTuning parses a tuning property string into a DatastoreTuning struct.
func parseTuning(raw string) *DatastoreTuning {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	props := parsePropertyString(raw)
	if len(props) == 0 {
		return nil
	}

	tuning := &DatastoreTuning{}

	if v, ok := props["chunk-order"]; ok {
		tuning.ChunkOrder = v
	}

	if v, ok := props["gc-atime-cutoff"]; ok {
		if iv, err := strconv.Atoi(v); err == nil {
			tuning.GCAtimeCutoff = &iv
		}
	}

	if v, ok := props["gc-atime-safety-check"]; ok {
		bv := false
		switch strings.ToLower(v) {
		case "1", "true":
			bv = true
		case "0", "false":
			bv = false
		default:
			if parsed, err := strconv.ParseBool(v); err == nil {
				bv = parsed
			}
		}
		tuning.GCAtimeSafetyCheck = &bv
	}

	if v, ok := props["gc-cache-capacity"]; ok {
		if iv, err := strconv.Atoi(v); err == nil {
			tuning.GCCacheCapacity = &iv
		}
	}

	if v, ok := props["sync-level"]; ok {
		tuning.SyncLevel = v
	}

	if tuning.ChunkOrder == "" && tuning.GCAtimeCutoff == nil && tuning.GCAtimeSafetyCheck == nil && tuning.GCCacheCapacity == nil && tuning.SyncLevel == "" {
		return nil
	}

	return tuning
}

// parsePropertyString converts a PBS "property string" (key=value pairs separated by commas)
// into a map. It understands quoted values and simple escaping.
func parsePropertyString(raw string) map[string]string {
	result := make(map[string]string)
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}

	var (
		keyBuilder   strings.Builder
		valueBuilder strings.Builder
		readingKey   = true
		inQuotes     = false
		escaped      = false
	)

	flush := func() {
		key := strings.TrimSpace(keyBuilder.String())
		value := strings.TrimSpace(valueBuilder.String())
		if key != "" {
			result[strings.ToLower(key)] = value
		}
		keyBuilder.Reset()
		valueBuilder.Reset()
		readingKey = true
	}

	for _, r := range raw {
		switch {
		case escaped:
			if readingKey {
				keyBuilder.WriteRune(r)
			} else {
				valueBuilder.WriteRune(r)
			}
			escaped = false

		case r == '\\' && inQuotes:
			escaped = true

		case r == '"':
			inQuotes = !inQuotes

		case !inQuotes && r == '=' && readingKey:
			readingKey = false

		case !inQuotes && r == ',':
			flush()

		default:
			if readingKey {
				keyBuilder.WriteRune(r)
			} else {
				valueBuilder.WriteRune(r)
			}
		}
	}

	flush()

	// Remove surrounding quotes from values.
	for k, v := range result {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") && len(v) >= 2 {
			v = v[1 : len(v)-1]
			v = strings.ReplaceAll(v, "\\\"", "\"")
		}
		result[k] = v
	}

	return result
}

// formatPropertyString builds a property string from the given map, producing a stable ordering.
func formatPropertyString(entries map[string]string) string {
	if len(entries) == 0 {
		return ""
	}

	keys := make([]string, 0, len(entries))
	for k, v := range entries {
		if v != "" {
			keys = append(keys, k)
		}
	}

	if len(keys) == 0 {
		return ""
	}

	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		value := entries[k]
		if strings.ContainsAny(value, ",\" ") {
			value = fmt.Sprintf("\"%s\"", escapeQuotes(value))
		}
		parts = append(parts, fmt.Sprintf("%s=%s", k, value))
	}

	return strings.Join(parts, ",")
}

// FormatBackendString builds a backend configuration string with the provided type and key/value pairs.
func FormatBackendString(backendType string, params map[string]string) string {
	entries := make(map[string]string, len(params)+1)
	entries["type"] = strings.ToLower(strings.TrimSpace(backendType))
	for k, v := range params {
		value := strings.TrimSpace(v)
		if value == "" {
			continue
		}
		entries[strings.ToLower(strings.TrimSpace(k))] = value
	}

	return formatPropertyString(entries)
}

// ParseBackendString parses a backend configuration string and returns the backend type and remaining parameters.
func ParseBackendString(raw string) (string, map[string]string) {
	props := parsePropertyString(raw)
	if len(props) == 0 {
		return "", map[string]string{}
	}

	typeValue := props["type"]
	delete(props, "type")

	return typeValue, props
}

func escapeQuotes(in string) string {
	return strings.ReplaceAll(in, "\"", "\\\"")
}
