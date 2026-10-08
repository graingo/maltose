package openapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	Source, Output, Format, Version, Extensions string
	Check                                       bool
}
type sourcePackage struct {
	Dir, ImportPath string
	GoFiles         []string
}
type endpoint struct{ Package, Name string }

func Generate(src, output, format string) error {
	return Run(context.Background(), Config{Source: src, Output: output, Format: format})
}

// Run discovers types and compiles a temporary exporter against the application's
// own Maltose version. Field semantics live entirely in the shared compiler.
func Run(ctx context.Context, c Config) error {
	if c.Format != "yaml" && c.Format != "json" {
		return fmt.Errorf("unsupported OpenAPI output format %q: use yaml or json", c.Format)
	}
	if c.Version == "" {
		c.Version = "3.1.0"
	}
	if c.Version != "3.0.0" && c.Version != "3.1.0" {
		return fmt.Errorf("unsupported OpenAPI version %q: use 3.0.0 or 3.1.0", c.Version)
	}
	src, err := filepath.Abs(c.Source)
	if err != nil {
		return err
	}
	endpoints, err := discover(ctx, src)
	if err != nil {
		return err
	}
	if len(endpoints) == 0 {
		return fmt.Errorf("no m.Meta request structs found in %s", src)
	}
	dir, err := os.MkdirTemp(src, ".maltose-openapi-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	program, err := exporter(endpoints, c)
	if err != nil {
		return err
	}
	runner := filepath.Join(dir, "main.go")
	if err = os.WriteFile(runner, program, 0600); err != nil {
		return err
	}
	generated := filepath.Join(dir, "document")
	cmd := exec.CommandContext(ctx, "go", "run", runner, generated)
	cmd.Dir = src
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile API exporter (requires Maltose contract compiler): %w\n%s", err, output)
	}
	document, err := os.ReadFile(generated)
	if err != nil {
		return err
	}
	manifest, err := os.ReadFile(generated + ".manifest.json")
	if err != nil {
		return err
	}
	outputs := []struct {
		path string
		data []byte
	}{{c.Output, document}, {c.Output + ".manifest.json", manifest}}
	if c.Check {
		for _, out := range outputs {
			existing, err := os.ReadFile(out.path)
			if err != nil {
				return err
			}
			if !bytes.Equal(existing, out.data) {
				return fmt.Errorf("generated file is stale: %s", out.path)
			}
		}
		return nil
	}
	for _, out := range outputs {
		if err = writeFile(out.path, out.data); err != nil {
			return err
		}
	}
	return nil
}
func writeFile(name string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(name), ".openapi-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Chmod(0644); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), name)
}
func discover(ctx context.Context, src string) ([]endpoint, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", "./...")
	cmd.Dir = src
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("load API packages: %w\n%s", err, stderr.String())
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	var endpoints []endpoint
	for {
		var pkg sourcePackage
		err := decoder.Decode(&pkg)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		for _, file := range pkg.GoFiles {
			node, err := parser.ParseFile(token.NewFileSet(), filepath.Join(pkg.Dir, file), nil, 0)
			if err != nil {
				return nil, err
			}
			ast.Inspect(node, func(n ast.Node) bool {
				spec, ok := n.(*ast.TypeSpec)
				if !ok || !ast.IsExported(spec.Name.Name) || !strings.HasSuffix(spec.Name.Name, "Req") {
					return true
				}
				body, ok := spec.Type.(*ast.StructType)
				if !ok {
					return false
				}
				for _, field := range body.Fields.List {
					if len(field.Names) != 0 {
						continue
					}
					selector, ok := field.Type.(*ast.SelectorExpr)
					if ok && selector.Sel.Name == "Meta" {
						endpoints = append(endpoints, endpoint{pkg.ImportPath, spec.Name.Name})
						break
					}
				}
				return false
			})
		}
	}
	sort.Slice(endpoints, func(i, j int) bool {
		return endpoints[i].Package+endpoints[i].Name < endpoints[j].Package+endpoints[j].Name
	})
	return endpoints, nil
}
func exporter(endpoints []endpoint, c Config) ([]byte, error) {
	var b strings.Builder
	b.WriteString("package main\nimport (\"os\";\"fmt\";\"github.com/graingo/maltose/net/mhttp/contract\"\n")
	aliases := map[string]string{}
	for _, e := range endpoints {
		if aliases[e.Package] == "" {
			alias := fmt.Sprintf("api%d", len(aliases))
			aliases[e.Package] = alias
			fmt.Fprintf(&b, "%s %q\n", alias, e.Package)
		}
	}
	if c.Extensions != "" {
		fmt.Fprintf(&b, "extensions %q\n", c.Extensions)
	}
	b.WriteString(")\nfunc main(){if err:=run();err!=nil{fmt.Fprintln(os.Stderr,err);os.Exit(1)}}\nfunc run()error{operations:=[]*contract.Operation{}\n")
	for _, e := range endpoints {
		alias := aliases[e.Package]
		fmt.Fprintf(&b, "{op,err:=contract.Compile(contract.TypeOf[%s.%s](),contract.TypeOf[%s.%sRes]());if err!=nil{return err};operations=append(operations,op)}\n", alias, e.Name, alias, strings.TrimSuffix(e.Name, "Req"))
	}
	configure := "nil"
	if c.Extensions != "" {
		configure = "extensions.Configure"
	}
	fmt.Fprintf(&b, "artifact,err:=contract.Generate(operations,contract.Options{Version:%q,Format:%q},%s);if err!=nil{return err};if err=os.WriteFile(os.Args[1],artifact.Document,0600);err!=nil{return err};return os.WriteFile(os.Args[1]+\".manifest.json\",artifact.Manifest,0600) }", c.Version, c.Format, configure)
	return format.Source([]byte(b.String()))
}
