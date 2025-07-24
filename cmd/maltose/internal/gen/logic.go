// Package gen contains the common logic for code generation.
package gen

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"go/format"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
)

// LogicGenerator holds the configuration for generating logic files.
type LogicGenerator struct {
	Src        string
	Dst        string
	ModuleName string
	ModuleRoot string
	Overwrite  bool
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
	Module     string
	Service    string
	ApiModule  string
	ApiPkg     string
	FileName   string
	Functions  []logicFunction
	SvcPackage string
}

func NewLogicGenerator(src, dst string, overwrite bool) (*LogicGenerator, error) {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return nil, merror.Wrap(err, "failed to get absolute source path")
	}

	moduleName, moduleRoot, err := utils.GetModuleInfo(absSrc)
	if err != nil {
		return nil, merror.Wrap(err, "could not find go.mod")
	}

	return &LogicGenerator{
		Src:        absSrc,
		Dst:        dst,
		ModuleName: moduleName,
		ModuleRoot: moduleRoot,
		Overwrite:  overwrite,
	}, nil
}

// Gen generates the logic file from a service interface file.
func (g *LogicGenerator) Gen() error {
	utils.PrintInfo("🔍 Scanning directory: {{.Path}}", utils.TplData{"Path": filepath.Base(g.Src)})
	generatedPackages := make(map[string]struct{})

	walkErr := filepath.Walk(g.Src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
			pkgPath, err := g.genFromFile(path)
			if err != nil {
				return err // A real error happened, stop the walk
			}
			if pkgPath != "" {
				generatedPackages[pkgPath] = struct{}{}
			}
		}
		return nil
	})

	if walkErr != nil {
		return walkErr
	}

	if len(generatedPackages) > 0 {
		packages := make([]string, 0, len(generatedPackages))
		for p := range generatedPackages {
			packages = append(packages, p)
		}
		sort.Strings(packages) // for stable order
		if err := g.generateLogicManifest(packages); err != nil {
			return err
		}
	}

	utils.PrintSuccess("✅ Logic files generated successfully.", nil)
	utils.PrintNotice("💡 Hint: Please add `_ \"{{.ModulePath}}/internal/logic\"` to your main.go to enable automatic service registration.", utils.TplData{"ModulePath": g.ModuleName})
	return nil
}

func (g *LogicGenerator) genFromFile(file string) (string, error) {
	p := &LogicParser{
		fset:   token.NewFileSet(),
		module: g.ModuleName,
	}

	genInfo, err := p.Parse(file)
	if err != nil {
		// Log other errors for debugging, but don't fail the whole process.
		// log.Printf("failed to parse file %s: %v", file, err)
		return "", nil
	}
	if genInfo == nil || len(genInfo.Functions) == 0 {
		// Not a valid service interface with methods, skip.
		return "", nil
	}

	// Logic file goes into internal/logic/<module>/<file>.go
	logicDir := filepath.Join(g.ModuleRoot, g.Dst, "logic", genInfo.Module)
	logicOutputPath := filepath.Join(logicDir, genInfo.FileName)

	pkgPath := path.Join(g.ModuleName, g.Dst, "logic", genInfo.Module)

	// Check if file exists
	if _, err := os.Stat(logicOutputPath); err == nil && !g.Overwrite {
		// File exists and we are in append mode (default), try to append if new methods are found.
		if _, err := g.appendToFile(logicOutputPath, genInfo); err != nil {
			return "", err
		}
		// Whether new methods were appended or not, the service is valid and should be in the manifest.
		return pkgPath, nil
	}

	// If we are here, it means file doesn't exist, OR it exists and we want to overwrite.
	if err := os.MkdirAll(logicDir, os.ModePerm); err != nil {
		return "", merror.Wrapf(err, "failed to create logic directory %s", logicDir)
	}

	if err := generateFile(logicOutputPath, "serviceLogic", TplGenServiceLogic, genInfo); err != nil {
		return "", err
	}
	relPath, _ := filepath.Rel(g.ModuleRoot, logicOutputPath)
	utils.PrintInfo("📄 Generated logic file: {{.Path}}", utils.TplData{"Path": relPath})
	return pkgPath, nil
}

