// Package gen contains the common logic for code generation.
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

	"github.com/fatih/color"
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

// Function holds the parsed information of a function.
type serviceFunction struct {
	Name         string
	ReqName      string
	ResName      string
	ReqIsPointer bool
	ResIsPointer bool
}

// ServiceTplData is the data structure for the service template.
type serviceTplData struct {
	Module       string
	Service      string
	Controller   string
	SvcName      string
	ApiModule    string
	ApiPkg       string
	FileName     string
	Version      string
	VersionLower string
	Functions    []serviceFunction
	SvcPackage   string
}

// Gen generates the service and controller files.
func (g *ServiceGenerator) Gen() error {
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

	// The package for the controller to import is always ".../internal/service".
	genInfo.SvcPackage = strings.ReplaceAll(filepath.Join(g.Module, "internal", "service"), "\\", "/")

	// --- Service File Generation (Create if not exist, skip if exist) ---
	svcOutputPath := filepath.Join(g.DstPath, "service", genInfo.FileName)
	if _, err := os.Stat(svcOutputPath); os.IsNotExist(err) {
		// File does not exist, generate skeleton.
		templateName := TplGenService
		if g.InterfaceMode {
			templateName = TplGenServiceInterface
		}
		if err := generateFile(svcOutputPath, "service", templateName, genInfo); err != nil {
			return fmt.Errorf("failed to generate service skeleton: %w", err)
		}
	} else {
		// File exists, skip with a warning.
		color.Yellow("Warning: Service file [%s] already exists, skipping generation.", genInfo.FileName)
	}

	// --- Controller Generation (Create or Append) ---
	// Case 1: Professional layout like api/<module>/<version>/...
	if genInfo.Version != "" && !strings.EqualFold(genInfo.Module, genInfo.Version) {
		// Handle controller struct file (create if not exist, otherwise skip)
		controllerStructPath := filepath.Join(g.DstPath, "controller", genInfo.Module, genInfo.Module+".go")
		if _, err := os.Stat(controllerStructPath); os.IsNotExist(err) {
			if err := generateFile(controllerStructPath, "controllerStruct", TplGenControllerStruct, genInfo); err != nil {
				return fmt.Errorf("failed to generate controller struct: %w", err)
			}
		}

		// Handle controller method file (create or append)
		methodFileName := fmt.Sprintf("%s_%s.go", genInfo.Module, strings.ToLower(genInfo.Version))
		controllerMethodPath := filepath.Join(g.DstPath, "controller", genInfo.Module, methodFileName)
		return g.generateOrAppend(controllerMethodPath, TplGenControllerMethod, TplGenControllerMethodOnly, genInfo)
	}

	// Case 2: Simple layout like api/<version>/...
	controllerPath := filepath.Join(g.DstPath, "controller", genInfo.VersionLower, genInfo.FileName)
	return g.generateOrAppend(controllerPath, TplGenController, TplGenControllerMethodOnly, genInfo)
}

// generateOrAppend handles the logic of creating a new file or appending to an existing one.
func (g *ServiceGenerator) generateOrAppend(filePath, fullTpl, appendTpl string, data *serviceTplData) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File does not exist, generate a new one from scratch.
		return generateFile(filePath, "controller", fullTpl, data)
	}

	// File exists, proceed with append logic.
	existingMethods, err := parseGoFileForMethods(filePath)
	if err != nil {
		return fmt.Errorf("could not parse existing controller file %s: %w", filePath, err)
	}

	var methodsToAppend []serviceFunction
	for _, neededMethod := range data.Functions {
		if _, exists := existingMethods[neededMethod.Name]; !exists {
			methodsToAppend = append(methodsToAppend, neededMethod)
		}
	}

	if len(methodsToAppend) > 0 {
		color.Yellow("Warning: Appending %d new method(s) to existing controller file: %s", len(methodsToAppend), filePath)
		appendData := *data
		appendData.Functions = methodsToAppend
		return appendToFile(filePath, appendTpl, &appendData)
	}

	return nil // Nothing to append
}

// appendToFile executes a template and appends the result to a file.
func appendToFile(filePath, tplContent string, data *serviceTplData) error {
	var buffer bytes.Buffer
	tpl, err := template.New("method").Parse(tplContent)
	if err != nil {
		return fmt.Errorf("failed to parse append template: %w", err)
	}
	if err := tpl.Execute(&buffer, data); err != nil {
		return fmt.Errorf("failed to execute append template: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for appending: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to append to file: %w", err)
	}
	return nil
}

