package gen

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
)

// ServiceGenerator holds the configuration for generating services.
type ServiceGenerator struct {
	SrcPath       string // Source path for API definition files
	DstPath       string // Destination path for generated files
	Module        string // Go module name
	ModuleRoot    string // File system path to the module root
	InterfaceMode bool   // Whether to generate service with interface
}

// Gen generates the service and controller files.
func (g *ServiceGenerator) Gen() error {
	return filepath.Walk(g.SrcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			if err := g.genFromFile(path); err != nil {
				return fmt.Errorf("failed to generate from file %s: %w", path, err)
			}
		}
		return nil
	})
}

func (g *ServiceGenerator) genFromFile(file string) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	parser := &Parser{
		file:       node,
		fset:       fset,
		module:     g.Module,
		moduleRoot: g.ModuleRoot,
	}

	genInfo, err := parser.Parse()
	if err != nil {
		return err
	}

	genInfo.Interface = g.InterfaceMode

	// Set SvcModule path
	svcPath := filepath.Join(g.Module, g.DstPath, "service", genInfo.Module)
	genInfo.SvcModule = strings.ReplaceAll(svcPath, "\\", "/")

	// Generate Service
	svcOutputPath := filepath.Join(g.DstPath, "service", genInfo.Module, genInfo.FileName)
	if err := generateFile(svcOutputPath, "service", TplGenService, genInfo); err != nil {
		return err
	}

	// Generate Controller
	controllerOutputPath := filepath.Join(g.DstPath, "controller", genInfo.Module, genInfo.FileName)
	return generateFile(controllerOutputPath, "controller", TplGenController, genInfo)
}

type Parser struct {
	file       *ast.File
	fset       *token.FileSet
	module     string
	moduleRoot string
}

type Function struct {
	Name    string
	ReqName string
	ResName string
}

func (p *Parser) Parse() (*ServiceTplData, error) {
	var functions []Function
	var currentReq, currentRes string

	fileName := strings.TrimSuffix(filepath.Base(p.file.Name.Name), ".go")

	info := &ServiceTplData{
		Module:     fileName,
		Service:    strcase.ToCamel(fileName),
		Controller: strcase.ToCamel(fileName),
		SvcName:    strcase.ToCamel(fileName),
		ApiModule:  "", // Will be calculated below
		SvcModule:  "", // Will be set later
		ApiPkg:     p.file.Name.Name,
		FileName:   fileName + ".go",
		Functions:  nil,
	}

	fullPath := p.fset.File(p.file.Pos()).Name()
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path for %s: %w", fullPath, err)
	}

	if p.moduleRoot != "" {
		relPath, err := filepath.Rel(p.moduleRoot, filepath.Dir(absPath))
		if err != nil {
			return nil, fmt.Errorf("could not get relative path for %s: %w", absPath, err)
		}
		info.ApiModule = filepath.Join(p.module, relPath)
	} else {
		// Fallback for when module root is not found
		apiModuleDir := filepath.Dir(strings.Replace(fullPath, "\\\\", "/", -1))
		apiModuleDir = strings.ReplaceAll(apiModuleDir, "\\", "/")
		if p.module != "" {
			if i := strings.Index(apiModuleDir, p.module); i != -1 {
				info.ApiModule = apiModuleDir[i:]
			}
		}
	}
	info.ApiModule = strings.ReplaceAll(info.ApiModule, "\\", "/")

	ast.Inspect(p.file, func(n ast.Node) bool {
		decl, ok := n.(*ast.GenDecl)
		if !ok || decl.Tok != token.TYPE {
			return true
		}

		for _, spec := range decl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			if strings.HasSuffix(typeSpec.Name.Name, "Req") {
				currentReq = typeSpec.Name.Name
			} else if strings.HasSuffix(typeSpec.Name.Name, "Res") {
				currentRes = typeSpec.Name.Name
			}

			if currentReq != "" && currentRes != "" {
				funcName := strings.TrimSuffix(currentReq, "Req")
				if strings.TrimSuffix(currentRes, "Res") == funcName {
					functions = append(functions, Function{
						Name:    funcName,
						ReqName: currentReq,
						ResName: currentRes,
					})
					currentReq, currentRes = "", ""
				}
			}
		}

		return true
	})

	if len(functions) == 0 {
		return nil, fmt.Errorf("no matching Req/Res struct pairs found")
	}

	info.Functions = functions
	return info, nil
}
