package openapi

import (
	"go/ast"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDirRejectsDuplicateStructNames(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "one"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "two"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "one", "types.go"), []byte("package one\ntype User struct{}"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "two", "types.go"), []byte("package two\ntype User struct{}"), 0o644))

	_, _, err := ParseDir(dir)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate struct name")
}

func TestParseDirReturnsSyntaxErrors(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "broken.go"), []byte("package api\nfunc {"), 0o644))
	_, _, err := ParseDir(dir)
	assert.Error(t, err)
}

func TestBuildSpecRejectsDuplicateOperationsAndUnsupportedMethods(t *testing.T) {
	definition := APIDefinition{Method: "GET", Path: "/users"}
	_, err := BuildSpec([]APIDefinition{definition, definition}, "test", map[string]*ast.StructType{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate GET")

	definition.Method = "PATCH"
	_, err = BuildSpec([]APIDefinition{definition}, "test", map[string]*ast.StructType{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported HTTP method")
}