func (g *LogicGenerator) generateLogicManifest(packages []string) error {
	logicFilePath := filepath.Join(g.ModuleRoot, g.Dst, "logic", "logic.go")
	err := generateFile(logicFilePath, "logicManifest", TplGenLogicManifest, map[string]interface{}{
		"Packages": packages,
	})
	if err != nil {
		return merror.Wrap(err, "failed to generate logic manifest file")
	}

	displayPath := logicFilePath
	if relPath, err := filepath.Rel(g.ModuleRoot, logicFilePath); err == nil {
		displayPath = relPath
	}
	utils.PrintInfo("📄 Generated logic manifest file: {{.Path}}", utils.TplData{"Path": displayPath})
	return nil
}

func (g *LogicGenerator) appendToFile(path string, genInfo *logicTplData) (bool, error) {
	existingMethods, err := parseExistingLogicMethods(path)
	if err != nil {
		return false, err // Or maybe just warn and skip? Better to return error.
	}

	var methodsToAppend []logicFunction
	for _, f := range genInfo.Functions {
		if _, ok := existingMethods[f.Name]; !ok {
			methodsToAppend = append(methodsToAppend, f)
		}
	}

	displayPath := path
	if relPath, err := filepath.Rel(g.ModuleRoot, path); err == nil {
		displayPath = relPath
	}

	if len(methodsToAppend) == 0 {
		utils.PrintNotice("⏩ Logic file {{.File}} is up to date, skipping.", utils.TplData{"File": displayPath})
		return false, nil
	}

	// We have methods to append.
	appendData := *genInfo // copy
	appendData.Functions = methodsToAppend

	// Generate the code snippet to append
	var buffer bytes.Buffer
	tpl, err := template.New("serviceLogicAppend").Parse(TplGenServiceLogicAppend)
	if err != nil {
		return false, merror.Wrap(err, "failed to parse append template")
	}
	if err = tpl.Execute(&buffer, appendData); err != nil {
		return false, merror.Wrap(err, "failed to execute append template")
	}

	// Format the generated code before appending.
	formatted, err := format.Source(buffer.Bytes())
	if err != nil {
		utils.PrintWarn("failed to format source for {{.Path}}, writing unformatted code. Error: {{.Error}}", utils.TplData{"Path": path, "Error": err})
		formatted = buffer.Bytes() // Append unformatted code on error
	}

	// Append to file
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return false, merror.Wrap(err, "failed to open logic file for appending")
	}
	defer f.Close()

	if _, err = f.Write(formatted); err != nil {
		return false, merror.Wrap(err, "failed to append new methods to logic file")
	}

	utils.PrintSuccess("Appended {{.Count}} new methods to {{.File}}.", utils.TplData{
		"Count": len(methodsToAppend),
		"File":  displayPath,
	})

	return true, nil
}

