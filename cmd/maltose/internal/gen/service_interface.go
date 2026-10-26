package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"text/template"
)

// generateInterface inserts signatures inside the service interface, preserving
// application code and existing methods. Function declarations use a separate path.
func (g *ServiceGenerator) generateInterface(path string, data *serviceTplData) error {
	source, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return generateFile(path, "serviceInterface", TplGenServiceInterface, data)
	}
	if err != nil {
		return err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, source, parser.ParseComments)
	if err != nil {
		return err
	}
	var target *ast.InterfaceType
	ast.Inspect(file, func(node ast.Node) bool {
		spec, ok := node.(*ast.TypeSpec)
		if ok && spec.Name.Name == "I"+data.Service {
			target, _ = spec.Type.(*ast.InterfaceType)
		}
		return true
	})
	if target == nil {
		return fmt.Errorf("%s requires interface I%s", path, data.Service)
	}
	existing := map[string]bool{}
	for _, method := range target.Methods.List {
		for _, name := range method.Names {
			existing[name.Name] = true
		}
	}
	addition := *data
	addition.Functions = nil
	for _, method := range data.Functions {
		if !existing[method.Name] {
			addition.Functions = append(addition.Functions, method)
		}
	}
	if len(addition.Functions) == 0 {
		return nil
	}
	// Reuse an application's import alias when it already imports this API package.
	imports := map[string]bool{}
	for _, spec := range file.Imports {
		imported, err := strconv.Unquote(spec.Path.Value)
		if err != nil {
			return err
		}
		imports[imported] = true
		if imported == "context" && spec.Name != nil && spec.Name.Name != "context" {
			return fmt.Errorf("%s: service generation requires the context import name context", path)
		}
		if imported == data.APIModule && spec.Name != nil {
			if spec.Name.Name == "_" || spec.Name.Name == "." {
				return fmt.Errorf("%s: API import requires a named alias", path)
			}
			addition.APIPkg = spec.Name.Name
		}
	}
	tpl, err := template.New("interfaceMethods").Parse(TplGenServiceInterfaceMethodOnly)
	if err != nil {
		return err
	}
	var methods bytes.Buffer
	if err = tpl.Execute(&methods, &addition); err != nil {
		return err
	}
	var importText bytes.Buffer
	if !imports["context"] {
		fmt.Fprintln(&importText, "\nimport \"context\"")
	}
	if !imports[data.APIModule] {
		fmt.Fprintf(&importText, "\nimport %s %q\n", addition.APIPkg, data.APIModule)
	}
	packageEnd := fset.Position(file.Name.End()).Offset
	closing := fset.Position(target.Methods.Closing).Offset
	updated := append([]byte{}, source[:packageEnd]...)
	updated = append(updated, importText.Bytes()...)
	updated = append(updated, source[packageEnd:closing]...)
	updated = append(updated, methods.Bytes()...)
	updated = append(updated, source[closing:]...)
	formatted, err := format.Source(updated)
	if err != nil {
		return fmt.Errorf("format service interface %s: %w", path, err)
	}
	return os.WriteFile(path, formatted, 0644)
}
