package mcfg

import "github.com/graingo/maltose/os/mcfg/internal"

// Merge merges the `src` map into the `dest` map.
// It performs a deep, case-insensitive merge.
// The `dest` map is modified in place.
func Merge(dest, src map[string]any) map[string]any {
	internal.DeepMergeMaps(dest, src)
	return dest
}

func deepCopyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	copyData := make(map[string]any, len(source))
	for key, value := range source {
		copyData[key] = deepCopyValue(value)
	}
	return copyData
}

func deepCopyValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return deepCopyMap(typed)
	case []any:
		copySlice := make([]any, len(typed))
		for i, item := range typed {
			copySlice[i] = deepCopyValue(item)
		}
		return copySlice
	default:
		return value
	}
}
