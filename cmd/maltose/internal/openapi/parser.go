package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
)

// APIDefinition holds the extracted information for a single API endpoint.
type APIDefinition struct {
	Method      string
	Path        string
	Summary     string
	Tag         string
	Description string
	Request     StructInfo
	Response    StructInfo
}

// StructInfo holds information about a request or response struct.
type StructInfo struct {
	Name   string
	Fields []FieldInfo
}

// FieldInfo holds information about a single struct field.
type FieldInfo struct {
	Name        string
	JSONName    string
	Type        string
	Tag         reflect.StructTag
	Description string
	Required    bool
}

// ParseDir parses all .go files in a directory and its subdirectories,
// and extracts API definitions from files containing "m.Meta".
func ParseDir(dir string) ([]APIDefinition, error) {
	fset := token.NewFileSet()
	goFiles := make([]string, 0)

	utils.PrintInfo("scanning_directory", utils.TplData{"Path": filepath.Base(dir)})
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			goFiles = append(goFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	var apiDefs []APIDefinition
	allStructs := make(map[string]*ast.StructType)

	// First pass: Parse files and collect all structs from files containing m.Meta
	for _, path := range goFiles {
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			// In case of parsing errors, we can choose to log them and continue,
			// or stop the process. For now, let's continue.
			fmt.Fprintf(os.Stderr, "warn: could not parse %s: %v\n", path, err)
			continue
		}

		if !containsMeta(file) {
			continue
		}

		for _, decl := range file.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							allStructs[typeSpec.Name.Name] = structType
						}
					}
				}
			}
		}
	}

	// Second pass: find "Req" structs and build API definitions
	for name, structType := range allStructs {
		if !strings.HasSuffix(name, "Req") {
			continue
		}

		apiDef := APIDefinition{}
		isAPIEntry := false

		// Find m.Meta and extract endpoint info
		for _, field := range structType.Fields.List {
			// Check for embedded m.Meta
			if field.Names == nil {
				if selExpr, ok := field.Type.(*ast.SelectorExpr); ok {
					if x, ok := selExpr.X.(*ast.Ident); ok && x.Name == "m" && selExpr.Sel.Name == "Meta" {
						if field.Tag != nil {
							isAPIEntry = true
							tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
							apiDef.Method = tag.Get("method")
							apiDef.Path = tag.Get("path")
							apiDef.Summary = tag.Get("summary")
							apiDef.Tag = tag.Get("tag")
							apiDef.Description = tag.Get("dc")
						}
						break
					}
				}
			}
		}

		if !isAPIEntry {
			continue
		}

		// Populate request struct info
		apiDef.Request = parseStructInfo(name, structType, apiDef.Method)

		// Find and populate response struct info
		resName := strings.TrimSuffix(name, "Req") + "Res"
		if resStructType, ok := allStructs[resName]; ok {
			apiDef.Response = parseStructInfo(resName, resStructType, "POST") // Method doesn't matter for response schema
		}

		apiDefs = append(apiDefs, apiDef)
	}

	return apiDefs, nil
}

func containsMeta(file *ast.File) bool {
	hasMeta := false
	ast.Inspect(file, func(n ast.Node) bool {
		if selExpr, ok := n.(*ast.SelectorExpr); ok {
			if x, ok := selExpr.X.(*ast.Ident); ok && x.Name == "m" && selExpr.Sel.Name == "Meta" {
				hasMeta = true
				return false // stop inspecting
			}
		}
		return !hasMeta // continue inspecting if meta not found
	})
	return hasMeta
}

func parseStructInfo(name string, structType *ast.StructType, method string) StructInfo {
	info := StructInfo{Name: name}
	for _, field := range structType.Fields.List {
		if len(field.Names) == 0 { // Skip embedded fields
			continue
		}
		fieldName := field.Names[0].Name

		var fieldTypeName string
		if ident, ok := field.Type.(*ast.Ident); ok {
			fieldTypeName = ident.Name
		}

		if fieldTypeName == "" {
			continue // Skip unhandled types for now
		}

		var tag reflect.StructTag
		if field.Tag != nil {
			tag = reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
		}

		jsonName := tag.Get("json")
		if method == "GET" {
			jsonName = tag.Get("form")
		}
		if parts := strings.Split(jsonName, ","); len(parts) > 0 {
			jsonName = parts[0]
		}

		if jsonName == "" {
			jsonName = fieldName
		}

		info.Fields = append(info.Fields, FieldInfo{
			Name:        fieldName,
			JSONName:    jsonName,
			Type:        fieldTypeName,
			Tag:         tag,
			Description: tag.Get("dc"),
			Required:    strings.Contains(tag.Get("binding"), "required"),
		})
	}
	return info
}
