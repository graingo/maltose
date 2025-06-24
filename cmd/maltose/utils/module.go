package utils

import (
	"os"
	"path/filepath"

	"github.com/graingo/maltose/errors/merror"
	"golang.org/x/mod/modfile"
)

// GetModuleInfo finds the go.mod file by traversing up from fromPath,
// and returns the module name and the root path of the module.
func GetModuleInfo(fromPath string) (name, rootPath string, err error) {
	currentPath, err := filepath.Abs(fromPath)
	if err != nil {
		return "", "", merror.Wrapf(err, "failed to get absolute path for %s", fromPath)
	}

	for {
		goModPath := filepath.Join(currentPath, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err == nil {
			return modfile.ModulePath(content), currentPath, nil
		}
		if !os.IsNotExist(err) {
			// File exists but cannot be read.
			return "", "", merror.Wrapf(err, "failed to read go.mod at %s", goModPath)
		}

		parent := filepath.Dir(currentPath)
		if parent == currentPath { // Reached the root directory
			return "", "", merror.Newf("go.mod not found in any parent directory of %s", fromPath)
		}
		currentPath = parent
	}
}
