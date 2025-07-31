package mcfg

import "github.com/graingo/maltose/os/mcfg/internal"

// Merge merges the `src` map into the `dest` map.
// It performs a deep, case-insensitive merge.
// The `dest` map is modified in place.
func Merge(dest, src map[string]any) map[string]any {
	internal.DeepMergeMaps(dest, src)
	return dest
}
