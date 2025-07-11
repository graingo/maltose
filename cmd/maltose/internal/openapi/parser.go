package openapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
)

// APIDefinition holds the extracted information for a single API endpoint.
type APIDefinition struct {
	Method      string
	Path        string
	Group       string // The group prefix for the path
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
func ParseDir(dir string) ([]APIDefinition, map[string]*ast.StructType, error) {
	fset := token.NewFileSet()
	var apiDefs []APIDefinition
	allStructs := make(map[string]*ast.StructType)
	fileToPkgPath := make(map[string]string) // file path -> package path

	// First pass: Walk files to get their package paths relative to the root `dir`
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			relPath, _ := filepath.Rel(dir, filepath.Dir(path))
			fileToPkgPath[path] = filepath.ToSlash(relPath)
		}
		return nil
	})
	if err != nil {
		return nil, nil, merror.Wrapf(err, "failed to walk directory %s", dir)
	}

	// Second pass: Parse files and collect all structs
	for path := range fileToPkgPath {
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			utils.PrintWarn("⚠️ Could not parse {{.Path}}: {{.Error}}", utils.TplData{"Path": path, "Error": err})
			continue
		}

		if !containsMeta(file) {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.TYPE {
				for _, spec := range genDecl.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if structType, ok := typeSpec.Type.(*ast.StructType); ok {
							// Store the struct along with its source file path for later lookup
							allStructs[typeSpec.Name.Name] = structType
							// A bit of a hack: store file path in a field's doc comment
							if structType.Fields != nil && len(structType.Fields.List) > 0 {
								if structType.Fields.List[0].Doc == nil {
									structType.Fields.List[0].Doc = &ast.CommentGroup{}
								}
								structType.Fields.List[0].Doc.List = append(structType.Fields.List[0].Doc.List, &ast.Comment{Text: "// @source:" + path})
							}
						}
					}
				}
			}
			return true
		})
	}

	// Third pass: find "Req" structs, build API definitions and calculate paths
	for name, structType := range allStructs {
		if !strings.HasSuffix(name, "Req") {
			continue
		}

		// Retrieve the source file path from our hack
		var sourcePath string
		if structType.Fields != nil && len(structType.Fields.List) > 0 && structType.Fields.List[0].Doc != nil {
			for _, comment := range structType.Fields.List[0].Doc.List {
				if strings.HasPrefix(comment.Text, "// @source:") {
					sourcePath = strings.TrimPrefix(comment.Text, "// @source:")
					break
				}
			}
		}
		if sourcePath == "" {
			continue // Should not happen if our hack works
		}
		pkgPath := fileToPkgPath[sourcePath]

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
							apiDef.Group = tag.Get("group")
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

		// Calculate final path
		if apiDef.Group != "" {
			// If group is explicitly set, use it.
			if apiDef.Group == "/" {
				// Use path directly, but ensure it starts with a slash
				apiDef.Path = "/" + strings.TrimPrefix(apiDef.Path, "/")
			} else {
				apiDef.Path = "/" + strings.TrimPrefix(filepath.ToSlash(filepath.Join(apiDef.Group, apiDef.Path)), "/")
			}
		} else {
			// Otherwise, derive a prefix from the file's directory path.
			var prefix string
			if pkgPath != "" && pkgPath != "." {
				prefix = filepath.ToSlash(pkgPath) // Default prefix is the full directory path.
				parts := strings.Split(prefix, "/")

				// Try to find a version string like "v1", "v2", etc.
				for _, part := range parts {
					if len(part) > 1 && part[0] == 'v' && part[1] >= '0' && part[1] <= '9' {
						// If found, construct the prefix as "api/<version>" per user request.
						prefix = "api/" + part
						break
					}
				}
			}

			// Join the calculated prefix with the path from the tag.
			if prefix != "" {
				apiDef.Path = filepath.Join(prefix, apiDef.Path)
			}
			// Ensure the final path starts with a single slash.
			apiDef.Path = "/" + strings.TrimPrefix(filepath.ToSlash(apiDef.Path), "/")
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

	return apiDefs, allStructs, nil
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

		fieldTypeName := extractTypeName(field.Type)
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

// extractTypeName extracts type name from ast.Expr, supporting various type forms
func extractTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		// Simple type: string, int, User
		return t.Name
	case *ast.StarExpr:
		// Pointer type: *User
		if innerType := extractTypeName(t.X); innerType != "" {
			return "*" + innerType
		}
	case *ast.ArrayType:
		// Array/slice type: []User, [5]int
		if innerType := extractTypeName(t.Elt); innerType != "" {
			return "[]" + innerType
		}
	case *ast.MapType:
		// Map type: map[string]User
		keyType := extractTypeName(t.Key)
		valueType := extractTypeName(t.Value)
		if keyType != "" && valueType != "" {
			return "map[" + keyType + "]" + valueType
		}
	case *ast.SelectorExpr:
		// Qualified type: time.Time, pkg.User
		if x, ok := t.X.(*ast.Ident); ok {
			return x.Name + "." + t.Sel.Name
		}
	case *ast.InterfaceType:
		// Interface type: interface{}
		return "interface{}"
	}
	return ""
}
