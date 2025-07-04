// Package openapi handles the generation of the OpenAPI specification.
package openapi

import (
	"os"
	"path"
	"path/filepath"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"gopkg.in/yaml.v3"
)

// Generate scans the given directory for Go files, parses them,
// builds an OpenAPI specification, and writes it to the output file.
func Generate(src, outputFile string) error {
	utils.PrintInfo("🔍 Scanning directory: {{.Path}}", utils.TplData{"Path": filepath.Base(src)})

	// Step 1: Parse the source code in the directory.
	// The parser will return a structured representation of the API definitions.
	apiDefs, err := ParseDir(src)
	if err != nil {
		return merror.Wrap(err, "failed to parse source directory")
	}

	if len(apiDefs) == 0 {
		return merror.Newf("no API definitions (structs with m.Meta) found in %s", src)
	}

	utils.PrintInfo("ℹ️  Found {{.Count}} API endpoint definitions.", utils.TplData{"Count": len(apiDefs)})

	moduleName, _, err := utils.GetModuleInfo(".")
	if err != nil {
		return merror.Wrap(err, "failed to get project name")
	}
	projectName := path.Base(moduleName)

	// Step 2: Build the OpenAPI specification from the parsed definitions.
	spec, err := BuildSpec(apiDefs, projectName)
	if err != nil {
		return merror.Wrap(err, "failed to build OpenAPI spec")
	}

	// Step 3: Marshal the specification to YAML.
	yamlData, err := yaml.Marshal(spec)
	if err != nil {
		return merror.Wrap(err, "failed to marshal spec to YAML")
	}

	// Step 4: Write the YAML data to the output file.
	if err := os.WriteFile(outputFile, yamlData, 0644); err != nil {
		return merror.Wrapf(err, "failed to write OpenAPI spec to %s", outputFile)
	}

	return nil
}
