package openapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
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

// ParseDir parses all .go files in a directory and extracts API definitions.
func ParseDir(dir string) ([]APIDefinition, error) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, func(fi os.FileInfo) bool {
		return !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go")
	}, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory %s: %w", dir, err)
	}

	var apiDefs []APIDefinition
	allStructs := make(map[string]*ast.StructType)
	// First pass: collect all struct type specs
	for _, pkg := range pkgs {
		for _, file := range pkg.Files {
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

func parseStructInfo(name string, structType *ast.StructType, method string) StructInfo {
	info := StructInfo{Name: name}
	for _, field := range structType.Fields.List {
		if field.Names == nil || len(field.Names) == 0 { // Skip embedded fields
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
