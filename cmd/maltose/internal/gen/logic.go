// Package gen contains the common logic for code generation.
package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// LogicGenerator holds the configuration for generating logic files.
type LogicGenerator struct {
	SrcPath    string
	DstPath    string
	Module     string
	ModuleRoot string
}

// Gen generates the logic file from a service interface file.
func (g *LogicGenerator) Gen() error {
	// This generator works on a single file, not a directory.
	return g.genFromFile(g.SrcPath)
}

func (g *LogicGenerator) genFromFile(file string) error {
	p := &LogicParser{
		fset:   token.NewFileSet(),
		module: g.Module,
	}

	genInfo, err := p.Parse(file)
	if err != nil {
		return fmt.Errorf("failed to parse service file %s: %w", file, err)
	}

	// Logic file goes into internal/logic/<module>/<file>.go
	logicOutputPath := filepath.Join(g.DstPath, "logic", genInfo.Module, genInfo.FileName)
	return generateFile(logicOutputPath, "serviceLogic", TplGenServiceLogic, genInfo)
}

// LogicParser parses a service interface file.
type LogicParser struct {
	fset   *token.FileSet
	module string
}

// Parse parses the service interface file and extracts necessary data for the template.
func (p *LogicParser) Parse(filePath string) (*ServiceTplData, error) {
	node, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var serviceName string
	var functions []Function
	imports := make(map[string]string) // alias -> full path

	// 1. Get imports
	for _, i := range node.Imports {
		path := strings.Trim(i.Path.Value, `"`)
		if i.Name != nil {
			imports[i.Name.Name] = path
		} else {
			parts := strings.Split(path, "/")
			imports[parts[len(parts)-1]] = path
		}
	}

	var apiModule, apiPkg string

	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		iface, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}

		serviceName = strings.TrimPrefix(typeSpec.Name.Name, "I")

		for _, method := range iface.Methods.List {
			if len(method.Names) == 0 {
				continue
			}
			funcType, ok := method.Type.(*ast.FuncType)
			if !ok || funcType.Params.NumFields() != 2 || funcType.Results.NumFields() != 2 {
				continue
			}

			reqField := funcType.Params.List[1]
			starExpr, ok := reqField.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			selectorExpr, ok := starExpr.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			reqPkgIdent, ok := selectorExpr.X.(*ast.Ident)
			if !ok {
				continue
			}
			reqPkg := reqPkgIdent.Name
			reqName := selectorExpr.Sel.Name

			resField := funcType.Results.List[0]
			starExpr, ok = resField.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			selectorExpr, ok = starExpr.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			resName := selectorExpr.Sel.Name

			functions = append(functions, Function{
				Name:    method.Names[0].Name,
				ReqName: reqName,
				ResName: resName,
			})

			if apiPkg == "" {
				apiPkg = reqPkg
				apiModule = imports[apiPkg]
			}
		}

		return false // Stop after finding the first interface
	})

	if serviceName == "" {
		return nil, fmt.Errorf("no service interface found in %s", filePath)
	}

	fileName := filepath.Base(filePath)
	moduleName := strings.TrimSuffix(fileName, ".go")
	svcPackage := strings.ReplaceAll(filepath.Join(p.module, "internal", "service"), "\\", "/")

	info := &ServiceTplData{
		Module:     moduleName,
		Service:    serviceName,
		ApiModule:  apiModule,
		ApiPkg:     apiPkg,
		FileName:   fileName,
		Functions:  functions,
		SvcPackage: svcPackage,
	}

	return info, nil
}
