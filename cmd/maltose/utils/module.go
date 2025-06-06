package utils

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// GetModuleInfo finds the go.mod file by traversing up from fromPath,
// and returns the module name and the root path of the module.
func GetModuleInfo(fromPath string) (name, rootPath string, err error) {
	currentPath, err := filepath.Abs(fromPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path for %s: %w", fromPath, err)
	}

	for {
		goModPath := filepath.Join(currentPath, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err == nil {
			return modfile.ModulePath(content), currentPath, nil
		}
		if !os.IsNotExist(err) {
			// File exists but cannot be read.
			return "", "", fmt.Errorf("failed to read go.mod at %s: %w", goModPath, err)
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath { // Reached the root directory
			return "", "", fmt.Errorf("go.mod not found in any parent directory of %s", fromPath)
		}
		currentPath = parent
	}
}
