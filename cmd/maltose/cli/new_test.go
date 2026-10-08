package cli

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/graingo/maltose/cmd/maltose/internal/openapi"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectTargetStaysInsideWorkingDirectory(t *testing.T) {
	cwd := t.TempDir()
	target, err := projectTarget(cwd, "services/example")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(cwd, "services", "example"), target)

	for _, invalid := range []string{"", ".", "..", "../outside", filepath.Join(string(filepath.Separator), "tmp", "outside")} {
		_, err := projectTarget(cwd, invalid)
		assert.Error(t, err, invalid)
	}
}

func TestCreateProjectRewritesModuleAndBuilds(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("local file URLs have platform-specific Git behavior on Windows")
	}
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for the project creation test")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go is required for the project creation test")
	}

	const (
		templateModule = "example.com/maltose-template"
		projectModule  = "example.com/acme/service"
	)

	templateDir := filepath.Join(t.TempDir(), "template")
	require.NoError(t, os.MkdirAll(filepath.Join(templateDir, "internal", "greeting"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "go.mod"), []byte("module "+templateModule+"\n\ngo 1.23\n"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "main.go"), []byte(`package main

import (
	"fmt"

	"example.com/maltose-template/internal/greeting"
)

func main() {
	fmt.Println(greeting.Message())
}
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(templateDir, "internal", "greeting", "greeting.go"), []byte(`package greeting

// Message returns the example greeting.
func Message() string { return "hello" }
`), 0644))
	runCommand(t, templateDir, "git", "init", "--quiet")
	runCommand(t, templateDir, "git", "add", ".")
	runCommand(t, templateDir, "git", "-c", "user.name=Maltose Test", "-c", "user.email=maltose@example.com", "commit", "--quiet", "-m", "initial template")

	workingDir := t.TempDir()
	err := createProject(context.Background(), os.Stdout, os.Stderr, workingDir, "service", projectModule, templateDir)
	require.NoError(t, err)

	projectDir := filepath.Join(workingDir, "service")
	goMod, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	require.NoError(t, err)
	assert.Contains(t, string(goMod), "module "+projectModule)

	mainSource, err := os.ReadFile(filepath.Join(projectDir, "main.go"))
	require.NoError(t, err)
	assert.Contains(t, string(mainSource), projectModule+"/internal/greeting")
	assert.NotContains(t, string(mainSource), templateModule)
	assert.NoDirExists(t, filepath.Join(projectDir, ".git"))

	runCommand(t, projectDir, "go", "test", "./...")
}

func TestRewriteGoImportsLeavesOrdinaryStringsUnchanged(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sample.go")
	source := `package sample

import "example.com/template/pkg"

const moduleName = "example.com/template/pkg"
`
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	require.NoError(t, rewriteGoImports(path, "example.com/template", "example.com/project"))

	rewritten, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(rewritten), `import "example.com/project/pkg"`)
	assert.Contains(t, string(rewritten), `const moduleName = "example.com/template/pkg"`)
}

func runCommand(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	require.NoError(t, err, "%s %s failed: %s", name, strings.Join(args, " "), output)
}

func TestCreateProjectRegeneratesContractWithTemplateDependency(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("local Git fixture uses Unix paths")
	}
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	framework := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	const templateModule = "example.com/contract-template"
	const projectModule = "example.com/contract-service"
	template := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(template, "api"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(template, "cmd"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(template, "go.mod"), []byte("module "+templateModule+"\n\ngo 1.25.0\n\nrequire github.com/graingo/maltose v0.5.0\n"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(template, "api", "api.go"), []byte(`package api
import "github.com/graingo/maltose/frame/m"
type GetReq struct { m.Meta `+"`method:\"GET\" path:\"/example\"`"+` }
type GetRes struct { Message string `+"`json:\"message\"`"+` }
`), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(template, "cmd", "openapi.yaml"), []byte("stale"), 0600))
	require.NoError(t, os.WriteFile(filepath.Join(template, "cmd", "openapi.yaml.manifest.json"), []byte("stale"), 0600))
	runCommand(t, template, "git", "init", "--quiet")
	runCommand(t, template, "git", "add", ".")
	runCommand(t, template, "git", "-c", "user.name=Maltose Test", "-c", "user.email=maltose@example.com", "commit", "--quiet", "-m", "contract template")

	// Exercise the template's next-version dependency against the checkout without
	// writing a local replace into the template or generated project.
	mod, err := os.ReadFile(filepath.Join(framework, "go.mod"))
	require.NoError(t, err)
	temporaryMod := filepath.Join(t.TempDir(), "project.mod")
	mod = []byte(strings.Replace(string(mod), "module github.com/graingo/maltose", "module "+projectModule, 1) + "\nrequire github.com/graingo/maltose v0.5.0\nreplace github.com/graingo/maltose => " + framework + "\n")
	require.NoError(t, os.WriteFile(temporaryMod, mod, 0600))
	sum, err := os.ReadFile(filepath.Join(framework, "go.sum"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(strings.TrimSuffix(temporaryMod, ".mod")+".sum", sum, 0600))
	t.Setenv("GOWORK", "off")
	t.Setenv("GOFLAGS", "-modfile="+temporaryMod)
	workingDir := t.TempDir()
	require.NoError(t, createProject(context.Background(), os.Stdout, os.Stderr, workingDir, "service", projectModule, template))
	project := filepath.Join(workingDir, "service")
	generatedMod, err := os.ReadFile(filepath.Join(project, "go.mod"))
	require.NoError(t, err)
	require.Contains(t, string(generatedMod), projectModule)
	require.Contains(t, string(generatedMod), "github.com/graingo/maltose v0.5.0")
	require.NotContains(t, string(generatedMod), "replace")
	doc := filepath.Join(project, "cmd", "openapi.yaml")
	data, err := os.ReadFile(doc)
	require.NoError(t, err)
	require.Contains(t, string(data), "/example:")
	require.NotContains(t, string(data), templateModule)
	require.NoError(t, openapi.Run(context.Background(), openapi.Config{Source: filepath.Join(project, "api"), Output: doc, Format: "yaml", Check: true}))
	runCommand(t, project, "go", "test", "./...")
}
