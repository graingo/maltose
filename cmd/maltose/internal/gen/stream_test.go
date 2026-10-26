package gen

import (
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRawStreamServiceAndLogicGeneration(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	framework := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	dir := t.TempDir()
	mod, err := os.ReadFile(filepath.Join(framework, "go.mod"))
	require.NoError(t, err)
	mod = []byte(strings.Replace(string(mod), "module github.com/graingo/maltose", "module example.test/streams", 1) + "\nrequire github.com/graingo/maltose v0.0.0\nreplace github.com/graingo/maltose => " + filepath.ToSlash(framework) + "\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), mod, 0600))
	sum, err := os.ReadFile(filepath.Join(framework, "go.sum"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0600))
	source, err := os.ReadFile(filepath.Join(framework, "examples/raw-stream/api/file.go"))
	require.NoError(t, err)
	api := filepath.Join(dir, "api", "files", "v1")
	require.NoError(t, os.MkdirAll(api, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(api, "file.go"), []byte(strings.Replace(string(source), "package api", "package v1", 1)), 0600))
	internal := filepath.Join(dir, "internal")
	snapshot := func() map[string]string {
		files := map[string]string{}
		require.NoError(t, filepath.WalkDir(internal, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			files[path] = string(content)
			return err
		}))
		return files
	}
	generate := func() {
		service, err := NewServiceGenerator(api, internal, "", true)
		require.NoError(t, err)
		require.NoError(t, service.Gen())
		logic, err := NewLogicGenerator(filepath.Join(internal, "service"), "internal", false)
		require.NoError(t, err)
		require.NoError(t, logic.Gen())
	}
	generate()
	first := snapshot()
	generate()
	require.Equal(t, first, snapshot())
	// Adding a new endpoint updates the interface, controller and logic together.
	addition := "\ntype InspectReq struct { mmeta.Meta `method:\"GET\" path:\"/info\"` }\ntype InspectRes struct { Count int `json:\"count\"` }\n"
	filePath := filepath.Join(api, "file.go")
	apiSource, err := os.ReadFile(filePath)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filePath, append(apiSource, []byte(addition)...), 0600))
	generate()
	updated := snapshot()
	require.NotEqual(t, first, updated)
	generate()
	require.Equal(t, updated, snapshot())
	require.NotEmpty(t, first)
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "")
	command := exec.Command("go", "test", "-mod=mod", "./...")
	command.Dir = dir
	output, err := command.CombinedOutput()
	require.NoError(t, err, "generated application: %s", output)
}
