// Package gen contains the common logic for code generation.
package gen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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

type logicFunction struct {
	Name         string
	ReqName      string
	ResName      string
	ReqIsPointer bool
	ResIsPointer bool
}

// logicTplData is the data structure for the logic template.
type logicTplData struct {
	Module       string
	Service      string
	Controller   string
	SvcName      string
	ApiModule    string
	ApiPkg       string
	FileName     string
	Version      string
	VersionLower string
	Functions    []logicFunction
	SvcPackage   string
}

// Gen generates the logic file from a service interface file.
func (g *LogicGenerator) Gen() error {
	return filepath.Walk(g.SrcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			return g.genFromFile(path)
		}
		return nil
	})
}

func (g *LogicGenerator) genFromFile(file string) error {
	p := &LogicParser{
		fset:   token.NewFileSet(),
		module: g.Module,
	}

	genInfo, err := p.Parse(file)
	if err != nil {
		// This can happen if the file is not a service interface, which is fine.
		// We just skip it.
		return nil
	}
	if genInfo == nil || len(genInfo.Functions) == 0 {
		// Not a valid service interface with methods, skip.
		return nil
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
func (p *LogicParser) Parse(filePath string) (*logicTplData, error) {
	node, err := parser.ParseFile(p.fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var serviceName string
	var functions []logicFunction
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

			// --- Relaxed Request Parsing ---
			reqField := funcType.Params.List[1]
			var reqPkg, reqName string

			// Handle both pointer (*pkg.Type) and non-pointer (pkg.Type)
			var reqSelector *ast.SelectorExpr
			reqIsPointer := false
			if starExpr, isStar := reqField.Type.(*ast.StarExpr); isStar {
				reqIsPointer = true
				if selector, isSelector := starExpr.X.(*ast.SelectorExpr); isSelector {
					reqSelector = selector
				}
			} else if selector, isSelector := reqField.Type.(*ast.SelectorExpr); isSelector {
				reqSelector = selector
			}
			if reqSelector == nil {
				continue
			}

			if reqPkgIdent, ok := reqSelector.X.(*ast.Ident); ok {
				reqPkg = reqPkgIdent.Name
				reqName = reqSelector.Sel.Name
			} else {
				continue
			}

			// --- Relaxed Response Parsing ---
			resField := funcType.Results.List[0] // We only care about the first return value
			var resName string
			var resSelector *ast.SelectorExpr
			resIsPointer := false

			if starExpr, isStar := resField.Type.(*ast.StarExpr); isStar {
				resIsPointer = true
				if selector, isSelector := starExpr.X.(*ast.SelectorExpr); isSelector {
					resSelector = selector
				}
			} else if selector, isSelector := resField.Type.(*ast.SelectorExpr); isSelector {
				resSelector = selector
			}
			if resSelector == nil {
				continue
			}

			// We don't need to validate the package part of the response, just get the type name.
			resName = resSelector.Sel.Name

			functions = append(functions, logicFunction{
				Name:         method.Names[0].Name,
				ReqName:      reqName,
				ResName:      resName,
				ReqIsPointer: reqIsPointer,
				ResIsPointer: resIsPointer,
			})

			if apiPkg == "" {
				apiPkg = reqPkg
				apiModule = imports[apiPkg]
			}
		}

		return false // Stop after finding the first interface
	})

	if serviceName == "" {
		return nil, nil // Not a service interface file, just skip.
	}

	fileName := filepath.Base(filePath)
	moduleName := strings.TrimSuffix(fileName, ".go")
	svcPackage := strings.ReplaceAll(filepath.Join(p.module, "internal", "service"), "\\", "/")

	info := &logicTplData{
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
