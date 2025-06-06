package gen

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/mod/modfile"
)

type TemplateData struct {
	Module     string
	StructName string
}

func GenerateService(path, design string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if info.IsDir() {
		return generateFromDir(path, design)
	}
	return generateFromFile(path, design)
}

func generateFromDir(dirPath, design string) error {
	files, err := filepath.Glob(filepath.Join(dirPath, "*.go"))
	if err != nil {
		return fmt.Errorf("failed to glob .go files: %w", err)
	}

	for _, file := range files {
		if err := generateFromFile(file, design); err != nil {
			fmt.Printf("Failed to generate from file %s: %v\n", file, err)
			// Decide if you want to continue or stop on error
		}
	}
	return nil
}

func generateFromFile(filePath, design string) error {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse Go file: %w", err)
	}

	moduleName, err := getModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	structName := strings.TrimSuffix(filepath.Base(filePath), ".go")
	data := TemplateData{
		Module:     moduleName,
		StructName: strings.Title(structName),
	}

	// For now, we assume one file generates one controller and one service.
	// We can add more complex logic to parse structs and methods later.

	if err := generateFileFromTemplate("internal/app/controller", strings.ToLower(structName)+".go", ControllerTemplate, data); err != nil {
		return err
	}

	var serviceTemplate string
	if design == "interface" {
		serviceTemplate = ServiceInterfaceTemplate
	} else {
		serviceTemplate = ServiceStructTemplate
	}

	return generateFileFromTemplate("internal/app/service", strings.ToLower(structName)+".go", serviceTemplate, data)
}

func generateFileFromTemplate(dir, fileName, tmplContent string, data TemplateData) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	filePath := filepath.Join(dir, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	tmpl, err := template.New("").Parse(tmplContent)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	return tmpl.Execute(file, data)
}

func getModuleName() (string, error) {
	goModBytes, err := os.ReadFile("go.mod")
	if err != nil {
		return "", err
	}
	return modfile.ModulePath(goModBytes), nil
}
