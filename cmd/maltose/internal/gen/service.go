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

	"go/format"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/iancoleman/strcase"
)

// ServiceGenerator holds the configuration for generating services.
type ServiceGenerator struct {
	Src               string // Source path for API definition files
	Dst               string // Destination path for generated files
	ServiceName       string // The name for single service generation.
	ModuleName        string // Go module name
	ModuleRoot        string // File system path to the module root
	InterfaceMode     bool   // Whether to generate service with interface
	processedServices map[string]bool
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

func NewServiceGenerator(src, dst, serviceName string, interfaceMode bool) (*ServiceGenerator, error) {
	pathForModule := src
	if src == "" {
		pathForModule = dst
	}

	absPath, err := filepath.Abs(pathForModule)
	if err != nil {
		return nil, merror.Wrap(err, "failed to get absolute source path")
	}

	moduleName, moduleRoot, err := utils.GetModuleInfo(absPath)
	if err != nil {
		return nil, merror.Wrap(err, "could not find go.mod")
	}

	return &ServiceGenerator{
		Src:               src,
		Dst:               dst,
		ServiceName:       serviceName,
		ModuleName:        moduleName,
		ModuleRoot:        moduleRoot,
		InterfaceMode:     interfaceMode,
		processedServices: make(map[string]bool),
	}, nil
}

// Gen generates the service and controller files.
func (g *ServiceGenerator) Gen() error {
	if g.ServiceName != "" {
		return g.genSimpleService()
	}

	utils.PrintInfo("🔍 Scanning directory: {{.Path}}", utils.TplData{"Path": filepath.Base(g.Src)})
	return filepath.Walk(g.Src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			if err := g.genFromFile(path); err != nil {
				// A real error occurred, stop the walk.
				return merror.Wrap(err, "failed to generate service file")
			}
		}
		return nil
	})
}

