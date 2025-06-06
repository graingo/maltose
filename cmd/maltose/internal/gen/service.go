package gen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

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
	if err := g.generateFile("service", TplGenService, genInfo); err != nil {
		return err
	}

	// Generate Controller
	return g.generateFile("controller", TplGenController, genInfo)
}

func (g *ServiceGenerator) generateFile(fileType, tplContent string, data *GenerationInfo) error {
	var tpl *template.Template
	var err error

	// Custom functions for templates
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
		"ToCamel": strcase.ToCamel,
		"ToKebab": strcase.ToKebab,
		"ToSnake": strcase.ToSnake,
		"FirstLower": func(s string) string {
			if len(s) == 0 {
				return ""
			}
			return strings.ToLower(string(s[0])) + s[1:]
		},
	}

	tpl, err = template.New(fileType).Funcs(funcMap).Parse(tplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Determine output path
	outputPath := filepath.Join(g.DstPath, fileType, data.Module)
	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
		return err
	}
	outputFile := filepath.Join(outputPath, data.FileName)

	// Write the generated file
	return os.WriteFile(outputFile, buffer.Bytes(), 0644)
}

type Parser struct {
	file       *ast.File
	fset       *token.FileSet
	module     string
	moduleRoot string
}

type GenerationInfo struct {
	Module     string
	Service    string
	Controller string
	SvcName    string
	ApiModule  string
	SvcModule  string
	ApiPkg     string
	FileName   string
	Interface  bool
	Functions  []Function
}

type Function struct {
	Name    string
	ReqName string
	ResName string
}

func (p *Parser) Parse() (*GenerationInfo, error) {
	var functions []Function
	var currentReq, currentRes string

	fileName := strings.TrimSuffix(filepath.Base(p.file.Name.Name), ".go")

	info := &GenerationInfo{
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