// parseGoFileForMethods parses a Go file and returns a set of its method names.
func parseGoFileForMethods(path string) (map[string]struct{}, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	methods := make(map[string]struct{})
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, isFn := n.(*ast.FuncDecl); isFn && fn.Recv != nil && len(fn.Recv.List) > 0 {
			methods[fn.Name.Name] = struct{}{}
		}
		return true
	})
	return methods, nil
}

type Parser struct {
	file       *ast.File
	fset       *token.FileSet
	module     string
	moduleRoot string
}

func (p *Parser) Parse() (*serviceTplData, error) {
	var functions []serviceFunction
	var moduleName, versionName, structBaseName, fileName string

	// --- Enhanced Path Parsing Logic ---
	fullPath := p.fset.File(p.file.Pos()).Name()
	relPath, err := filepath.Rel(p.moduleRoot, fullPath)
	if err != nil {
		return nil, fmt.Errorf("could not determine file path relative to module root: %w", err)
	}

	parts := strings.Split(filepath.ToSlash(relPath), "/")
	apiIndex := -1
	for i, part := range parts {
		if part == "api" {
			apiIndex = i
			break
		}
	}

	if apiIndex == -1 {
		return nil, fmt.Errorf("path does not contain 'api' directory: %s", fullPath)
	}

	// api/<module>/<version>/<file>.go
	if len(parts) > apiIndex+3 {
		moduleName = parts[apiIndex+1]
		versionName = parts[apiIndex+2]
		fileName = parts[len(parts)-1]
		structBaseName = strings.TrimSuffix(fileName, ".go")
		// api/v1/hello.go -> api/hello/v1/hello.go (module=hello, version=v1)
	} else if len(parts) > apiIndex+2 { // api/<version>/<file>.go
		versionName = parts[apiIndex+1]
		fileName = parts[len(parts)-1]
		structBaseName = strings.TrimSuffix(fileName, ".go")
		moduleName = versionName // In this case, the module is the version
	} else {
		return nil, fmt.Errorf("path format not supported. Use 'api/<version>/<file>.go' or 'api/<module>/<version>/<file>.go': %s", fullPath)
	}

	info := &serviceTplData{
		Module:       moduleName,
		Service:      strcase.ToCamel(structBaseName),
		Controller:   strcase.ToCamel(moduleName) + strcase.ToCamel(versionName),
		SvcName:      strcase.ToCamel(structBaseName),
		ApiModule:    "", // Will be calculated below
		ApiPkg:       p.file.Name.Name,
		FileName:     fileName,
		Version:      strcase.ToCamel(versionName),
		VersionLower: strings.ToLower(versionName),
		Functions:    nil,
	}

	// For simple case, controller name is c<Service>
	if strings.EqualFold(moduleName, versionName) {
		info.Controller = "c" + strcase.ToCamel(structBaseName)
	}

	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, fmt.Errorf("could not get absolute path for %s: %w", fullPath, err)
	}

	if p.moduleRoot != "" {
		relDir, err := filepath.Rel(p.moduleRoot, filepath.Dir(absPath))
		if err != nil {
			return nil, fmt.Errorf("could not get relative path for %s: %w", absPath, err)
		}
		info.ApiModule = filepath.ToSlash(filepath.Join(p.module, relDir))
	} else {
		// Fallback for when module root is not found
		apiModuleDir := filepath.ToSlash(filepath.Dir(fullPath))
		if p.module != "" {
			if i := strings.Index(apiModuleDir, p.module); i != -1 {
				info.ApiModule = apiModuleDir[i:]
			}
		}
	}

	// --- Robust Req/Res Parsing Logic ---
	reqs := make(map[string]bool)
	ress := make(map[string]bool)

	ast.Inspect(p.file, func(n ast.Node) bool {
		spec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}

		if strings.HasSuffix(spec.Name.Name, "Req") {
			reqs[spec.Name.Name] = true
		} else if strings.HasSuffix(spec.Name.Name, "Res") {
			ress[spec.Name.Name] = true
		}
		return true
	})

	for reqName := range reqs {
		funcName := strings.TrimSuffix(reqName, "Req")
		resName := funcName + "Res"
		if ress[resName] {
			functions = append(functions, serviceFunction{
				Name:    funcName,
				ReqName: reqName,
				ResName: resName,
			})
		}
	}

	if len(functions) == 0 {
		return nil, fmt.Errorf("no matching Req/Res struct pairs found in %s", fileName)
	}

	info.Functions = functions
	return info, nil
}