func (g *ServiceGenerator) genSimpleService() error {
	fileName := strings.TrimSuffix(g.ServiceName, ".go")
	outputPath := filepath.Join(g.ModuleRoot, g.Dst, "service", fileName+".go")

	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		utils.PrintWarn("  -> ⏩ Skipping service file {{.Path}} (already exists)", utils.TplData{"Path": outputPath})
		return nil
	}

	data := serviceTplData{
		Service: strcase.ToCamel(fileName),
	}

	return generateFile(outputPath, "serviceInterface", TplGenServiceInterface, &data)
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
		module:     g.ModuleName,
		moduleRoot: g.ModuleRoot,
	}

	info, err := parser.Parse()
	if err != nil {
		return merror.Wrapf(err, "failed to parse file %s", file)
	}

	if info == nil {
		// This happens when a file is parsed but contains no valid Req/Res structs.
		// It's not an error, we just skip it by returning nil.
		utils.PrintWarn("⚠️ No valid API definitions found in {{.File}}, skipping.", utils.TplData{"File": filepath.Base(file)})
		return nil
	}

	// The package for the controller to import is always ".../internal/service".
	info.SvcPackage = strings.ReplaceAll(filepath.Join(g.ModuleName, "internal", "service"), "\\", "/")

	// --- Service File Generation (Create or Append based on Module) ---
	svcOutputPath := filepath.Join(g.Dst, "service", info.Module+".go")

	svcFullTpl := TplGenService
	svcAppendTpl := TplGenServiceMethodOnly
	if g.InterfaceMode {
		svcFullTpl = TplGenServiceInterface
		svcAppendTpl = TplGenServiceInterfaceMethodOnly
	}

	// Use generateOrAppend to create the service file or append to it.
	// Note: We need to adapt the logic slightly for services, as the `data` for the template
	// needs to have its `Service` field based on the module, not the file.
	serviceData := *info
	serviceData.Service = strcase.ToCamel(info.Module) // Use module name for the service struct.

	if err := g.generateOrAppend(svcOutputPath, svcFullTpl, svcAppendTpl, &serviceData); err != nil {
		return merror.Wrap(err, "failed to generate or append service file")
	}

	// --- Controller Generation (Create or Append) ---
	// Case 1: Professional layout like api/<module>/<version>/...
	if info.Version != "" && !strings.EqualFold(info.Module, info.Version) {
		// Handle controller struct file (create if not exist, otherwise skip)
		controllerStructPath := filepath.Join(g.Dst, "controller", info.Module, info.Module+".go")
		if _, err := os.Stat(controllerStructPath); os.IsNotExist(err) {
			if err := generateFile(controllerStructPath, "controllerStruct", TplGenControllerStruct, info); err != nil {
				return merror.Wrap(err, "failed to generate controller struct")
			}
		}

		// Handle controller method file (create or append)
		methodFileName := fmt.Sprintf("%s_%s.go", info.Module, strings.ToLower(info.Version))
		controllerMethodPath := filepath.Join(g.Dst, "controller", info.Module, methodFileName)
		return g.generateOrAppend(controllerMethodPath, TplGenControllerMethod, TplGenControllerMethodOnly, info)
	}

	// Case 2: Simple layout like api/<version>/...
	controllerPath := filepath.Join(g.Dst, "controller", info.VersionLower, info.FileName)
	return g.generateOrAppend(controllerPath, TplGenController, TplGenControllerMethodOnly, info)
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
		return merror.Wrapf(err, "could not parse existing controller file %s", filePath)
	}

	var methodsToAppend []serviceFunction
	for _, neededMethod := range data.Functions {
		if _, exists := existingMethods[neededMethod.Name]; !exists {
			methodsToAppend = append(methodsToAppend, neededMethod)
		}
	}

	if len(methodsToAppend) > 0 {
		utils.PrintSuccess("✨ Appended {{.Count}} new methods to {{.Path}}.", utils.TplData{
			"Count": len(methodsToAppend),
			"Path":  filePath,
		})
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
		return merror.Wrap(err, "failed to parse append template")
	}
	if err := tpl.Execute(&buffer, data); err != nil {
		return merror.Wrap(err, "failed to execute append template")
	}

	// Format the generated code before appending.
	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		// This is unlikely to happen with well-formed templates, but handle it.
		utils.PrintWarn("⚠️ failed to format source for {{.Path}}, writing unformatted code. Error: {{.Error}}", utils.TplData{"Path": filePath, "Error": err})
		formatted = buffer.Bytes() // Append unformatted code on error
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return merror.Wrap(err, "failed to open file for appending")
	}
	defer f.Close()

	if _, err := f.Write(formatted); err != nil {
		return merror.Wrap(err, "failed to append to file")
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
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return nil, merror.Wrapf(err, "could not get absolute path for %s", fullPath)
	}

	relPath, err := filepath.Rel(p.moduleRoot, absPath)
	if err != nil {
		return nil, merror.Wrap(err, "could not determine file path relative to module root")
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
		return nil, merror.Newf("path does not contain 'api' directory: %s", fullPath)
	}

	// api/<module>/<version>/<file>.go
	if len(parts) > apiIndex+3 {
		moduleName = parts[apiIndex+1]
		versionName = parts[apiIndex+2]
		fileName = parts[len(parts)-1]
		structBaseName = strings.TrimSuffix(fileName, ".go")
		// api/v1/hello.go -> api/hello/v1/hello.go (module=hello, version=v1)
	} else if len(parts) > apiIndex+2 { // Covers api/<version>/<file>.go AND api/<module>/<file>.go
		part1 := parts[apiIndex+1]
		fileName = parts[len(parts)-1]
		structBaseName = strings.TrimSuffix(fileName, ".go")

		// Heuristic: if the directory name looks like a version (v1, v2...), treat it as a version.
		isVersionLike := len(part1) > 1 && part1[0] == 'v' && part1[1] >= '0' && part1[1] <= '9'

		if isVersionLike { // Case: api/v1/hello.go
			versionName = part1
			moduleName = part1 // Treat version as module for simplicity in this layout
		} else { // Case: api/hello/hello.go
			moduleName = part1
			versionName = "v1" // Default version to v1 as per documentation
		}
	} else {
		return nil, merror.Newf("path format not supported. Use 'api/<version>/<file>.go' or 'api/<module>/<version>/<file>.go' or 'api/<module>/<file>.go': %s", fullPath)
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

	if p.moduleRoot != "" {
		relDir, err := filepath.Rel(p.moduleRoot, filepath.Dir(absPath))
		if err != nil {
			return nil, merror.Wrapf(err, "could not get relative path for %s", absPath)
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

	// After iterating through all declarations, if no functions were found,
	// it means the file might be a support file (e.g., defining shared types)
	// and not an API entrypoint file, so we should skip it.
	if len(functions) == 0 {
		return nil, nil
	}

	info.Functions = functions
	return info, nil
}
