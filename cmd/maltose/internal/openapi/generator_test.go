package openapi

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestGenerateRejectsUnsupportedFormat(t *testing.T) {
	err := Generate(".", "openapi.toml", "toml")
	assert.EqualError(t, err, `unsupported OpenAPI output format "toml": use yaml or json`)
}

func TestSharedCompilerEndToEnd(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	framework := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../.."))
	dir := t.TempDir()
	mod, err := os.ReadFile(filepath.Join(framework, "go.mod"))
	require.NoError(t, err)
	mod = []byte(strings.Replace(string(mod), "module github.com/graingo/maltose", "module example.test/app", 1) + "\nrequire github.com/graingo/maltose v0.0.0\nreplace github.com/graingo/maltose => " + framework + "\n")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), mod, 0600))
	sum, err := os.ReadFile(filepath.Join(framework, "go.sum"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), sum, 0600))
	api := filepath.Join(dir, "api")
	require.NoError(t, os.MkdirAll(api, 0755))
	source := "package api\nimport \"github.com/graingo/maltose/util/mmeta\"\ntype CreateReq struct {mmeta.Meta `method:\"POST\" path:\"/items\" status:\"201\" operation_id:\"create\"`; Count int `json:\"count\" field:\"required\" binding:\"min=0\"`}\ntype CreateRes struct {ID string `json:\"id\"`}\n"
	require.NoError(t, os.WriteFile(filepath.Join(api, "api.go"), []byte(source), 0600))
	extension := filepath.Join(dir, "extension")
	require.NoError(t, os.MkdirAll(extension, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(extension, "extension.go"), []byte("package extension\nimport \"github.com/graingo/maltose/net/mhttp/contract\"\nfunc Configure(e *contract.Extensions)error{return e.ResponseHeader(\"create\",\"ETag\",\"Version\",map[string]any{\"type\":\"string\"})}"), 0600))
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "-mod=mod")
	output := filepath.Join(dir, "openapi.json")
	config := Config{Source: api, Output: output, Format: "json", Version: "3.1.0", Extensions: "example.test/app/extension"}
	require.NoError(t, Run(context.Background(), config))
	first, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Contains(t, string(first), `"201"`)
	require.Contains(t, string(first), `"Etag"`)
	require.NoError(t, Run(context.Background(), config))
	second, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Equal(t, first, second)
	// The command's title/version are applied through the same Info API as the extension.
	for _, format := range []string{"json", "yaml"} {
		config.Format = format
		config.Title = "App API"
		config.APIVersion = "0.2.0"
		require.NoError(t, Run(context.Background(), config))
		document, err := os.ReadFile(output)
		require.NoError(t, err)
		require.Contains(t, string(document), "App API")
		require.Contains(t, string(document), "0.2.0")
		manifest, err := os.ReadFile(output + ".manifest.json")
		require.NoError(t, err)
		config.Check = true
		require.NoError(t, Run(context.Background(), config))
		config.Title = "Changed API"
		require.ErrorContains(t, Run(context.Background(), config), "stale")
		unchanged, err := os.ReadFile(output)
		require.NoError(t, err)
		require.Equal(t, document, unchanged)
		unchanged, err = os.ReadFile(output + ".manifest.json")
		require.NoError(t, err)
		require.Equal(t, manifest, unchanged)
		config.Check = false
	}
	config.Title = "App API"
	require.NoError(t, os.WriteFile(filepath.Join(extension, "extension.go"), []byte(`package extension
import "github.com/graingo/maltose/net/mhttp/contract"
func Configure(e *contract.Extensions)error {
 if err:=e.Info(contract.Info{Title:"App API",Description:"Application description"});err!=nil{return err}
 if err:=e.Server(contract.Server{URL:"/"});err!=nil{return err}
 return e.Tag(contract.Tag{Name:"Items",Description:"Item operations"})
}`), 0600))
	require.NoError(t, Run(context.Background(), config))
	config.Title = "Conflicting API"
	require.ErrorContains(t, Run(context.Background(), config), "info.title conflicts")
	config.Title = "App API"

	config.Check = true
	require.NoError(t, Run(context.Background(), config))
	require.NoError(t, os.WriteFile(output, []byte("stale"), 0600))
	require.ErrorContains(t, Run(context.Background(), config), "stale")
	current, err := os.ReadFile(output)
	require.NoError(t, err)
	require.Equal(t, "stale", string(current))
}