func parseExistingLogicMethods(filePath string) (map[string]struct{}, error) {
	methods := make(map[string]struct{})
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, merror.Wrapf(err, "failed to parse existing logic file %s", filePath)
	}

	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || len(fn.Recv.List) == 0 {
			return true
		}
		methods[fn.Name.Name] = struct{}{}
		return true
	})

	return methods, nil
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
	var foundInterface bool

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

		// Check for interface
		if iface, ok := typeSpec.Type.(*ast.InterfaceType); ok {
			if strings.HasPrefix(typeSpec.Name.Name, "I") {
				foundInterface = true
				serviceName = strings.TrimPrefix(typeSpec.Name.Name, "I")

				for _, method := range iface.Methods.List {
					if len(method.Names) == 0 {
						continue
					}
					funcType, ok := method.Type.(*ast.FuncType)
					if !ok {
						continue
					}

					methodName := method.Names[0].Name
					params := funcType.Params
					results := funcType.Results

					// --- Method signature validation ---
					if params == nil || results == nil ||
						params.NumFields() < 1 || params.NumFields() > 2 ||
						results.NumFields() < 1 || results.NumFields() > 2 ||
						!isContextContext(params.List[0]) ||
						!isError(results.List[results.NumFields()-1]) {
						utils.PrintWarn("⚠️ Skipping method '{{.Method}}' in service '{{.Service}}' due to unsupported signature. Supported formats are:\n  - Method(context.Context) error\n  - Method(context.Context) (*Output, error)\n  - Method(context.Context, *Input) error\n  - Method(context.Context, *Input) (*Output, error)",
							utils.TplData{"Method": methodName, "Service": serviceName})
						continue
					}

					// --- Request Parsing (if it exists) ---
					var reqPkg, reqName string
					var reqIsPointer bool
					if params.NumFields() == 2 {
						reqField := params.List[1]
						reqPkg, reqName, reqIsPointer = parseType(reqField.Type)
					}

					// --- Response Parsing (if it exists) ---
					var resPkg, resName string
					var resIsPointer bool
					if results.NumFields() == 2 {
						resField := results.List[0]
						resPkg, resName, resIsPointer = parseType(resField.Type)
					}

					functions = append(functions, logicFunction{
						Name:         methodName,
						ReqName:      reqName,
						ResName:      resName,
						ReqIsPointer: reqIsPointer,
						ResIsPointer: resIsPointer,
					})

					if apiPkg == "" {
						if reqPkg != "" {
							apiPkg = reqPkg
							apiModule = imports[reqPkg]
						} else if resPkg != "" {
							apiPkg = resPkg
							apiModule = imports[resPkg]
						}
					}
				}
				return false // Stop after finding the first interface
			}
		}

		return true
	})

	if !foundInterface {
		utils.PrintWarn("not_have_service_interface", nil)
		return nil, nil // Not a service file we can process.
	}

	if serviceName == "" {
		return nil, nil // Not a service interface file, just skip.
	}

	fileName := filepath.Base(filePath)

	// The module name for logic should be derived from the service file name,
	// but sanitized to be a valid Go package name.
	// E.g., "user_center.go" -> "user_center" -> "usercenter"
	dirtyModuleName := strings.TrimSuffix(fileName, ".go")
	cleanModuleName := sanitizeModuleName(dirtyModuleName)

	svcPackage := strings.ReplaceAll(filepath.Join(p.module, "internal", "service"), "\\", "/")

	info := &logicTplData{
		Module:     cleanModuleName, // Use the sanitized name for package and directory
		Service:    serviceName,
		ApiModule:  apiModule,
		ApiPkg:     apiPkg,
		FileName:   fileName, // Keep the original filename with underscores for the output file
		Functions:  functions,
		SvcPackage: svcPackage,
	}

	return info, nil
}

func isContextContext(field *ast.Field) bool {
	if selExpr, ok := field.Type.(*ast.SelectorExpr); ok {
		if pkg, ok := selExpr.X.(*ast.Ident); ok {
			return pkg.Name == "context" && selExpr.Sel.Name == "Context"
		}
	}
	return false
}

func isError(field *ast.Field) bool {
	if ident, ok := field.Type.(*ast.Ident); ok {
		return ident.Name == "error"
	}
	return false
}

func parseType(expr ast.Expr) (pkg, name string, isPointer bool) {
	if starExpr, ok := expr.(*ast.StarExpr); ok {
		isPointer = true
		expr = starExpr.X
	}

	selector, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return
	}

	if pkgIdent, ok := selector.X.(*ast.Ident); ok {
		pkg = pkgIdent.Name
		name = selector.Sel.Name
	}
	return
}
