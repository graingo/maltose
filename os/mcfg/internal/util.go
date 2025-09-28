package internal

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"
)

// SearchMap is a case-insensitive search function for a given path in a map.
// It navigates through a map of string keys to find a value based on a path slice.
func SearchMap(source map[string]any, path []string) any {
	if len(path) == 0 {
		return source
	}

	// Find the next key in a case-insensitive manner.
	key := path[0]
	var nextVal any
	var found bool

	for k, v := range source {
		if strings.EqualFold(k, key) {
			nextVal = v
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	if len(path) == 1 {
		return nextVal
	}

	// If we need to go deeper, the next value must be a map.
	// We use cast.ToStringMap for robust conversion from map[any]any etc.
	nestedMap, err := cast.ToStringMapE(nextVal)
	if err != nil {
		// It's not a map, so we can't go deeper.
		return nil
	}

	return SearchMap(nestedMap, path[1:])
}

// DeepMergeMaps performs a deep, case-insensitive merge of `src` into `dest`.
func DeepMergeMaps(dest, src map[string]any) {
	for srcK, srcV := range src {
		// Find the key in dest, case-insensitively.
		destK, found := FindCaseInsensitiveKey(dest, srcK)

		if !found {
			// If the key doesn't exist in dest, just add it.
			dest[srcK] = srcV
			continue
		}

		// The key exists in dest.
		destV := dest[destK]

		// Try to cast both values to maps.
		srcMap, srcIsMap := srcV.(map[string]any)
		destMap, destIsMap := destV.(map[string]any)

		// If both are maps, we recurse.
		if srcIsMap && destIsMap {
			DeepMergeMaps(destMap, srcMap)
			dest[destK] = destMap
		} else {
			// Not both are maps, so src overwrites dest.
			// If keys have different casing (e.g., 'Server' vs 'server'),
			// we prefer the key from the src map.
			if destK != srcK {
				delete(dest, destK)
			}
			dest[srcK] = srcV
		}
	}
}

// FindCaseInsensitiveKey finds a key in a map case-insensitively
// and returns the actual key and a boolean indicating if it was found.
func FindCaseInsensitiveKey(source map[string]any, key string) (string, bool) {
	for k := range source {
		if strings.EqualFold(k, key) {
			return k, true
		}
	}
	return "", false
}

var (
	supportedFileTypes = []string{"yaml", "yml", "json", "toml"}
	defaultConfigDir   = []string{".", "/", "config", "/config"}
)

// SearchConfigFile searches for the configuration file in default directories.
// It searches for files with the given `name` and supported extensions.
func SearchConfigFile(name string) (path string, found bool) {
	for _, dir := range defaultConfigDir {
		for _, ext := range supportedFileTypes {
			filePath := filepath.Join(dir, name+"."+ext)
			if _, err := os.Stat(filePath); err == nil {
				return filePath, true
			}
		}
	}
	// Also check for the name directly, in case it includes the extension.
	if stat, err := os.Stat(name); err == nil && !stat.IsDir() {
		return name, true
	}
	return "", false
}
